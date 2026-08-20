package curriculumidentity

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"gopkg.in/yaml.v3"
)

const MigrationFileVersion = 1

var stableID = regexp.MustCompile(curriculum.StableIDPattern)

type MigrationFile struct {
	Version    int         `yaml:"version"`
	Migrations []Migration `yaml:"migrations"`
}

// Migration explicitly accounts for an intentional identity rename or
// removal. From and To are slash-qualified IDs; Removed must be true when an
// entity is intentionally retired without a replacement.
type Migration struct {
	Entity  Kind   `yaml:"entity"`
	From    string `yaml:"from"`
	To      string `yaml:"to,omitempty"`
	Removed bool   `yaml:"removed,omitempty"`

	fromParts []string
	toParts   []string
}

func (m Migration) key() string {
	return string(m.Entity) + "\x00" + m.From
}

func (m Migration) Source() Identity {
	return identity(m.Entity, m.fromParts...)
}

func (m Migration) Destination() (Identity, bool) {
	if m.Removed {
		return Identity{}, false
	}
	return identity(m.Entity, m.toParts...), true
}

func ParseMigrations(data []byte) ([]Migration, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var file MigrationFile
	if err := decoder.Decode(&file); err != nil {
		return nil, fmt.Errorf("decode identity migrations: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, errors.New("decode identity migrations: multiple YAML documents are not supported")
		}
		return nil, fmt.Errorf("decode identity migrations: %w", err)
	}
	if file.Version != MigrationFileVersion {
		return nil, fmt.Errorf("identity migration version must be %d", MigrationFileVersion)
	}
	if file.Migrations == nil {
		return nil, errors.New("identity migrations must be an array")
	}

	migrations := append([]Migration(nil), file.Migrations...)
	seen := make(map[string]struct{}, len(migrations))
	for index := range migrations {
		migration := &migrations[index]
		arity, ok := kindArity[migration.Entity]
		if !ok {
			return nil, fmt.Errorf("identity migration %d: unknown entity %q", index, migration.Entity)
		}
		if (migration.To == "") == !migration.Removed {
			return nil, fmt.Errorf("identity migration %d: set exactly one of to or removed: true", index)
		}
		fromParts, err := parseQualifiedID(migration.From, arity)
		if err != nil {
			return nil, fmt.Errorf("identity migration %d from: %w", index, err)
		}
		migration.fromParts = fromParts
		if migration.To != "" {
			toParts, err := parseQualifiedID(migration.To, arity)
			if err != nil {
				return nil, fmt.Errorf("identity migration %d to: %w", index, err)
			}
			migration.toParts = toParts
			if migration.From == migration.To {
				return nil, fmt.Errorf("identity migration %d: from and to must differ", index)
			}
		}
		if _, duplicate := seen[migration.key()]; duplicate {
			return nil, fmt.Errorf("identity migration %d: duplicate %s %s", index, migration.Entity, migration.From)
		}
		seen[migration.key()] = struct{}{}
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].key() < migrations[j].key() })
	return migrations, nil
}

// ValidateAppendOnly ensures the migration ledger remains durable history.
// Existing entries may not be removed or rewritten; new entries may be added.
func ValidateAppendOnly(base, head []Migration) error {
	headByKey := make(map[string]Migration, len(head))
	for _, migration := range head {
		headByKey[migration.key()] = migration
	}

	missing := make([]string, 0)
	changed := make([]string, 0)
	for _, previous := range base {
		current, ok := headByKey[previous.key()]
		if !ok {
			missing = append(missing, fmt.Sprintf("%s %s", previous.Entity, previous.From))
			continue
		}
		if current.To != previous.To || current.Removed != previous.Removed {
			changed = append(changed, fmt.Sprintf("%s %s", previous.Entity, previous.From))
		}
	}
	sort.Strings(missing)
	sort.Strings(changed)
	if len(missing) == 0 && len(changed) == 0 {
		return nil
	}

	parts := make([]string, 0, 2)
	if len(missing) != 0 {
		parts = append(parts, "removed entries: "+strings.Join(missing, ", "))
	}
	if len(changed) != 0 {
		parts = append(parts, "rewritten entries: "+strings.Join(changed, ", "))
	}
	return errors.New("identity migration ledger is append-only: " + strings.Join(parts, "; "))
}

// ValidateNewMigrationsApplied prevents ledger entries from pre-authorizing a
// future identity change. Every entry first introduced by the head revision
// must account for an identity that disappeared in that same comparison.
func ValidateNewMigrationsApplied(base, head, applied []Migration) error {
	baseKeys := make(map[string]struct{}, len(base))
	for _, migration := range base {
		baseKeys[migration.key()] = struct{}{}
	}
	appliedKeys := make(map[string]struct{}, len(applied))
	for _, migration := range applied {
		appliedKeys[migration.key()] = struct{}{}
	}

	unused := make([]string, 0)
	for _, migration := range head {
		if _, existed := baseKeys[migration.key()]; existed {
			continue
		}
		if _, used := appliedKeys[migration.key()]; !used {
			unused = append(unused, fmt.Sprintf("%s %s", migration.Entity, migration.From))
		}
	}
	sort.Strings(unused)
	if len(unused) != 0 {
		return errors.New("new identity migrations must be exercised by the current base-to-head change: " + strings.Join(unused, ", "))
	}
	return nil
}

// ValidateReservedMigrationSources prevents a retired identity from being
// reused while its historical migration still maps that identity elsewhere.
func ValidateReservedMigrationSources(head []Identity, migrations []Migration) error {
	headKeys := make(map[string]struct{}, len(head))
	for _, current := range head {
		headKeys[current.key()] = struct{}{}
	}

	reused := make([]string, 0)
	for _, migration := range migrations {
		source := migration.Source()
		if _, exists := headKeys[source.key()]; exists {
			reused = append(reused, source.String())
		}
	}
	sort.Strings(reused)
	if len(reused) != 0 {
		return errors.New("identity migration sources are reserved and cannot exist in the current curriculum: " + strings.Join(reused, ", "))
	}
	return nil
}

func parseQualifiedID(value string, arity int) ([]string, error) {
	parts := strings.Split(value, "/")
	if len(parts) != arity {
		return nil, fmt.Errorf("expected %d slash-qualified IDs, got %q", arity, value)
	}
	for _, part := range parts {
		if !stableID.MatchString(part) {
			return nil, fmt.Errorf("invalid stable ID %q", part)
		}
	}
	return parts, nil
}
