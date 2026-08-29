package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/johncrowleydev/helix-academy/server/internal/curriculum"
	"github.com/johncrowleydev/helix-academy/server/internal/tutor"
	"github.com/johncrowleydev/helix-academy/server/internal/worksheetpdf"
)

func TestModuleWorkbookEndpoint(t *testing.T) {
	renderer := &fakeWorksheetDocumentRenderer{}
	app := newAPI(tutor.NewService(tutor.NewUnavailableProvider()), testCatalog(t), nil, nil, renderer)

	for _, test := range []struct {
		workbookID string
		variant    worksheetpdf.Variant
		filename   string
	}{
		{workbookID: "student", variant: worksheetpdf.Student, filename: "python-workbook.pdf"},
		{workbookID: "solutions", variant: worksheetpdf.Solutions, filename: "python-workbook-solutions.pdf"},
	} {
		t.Run(test.workbookID, func(t *testing.T) {
			response := serve(t, app.Handler, http.MethodGet, "/api/courses/ai-ml/modules/python/workbooks/"+test.workbookID)
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
			if got := renderer.workbookVariants[len(renderer.workbookVariants)-1]; got != test.variant {
				t.Fatalf("rendered variant = %q, want %q", got, test.variant)
			}
		})
	}
}

func TestModuleWorkbookEndpointMissingResources(t *testing.T) {
	renderer := &fakeWorksheetDocumentRenderer{}
	app := newAPI(tutor.NewService(tutor.NewUnavailableProvider()), testCatalog(t), nil, nil, renderer)
	for _, path := range []string{
		"/api/courses/ai-ml/modules/python/workbooks/answer-key",
		"/api/courses/ai-ml/modules/missing/workbooks/student",
		"/api/courses/ai-ml/modules/foundations/workbooks/student",
		"/api/courses/other/modules/python/workbooks/student",
	} {
		t.Run(path, func(t *testing.T) {
			response := serve(t, app.Handler, http.MethodGet, path)
			if response.Code != http.StatusNotFound {
				t.Fatalf("status = %d: %s", response.Code, response.Body.String())
			}
		})
	}
	if len(renderer.workbookVariants) != 0 {
		t.Fatalf("renderer called for missing workbooks: %v", renderer.workbookVariants)
	}
}

func TestModuleWorkbookEndpointRenderingErrors(t *testing.T) {
	for name, test := range map[string]struct {
		err    error
		status int
	}{
		"unavailable tooling": {err: errors.Join(worksheetpdf.ErrToolUnavailable, errors.New("tectonic not found")), status: http.StatusServiceUnavailable},
		"render failure":      {err: errors.New("private renderer detail"), status: http.StatusInternalServerError},
	} {
		t.Run(name, func(t *testing.T) {
			renderer := &fakeWorksheetDocumentRenderer{err: test.err}
			app := newAPI(tutor.NewService(tutor.NewUnavailableProvider()), testCatalog(t), nil, nil, renderer)
			response := serve(t, app.Handler, http.MethodGet, "/api/courses/ai-ml/modules/python/workbooks/student")
			if response.Code != test.status {
				t.Fatalf("status = %d: %s", response.Code, response.Body.String())
			}
			if strings.Contains(response.Body.String(), "private renderer detail") {
				t.Fatalf("renderer detail leaked: %s", response.Body.String())
			}
		})
	}
}

func TestModuleWorkbookOpenAPIContract(t *testing.T) {
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), curriculum.NewEmptyCatalog(), nil, nil)
	path := app.Spec.Paths["/api/courses/{courseId}/modules/{moduleId}/workbooks/{workbookId}"]
	if path == nil || path.Get == nil || path.Get.OperationID != "getCourseModuleWorkbook" {
		t.Fatalf("missing module workbook operation: %#v", path)
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
