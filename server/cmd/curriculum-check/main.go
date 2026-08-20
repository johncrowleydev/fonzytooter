package main

import (
	"fmt"
	"io"
	"os"

	"github.com/johncrowleydev/fonzytooter/server/internal/config"
	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
)

func main() {
	cfg := config.FromEnv()
	pathName := cfg.CurriculumPath
	if len(os.Args) > 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/curriculum-check [curriculum-path]")
		os.Exit(2)
	}
	if len(os.Args) == 2 {
		pathName = os.Args[1]
	}
	if info, err := os.Stat(pathName); err != nil {
		fmt.Fprintf(os.Stderr, "open curriculum path %q: %v\n", pathName, err)
		os.Exit(1)
	} else if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "open curriculum path %q: not a directory\n", pathName)
		os.Exit(1)
	}

	catalog, err := curriculum.Load(os.DirFS(pathName))
	if err != nil {
		fmt.Fprintf(os.Stderr, "curriculum invalid: %v\n", err)
		os.Exit(1)
	}
	writeUnusedSourceWarnings(os.Stderr, catalog)

	fmt.Printf("curriculum valid: %d courses, %d modules, %d lessons, %d objectives, %d sources\n", catalog.CourseCount(), catalog.ModuleCount(), catalog.LessonCount(), catalog.ObjectiveCount(), catalog.SourceCount())
}

func writeUnusedSourceWarnings(output io.Writer, catalog *curriculum.Catalog) {
	for _, sourceID := range catalog.UnusedSourceIDs() {
		fmt.Fprintf(output, "warning: unused source id %q\n", sourceID)
	}
}
