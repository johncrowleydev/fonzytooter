package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/johncrowleydev/fonzytooter/server/internal/config"
	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/curriculumidentity"
	"github.com/johncrowleydev/fonzytooter/server/internal/curriculumstate"
)

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || (args[0] != "audit" && args[0] != "migrate") {
		writeUsage(stderr)
		return 2
	}
	mode := args[0]
	cfg := config.FromEnv()
	flags := flag.NewFlagSet("curriculum-state-check "+mode, flag.ContinueOnError)
	flags.SetOutput(stderr)
	databasePath := flags.String("database", cfg.DatabasePath, "path to an existing Fonzytooter SQLite database")
	curriculumPath := flags.String("curriculum", cfg.CurriculumPath, "path to the current curriculum")
	migrationsPath := flags.String("migrations", cfg.CurriculumPath+"/identity-migrations.yaml", "path to the identity migration ledger")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		writeUsage(stderr)
		return 2
	}

	catalog, err := curriculum.Load(os.DirFS(*curriculumPath))
	if err != nil {
		fmt.Fprintf(stderr, "load current curriculum: %v\n", err)
		return 1
	}
	if mode == "audit" {
		db, err := curriculumstate.OpenReadOnly(ctx, *databasePath)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		defer db.Close()
		return writeAudit(ctx, stdout, stderr, db, catalog)
	}

	data, err := os.ReadFile(*migrationsPath)
	if err != nil {
		fmt.Fprintf(stderr, "read identity migrations: %v\n", err)
		return 1
	}
	migrations, err := curriculumidentity.ParseMigrations(data)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	db, err := curriculumstate.OpenReadWrite(ctx, *databasePath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer db.Close()
	updates, err := curriculumstate.ApplyMigrations(ctx, db, catalog, migrations)
	if err != nil {
		fmt.Fprintf(stderr, "apply curriculum identity migrations: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, "curriculum identity migration complete (no rows deleted)")
	for _, update := range updates {
		fmt.Fprintf(stdout, "- %s %s -> %s: %d identity values updated\n", update.Entity, update.From, update.To, update.Updates)
	}
	removed := 0
	for _, migration := range migrations {
		if migration.Removed {
			removed++
		}
	}
	if removed != 0 {
		fmt.Fprintf(stdout, "- %d explicit removals preserved for audit; no historical rows were deleted\n", removed)
	}
	fmt.Fprintln(stdout)
	return writeAudit(ctx, stdout, stderr, db, catalog)
}

func writeAudit(ctx context.Context, stdout, stderr io.Writer, db *sql.DB, catalog *curriculum.Catalog) int {
	report, err := curriculumstate.Audit(ctx, db, catalog)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := curriculumstate.WriteReport(stdout, report); err != nil {
		fmt.Fprintf(stderr, "write curriculum state report: %v\n", err)
		return 1
	}
	if !report.Clean() {
		return 1
	}
	return 0
}

func writeUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: curriculum-state-check <audit|migrate> [--database <path>] [--curriculum <path>] [--migrations <path>]")
}
