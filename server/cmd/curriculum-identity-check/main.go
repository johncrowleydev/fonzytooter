package main

import (
	"archive/tar"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing/fstest"

	"github.com/johncrowleydev/helix-academy/server/internal/curriculum"
	"github.com/johncrowleydev/helix-academy/server/internal/curriculumidentity"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("curriculum-identity-check", flag.ContinueOnError)
	flags.SetOutput(stderr)
	repository := flags.String("repository", "..", "path to the Git repository")
	baseRef := flags.String("base-ref", "", "Git revision containing the base curriculum")
	curriculumPath := flags.String("curriculum", "../curriculum", "path to the head curriculum")
	migrationsPath := flags.String("migrations", "../curriculum/identity-migrations.yaml", "path to the identity migration ledger")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 || *baseRef == "" {
		fmt.Fprintln(stderr, "usage: curriculum-identity-check --base-ref <git-revision> [--repository <path>] [--curriculum <path>] [--migrations <path>]")
		return 2
	}

	baseFS, err := gitCurriculumFS(*repository, *baseRef)
	if err != nil {
		fmt.Fprintf(stderr, "load base curriculum: %v\n", err)
		return 1
	}
	baseCatalog, err := curriculum.Load(baseFS)
	if err != nil {
		fmt.Fprintf(stderr, "base curriculum invalid: %v\n", err)
		return 1
	}
	headCatalog, err := curriculum.Load(os.DirFS(*curriculumPath))
	if err != nil {
		fmt.Fprintf(stderr, "head curriculum invalid: %v\n", err)
		return 1
	}
	baseMigrations, err := readBaseMigrations(baseFS)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	migrationData, err := os.ReadFile(*migrationsPath)
	if err != nil {
		fmt.Fprintf(stderr, "read identity migrations: %v\n", err)
		return 1
	}
	migrations, err := curriculumidentity.ParseMigrations(migrationData)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := curriculumidentity.ValidateAppendOnly(baseMigrations, migrations); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	base := curriculumidentity.Snapshot(baseCatalog)
	head := curriculumidentity.Snapshot(headCatalog)
	result := curriculumidentity.Compare(base, head, migrations)
	if len(result.BreakingChanges) != 0 {
		fmt.Fprintln(stderr, "breaking curriculum identity changes:")
		for _, change := range result.BreakingChanges {
			fmt.Fprintf(stderr, "- %s: %s\n", change.Identity, change.Reason)
		}
		fmt.Fprintln(stderr, "add an explicit rename or removal to curriculum/identity-migrations.yaml")
		return 1
	}
	if err := curriculumidentity.ValidateNewMigrationsApplied(baseMigrations, migrations, result.AppliedMigrations); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := curriculumidentity.ValidateReservedMigrationSources(head, migrations); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	fmt.Fprintf(stdout, "curriculum identity compatible: %d base identities, %d additions, %d applied migrations\n", len(base), len(result.Additions), len(result.AppliedMigrations))
	return 0
}

func readBaseMigrations(baseFS fs.FS) ([]curriculumidentity.Migration, error) {
	data, err := fs.ReadFile(baseFS, "identity-migrations.yaml")
	if errors.Is(err, fs.ErrNotExist) {
		return []curriculumidentity.Migration{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read base identity migrations: %w", err)
	}
	migrations, err := curriculumidentity.ParseMigrations(data)
	if err != nil {
		return nil, fmt.Errorf("parse base identity migrations: %w", err)
	}
	return migrations, nil
}

func gitCurriculumFS(repository, revision string) (fs.FS, error) {
	command := exec.Command("git", "-C", repository, "archive", "--format=tar", "--", revision+":curriculum")
	archive, err := command.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git archive %q: %s", revision, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("git archive %q: %w", revision, err)
	}
	return tarFS(archive)
}

func tarFS(archive []byte) (fs.FS, error) {
	result := fstest.MapFS{}
	reader := tar.NewReader(bytes.NewReader(archive))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read Git archive: %w", err)
		}
		name := path.Clean(strings.TrimPrefix(header.Name, "./"))
		if !fs.ValidPath(name) || name == "." {
			return nil, fmt.Errorf("Git archive contains invalid path %q", header.Name)
		}
		switch header.Typeflag {
		case tar.TypeDir:
			result[name] = &fstest.MapFile{Mode: fs.ModeDir}
		case tar.TypeReg, tar.TypeRegA:
			data, err := io.ReadAll(reader)
			if err != nil {
				return nil, fmt.Errorf("read %s from Git archive: %w", name, err)
			}
			result[name] = &fstest.MapFile{Data: data, Mode: 0o444}
		default:
			return nil, fmt.Errorf("Git archive contains unsupported entry %q", header.Name)
		}
	}
	return result, nil
}
