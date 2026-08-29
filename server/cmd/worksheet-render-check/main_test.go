package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/johncrowleydev/helix-academy/server/internal/curriculum"
	"github.com/johncrowleydev/helix-academy/server/internal/worksheetpdf"
)

type smokeRenderer struct {
	failSolutions bool
}

func (renderer smokeRenderer) Render(_ context.Context, _ curriculum.Worksheet, variant worksheetpdf.Variant) ([]byte, error) {
	if renderer.failSolutions && variant == worksheetpdf.Solutions {
		return nil, errors.New("deliberate solution failure")
	}
	return []byte("%PDF-1.7\ntest"), nil
}

func (renderer smokeRenderer) RenderWorkbook(_ context.Context, _ curriculum.Course, _ curriculum.Module, _ worksheetpdf.Variant) ([]byte, error) {
	return []byte("%PDF-1.7\ntest workbook"), nil
}

func TestRenderCatalogCoversRealWorksheetAndWorkbookVariants(t *testing.T) {
	catalog := loadRealCatalog(t)
	counts, err := renderCatalog(context.Background(), catalog, smokeRenderer{})
	if err != nil {
		t.Fatalf("render catalog: %v", err)
	}
	if counts.worksheetDocuments != catalog.WorksheetCount()*2 {
		t.Fatalf("worksheet documents = %d, want %d", counts.worksheetDocuments, catalog.WorksheetCount()*2)
	}
	if counts.workbookDocuments == 0 || counts.workbookDocuments%2 != 0 {
		t.Fatalf("unexpected workbook document count: %d", counts.workbookDocuments)
	}
}

func TestRenderCatalogReportsCourseModuleWorksheetIdentity(t *testing.T) {
	catalog := loadRealCatalog(t)
	_, err := renderCatalog(context.Background(), catalog, smokeRenderer{failSolutions: true})
	if err == nil {
		t.Fatal("expected render failure")
	}
	for _, expected := range []string{"course \"ai-ml\"", "module \"scientific-python\"", "worksheet \"python-execution-model-practice\"", "solutions document"} {
		if !strings.Contains(err.Error(), expected) {
			t.Fatalf("error omitted %q: %v", expected, err)
		}
	}
}

func loadRealCatalog(t *testing.T) *curriculum.Catalog {
	t.Helper()
	catalog, err := curriculum.Load(os.DirFS("../../../curriculum"))
	if err != nil {
		t.Fatalf("load real curriculum: %v", err)
	}
	return catalog
}
