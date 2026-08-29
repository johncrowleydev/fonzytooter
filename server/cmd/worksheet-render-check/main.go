package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/johncrowleydev/helix-academy/server/internal/curriculum"
	"github.com/johncrowleydev/helix-academy/server/internal/worksheetpdf"
)

type documentRenderer interface {
	Render(context.Context, curriculum.Worksheet, worksheetpdf.Variant) ([]byte, error)
	RenderWorkbook(context.Context, curriculum.Course, curriculum.Module, worksheetpdf.Variant) ([]byte, error)
}

type renderCounts struct {
	worksheetDocuments int
	workbookDocuments  int
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/worksheet-render-check <curriculum-path>")
		os.Exit(2)
	}
	curriculumPath := os.Args[1]
	info, err := os.Stat(curriculumPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open curriculum path %q: %v\n", curriculumPath, err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "open curriculum path %q: not a directory\n", curriculumPath)
		os.Exit(1)
	}
	catalog, err := curriculum.Load(os.DirFS(curriculumPath))
	if err != nil {
		fmt.Fprintf(os.Stderr, "load curriculum: %v\n", err)
		os.Exit(1)
	}
	counts, err := renderCatalog(context.Background(), catalog, worksheetpdf.NewRenderer())
	if err != nil {
		fmt.Fprintf(os.Stderr, "worksheet render check failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("worksheet render check passed: %d worksheet documents, %d workbook documents\n", counts.worksheetDocuments, counts.workbookDocuments)
}

func renderCatalog(ctx context.Context, catalog *curriculum.Catalog, renderer documentRenderer) (renderCounts, error) {
	counts := renderCounts{}
	for _, course := range catalog.Courses() {
		for _, module := range course.Modules {
			for _, worksheet := range module.Worksheets {
				for _, variant := range []worksheetpdf.Variant{worksheetpdf.Student, worksheetpdf.Solutions} {
					pdf, err := renderer.Render(ctx, worksheet, variant)
					if err != nil {
						return counts, fmt.Errorf("render course %q module %q worksheet %q %s document: %w", course.ID, module.ID, worksheet.ID, variant, err)
					}
					if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
						return counts, fmt.Errorf("render course %q module %q worksheet %q %s document: output is not a PDF", course.ID, module.ID, worksheet.ID, variant)
					}
					counts.worksheetDocuments++
				}
			}
			if len(module.Worksheets) == 0 {
				continue
			}
			for _, variant := range []worksheetpdf.Variant{worksheetpdf.Student, worksheetpdf.Solutions} {
				pdf, err := renderer.RenderWorkbook(ctx, course, module, variant)
				if err != nil {
					return counts, fmt.Errorf("render course %q module %q %s workbook: %w", course.ID, module.ID, variant, err)
				}
				if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
					return counts, fmt.Errorf("render course %q module %q %s workbook: output is not a PDF", course.ID, module.ID, variant)
				}
				counts.workbookDocuments++
			}
		}
	}
	return counts, nil
}
