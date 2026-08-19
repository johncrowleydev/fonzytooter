package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
)

func TestCurriculumReadAPI(t *testing.T) {
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), testCatalog(t))

	listResponse := serve(t, app.Handler, http.MethodGet, "/api/modules")
	if listResponse.Code != http.StatusOK {
		t.Fatalf("list modules status = %d: %s", listResponse.Code, listResponse.Body.String())
	}
	if got := strings.TrimSpace(listResponse.Body.String()); !strings.HasPrefix(got, "[") {
		t.Fatalf("expected pure JSON array, got %s", got)
	}
	var summaries []ModuleSummary
	decodeJSON(t, listResponse, &summaries)
	if len(summaries) != 1 || summaries[0].ID != "python" || summaries[0].Order != 2 {
		t.Fatalf("unexpected module summaries: %#v", summaries)
	}

	moduleResponse := serve(t, app.Handler, http.MethodGet, "/api/modules/python")
	if moduleResponse.Code != http.StatusOK {
		t.Fatalf("get module status = %d: %s", moduleResponse.Code, moduleResponse.Body.String())
	}
	var module ModuleResource
	decodeJSON(t, moduleResponse, &module)
	if module.ID != "python" || len(module.Objectives) != 1 || len(module.Videos) != 1 || len(module.Lessons) != 1 {
		t.Fatalf("unexpected module resource: %#v", module)
	}
	if module.Lessons[0].ID != "lesson.stable" || len(module.Lessons[0].ObjectiveIDs) != 1 {
		t.Fatalf("unexpected lesson summary: %#v", module.Lessons[0])
	}

	lessonResponse := serve(t, app.Handler, http.MethodGet, "/api/modules/python/lessons/lesson.stable")
	if lessonResponse.Code != http.StatusOK {
		t.Fatalf("get lesson status = %d: %s", lessonResponse.Code, lessonResponse.Body.String())
	}
	var lesson LessonResource
	decodeJSON(t, lessonResponse, &lesson)
	if lesson.ModuleID != "python" || lesson.Content != "# Lesson\n" {
		t.Fatalf("unexpected lesson resource: %#v", lesson)
	}
	if strings.Contains(lesson.Content, "id: lesson.stable") || len(lesson.Sources) != 1 || lesson.Sources[0].ID != "go-docs" {
		t.Fatalf("lesson did not expose body/source resource correctly: %#v", lesson)
	}
}

func TestCurriculumReadAPIMissingResourcesUseProblemResponses(t *testing.T) {
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), testCatalog(t))
	for _, path := range []string{"/api/modules/missing", "/api/modules/python/lessons/missing"} {
		t.Run(path, func(t *testing.T) {
			response := serve(t, app.Handler, http.MethodGet, path)
			if response.Code != http.StatusNotFound {
				t.Fatalf("status = %d: %s", response.Code, response.Body.String())
			}
			if got := response.Header().Get("Content-Type"); got != "application/problem+json" {
				t.Fatalf("content type = %q", got)
			}
			var problem map[string]any
			decodeJSON(t, response, &problem)
			if _, ok := problem["error"]; ok {
				t.Fatalf("unexpected endpoint-specific error envelope: %#v", problem)
			}
			if problem["status"] != float64(http.StatusNotFound) {
				t.Fatalf("problem status = %#v", problem["status"])
			}
		})
	}
}

func TestCurriculumOpenAPIContract(t *testing.T) {
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), curriculum.NewEmptyCatalog())
	for path, operationID := range map[string]string{
		"/api/modules":                               "listModules",
		"/api/modules/{moduleId}":                    "getModule",
		"/api/modules/{moduleId}/lessons/{lessonId}": "getLesson",
	} {
		item, ok := app.Spec.Paths[path]
		if !ok || item.Get == nil || item.Get.OperationID != operationID {
			t.Fatalf("missing curriculum operation %s %s", path, operationID)
		}
	}
	listSchema := app.Spec.Components.Schemas.Map()["ModuleSummaryList"]
	if listSchema == nil || listSchema.Type != "array" || listSchema.Items == nil {
		t.Fatalf("expected reusable non-null module list schema, got %#v", listSchema)
	}
	if schema := app.Spec.Paths["/api/modules"].Get.Responses["200"].Content["application/json"].Schema; schema.Ref != "#/components/schemas/ModuleSummaryList" {
		t.Fatalf("expected list response to reference reusable array schema, got %#v", schema)
	}
	contentSchema := app.Spec.Components.Schemas.Map()["LessonResource"].Properties["content"]
	if contentSchema == nil || !strings.Contains(contentSchema.Description, "frontmatter removed") {
		t.Fatalf("expected lesson content documentation, got %#v", contentSchema)
	}
}

func testCatalog(t *testing.T) *curriculum.Catalog {
	t.Helper()
	catalog, err := curriculum.Load(fstest.MapFS{
		"sources.yaml":                                 &fstest.MapFile{Data: []byte("sources:\n  go-docs:\n    title: Go documentation\n    url: https://go.dev/doc/\n")},
		"courses/ai-ml/course.yaml":                    &fstest.MapFile{Data: []byte("id: ai-ml\ntitle: AI & Machine Learning\ndescription: Learn AI and machine learning.\norder: 0\n")},
		"courses/ai-ml/modules/storage/module.yaml":    &fstest.MapFile{Data: []byte("id: python\ntitle: Python\norder: 2\nobjectives:\n  - id: python.variables\n    title: Use variables\n    description: Bind names to values.\n    prerequisites: []\nvideos:\n  - id: python-video\n    title: Python video\n    url: https://example.com/python\n    objectiveIds:\n      - python.variables\nlessons:\n  - lesson.stable\n")},
		"courses/ai-ml/modules/storage/not-the-id.mdx": &fstest.MapFile{Data: []byte("---\nid: lesson.stable\ntitle: Stable lesson\nobjectiveIds:\n  - python.variables\nsourceIds:\n  - go-docs\n---\n# Lesson\n")},
	})
	if err != nil {
		t.Fatalf("load test catalog: %v", err)
	}
	return catalog
}

func serve(t *testing.T, handler http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(method, path, nil))
	return response
}

func decodeJSON(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode JSON: %v; body=%s", err, response.Body.String())
	}
}
