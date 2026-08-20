package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/curriculum-schema <output-directory>")
		os.Exit(2)
	}

	outputDirectory := os.Args[1]
	if err := os.MkdirAll(outputDirectory, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "create output directory %q: %v\n", outputDirectory, err)
		os.Exit(1)
	}

	schemas, err := curriculum.GenerateEditorSchemas()
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate curriculum schemas: %v\n", err)
		os.Exit(1)
	}
	for _, schema := range schemas {
		outputPath := filepath.Join(outputDirectory, schema.Filename)
		if err := os.WriteFile(outputPath, schema.Data, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "write curriculum schema %q: %v\n", outputPath, err)
			os.Exit(1)
		}
	}
}
