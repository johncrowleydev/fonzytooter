package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
	"github.com/johncrowleydev/fonzytooter/server/internal/worksheetpdf"
)

type fakeWorksheetDocumentRenderer struct {
	err              error
	variants         []worksheetpdf.Variant
	workbookVariants []worksheetpdf.Variant
}

func (renderer *fakeWorksheetDocumentRenderer) RenderWorkbook(_ context.Context, _ curriculum.Course, _ curriculum.Module, variant worksheetpdf.Variant) ([]byte, error) {
	renderer.workbookVariants = append(renderer.workbookVariants, variant)
	if renderer.err != nil {
		return nil, renderer.err
	}
	return []byte("%PDF-1.7\ntest workbook"), nil
}

func (renderer *fakeWorksheetDocumentRenderer) Render(_ context.Context, _ curriculum.Worksheet, variant worksheetpdf.Variant) ([]byte, error) {
	renderer.variants = append(renderer.variants, variant)
	if renderer.err != nil {
		return nil, renderer.err
	}
	return []byte("%PDF-1.7\ntest"), nil
}

func TestWorksheetDocumentEndpoint(t *testing.T) {
	renderer := &fakeWorksheetDocumentRenderer{}
	app := newAPI(tutor.NewService(tutor.NewUnavailableProvider()), testCatalog(t), nil, nil, renderer)

	for _, test := range []struct {
		documentID string
		variant    worksheetpdf.Variant
		filename   string
	}{
		{documentID: "student", variant: worksheetpdf.Student, filename: "worksheet.pdf"},
		{documentID: "solutions", variant: worksheetpdf.Solutions, filename: "worksheet-solutions.pdf"},
	} {
		t.Run(test.documentID, func(t *testing.T) {
			path := "/api/courses/ai-ml/modules/python/worksheets/worksheet/documents/" + test.documentID
			response := serve(t, app.Handler, http.MethodGet, path)
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d: %s", response.Code, response.Body.String())
			}
			if got := response.Header().Get("Content-Type"); got != "application/pdf" {
				t.Fatalf("content type = %q", got)
			}
			wantDisposition := "attachment; filename=\"" + test.filename + "\""
			if got := response.Header().Get("Content-Disposition"); got != wantDisposition {
				t.Fatalf("content disposition = %q, want %q", got, wantDisposition)
			}
			if !strings.HasPrefix(response.Body.String(), "%PDF-") {
				t.Fatalf("body is not a PDF: %q", response.Body.String())
			}
			if got := renderer.variants[len(renderer.variants)-1]; got != test.variant {
				t.Fatalf("rendered variant = %q, want %q", got, test.variant)
			}
		})
	}
}

func TestWorksheetDocumentEndpointMissingResources(t *testing.T) {
	renderer := &fakeWorksheetDocumentRenderer{}
	app := newAPI(tutor.NewService(tutor.NewUnavailableProvider()), testCatalog(t), nil, nil, renderer)
	for _, path := range []string{
		"/api/courses/ai-ml/modules/python/worksheets/worksheet/documents/answer-key",
		"/api/courses/ai-ml/modules/python/worksheets/missing/documents/student",
		"/api/courses/other/modules/python/worksheets/worksheet/documents/student",
	} {
		t.Run(path, func(t *testing.T) {
			response := serve(t, app.Handler, http.MethodGet, path)
			if response.Code != http.StatusNotFound {
				t.Fatalf("status = %d: %s", response.Code, response.Body.String())
			}
		})
	}
	if len(renderer.variants) != 0 {
		t.Fatalf("renderer called for missing resources: %v", renderer.variants)
	}
}

func TestWorksheetDocumentEndpointUnavailableTooling(t *testing.T) {
	renderer := &fakeWorksheetDocumentRenderer{err: errors.Join(worksheetpdf.ErrToolUnavailable, errors.New("pandoc not found"))}
	app := newAPI(tutor.NewService(tutor.NewUnavailableProvider()), testCatalog(t), nil, nil, renderer)
	response := serve(t, app.Handler, http.MethodGet, "/api/courses/ai-ml/modules/python/worksheets/worksheet/documents/student")
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("content type = %q", got)
	}
}

func TestWorksheetDocumentOpenAPIContract(t *testing.T) {
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), curriculum.NewEmptyCatalog(), nil, nil)
	path := app.Spec.Paths["/api/courses/{courseId}/modules/{moduleId}/worksheets/{worksheetId}/documents/{documentId}"]
	if path == nil || path.Get == nil || path.Get.OperationID != "getCourseModuleWorksheetDocument" {
		t.Fatalf("missing worksheet document operation: %#v", path)
	}
	response := path.Get.Responses["200"]
	if response == nil || response.Content["application/pdf"] == nil {
		t.Fatalf("missing PDF success response: %#v", response)
	}
	schema := response.Content["application/pdf"].Schema
	if schema == nil || schema.Type != "string" || schema.Format != "binary" {
		t.Fatalf("PDF schema = %#v", schema)
	}
	if response.Headers["Content-Disposition"] == nil {
		t.Fatalf("missing content-disposition response header: %#v", response.Headers)
	}
	for _, status := range []string{"404", "500", "503"} {
		if path.Get.Responses[status] == nil {
			t.Fatalf("missing documented %s response", status)
		}
	}
}
