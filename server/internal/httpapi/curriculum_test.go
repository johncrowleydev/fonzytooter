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
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), testCatalog(t), nil)

	listResponse := serve(t, app.Handler, http.MethodGet, "/api/courses")
	if listResponse.Code != http.StatusOK {
		t.Fatalf("list courses status = %d: %s", listResponse.Code, listResponse.Body.String())
	}
	if got := strings.TrimSpace(listResponse.Body.String()); !strings.HasPrefix(got, "[") {
		t.Fatalf("expected pure JSON array, got %s", got)
	}
	var summaries []CourseSummary
	decodeJSON(t, listResponse, &summaries)
	if len(summaries) != 2 || summaries[0].ID != "ai-ml" || summaries[0].Title != "AI & Machine Learning" || summaries[0].Description != "Learn AI and machine learning." || summaries[0].Order != 0 {
		t.Fatalf("unexpected course summaries: %#v", summaries)
	}

	courseResponse := serve(t, app.Handler, http.MethodGet, "/api/courses/ai-ml")
	if courseResponse.Code != http.StatusOK {
		t.Fatalf("get course status = %d: %s", courseResponse.Code, courseResponse.Body.String())
	}
	var course CourseResource
	decodeJSON(t, courseResponse, &course)
	if course.ID != "ai-ml" || len(course.Modules) != 2 || course.Modules[0].ID != "foundations" || course.Modules[1].ID != "python" {
		t.Fatalf("unexpected course resource: %#v", course)
	}

	moduleResponse := serve(t, app.Handler, http.MethodGet, "/api/courses/ai-ml/modules/python")
	if moduleResponse.Code != http.StatusOK {
		t.Fatalf("get module status = %d: %s", moduleResponse.Code, moduleResponse.Body.String())
	}
	var module ModuleResource
	decodeJSON(t, moduleResponse, &module)
	if module.CourseID != "ai-ml" || module.ID != "python" || len(module.Objectives) != 1 || len(module.Videos) != 1 || len(module.Lessons) != 1 || len(module.Worksheets) != 1 {
		t.Fatalf("unexpected module resource: %#v", module)
	}
	if module.Lessons[0].ID != "lesson.stable" || len(module.Lessons[0].ObjectiveIDs) != 1 {
		t.Fatalf("unexpected lesson summary: %#v", module.Lessons[0])
	}
	if module.Worksheets[0].ID != "worksheet" || module.Worksheets[0].ProblemCount != 1 {
		t.Fatalf("unexpected worksheet summary: %#v", module.Worksheets[0])
	}

	lessonResponse := serve(t, app.Handler, http.MethodGet, "/api/courses/ai-ml/modules/python/lessons/lesson.stable")
	if lessonResponse.Code != http.StatusOK {
		t.Fatalf("get lesson status = %d: %s", lessonResponse.Code, lessonResponse.Body.String())
	}
	var lesson LessonResource
	decodeJSON(t, lessonResponse, &lesson)
	if lesson.CourseID != "ai-ml" || lesson.ModuleID != "python" || lesson.Content != "# Lesson\n" {
		t.Fatalf("unexpected lesson resource: %#v", lesson)
	}
	if strings.Contains(lesson.Content, "id: lesson.stable") || len(lesson.Sources) != 1 || lesson.Sources[0].ID != "go-docs" {
		t.Fatalf("lesson did not expose body/source resource correctly: %#v", lesson)
	}
	if len(lesson.Worksheets) != 1 || lesson.Worksheets[0].ID != "worksheet" {
		t.Fatalf("lesson did not expose worksheet summaries: %#v", lesson.Worksheets)
	}

	worksheetResponse := serve(t, app.Handler, http.MethodGet, "/api/courses/ai-ml/modules/python/worksheets/worksheet")
	if worksheetResponse.Code != http.StatusOK {
		t.Fatalf("get worksheet status = %d: %s", worksheetResponse.Code, worksheetResponse.Body.String())
	}
	worksheetJSON := worksheetResponse.Body.String()
	if strings.Contains(worksheetJSON, "expectedAnswer") || strings.Contains(worksheetJSON, "rubric") || strings.Contains(worksheetJSON, "Secret answer") || strings.Contains(worksheetJSON, "Secret criterion") {
		t.Fatalf("student worksheet leaked solution data: %s", worksheetJSON)
	}
	var worksheet WorksheetResource
	if err := json.Unmarshal([]byte(worksheetJSON), &worksheet); err != nil {
		t.Fatalf("decode worksheet JSON: %v", err)
	}
	if worksheet.CourseID != "ai-ml" || worksheet.ModuleID != "python" || worksheet.ID != "worksheet" || len(worksheet.Problems) != 1 || !worksheet.Problems[0].RequiresWork || worksheet.Problems[0].ResponseLines != 3 {
		t.Fatalf("unexpected worksheet resource: %#v", worksheet)
	}
}

func TestCurriculumReadAPIMissingResourcesUseProblemResponses(t *testing.T) {
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), testCatalog(t), nil)
	for _, path := range []string{
		"/api/courses/missing",
		"/api/courses/other/modules/python",
		"/api/courses/other/modules/python/lessons/lesson.stable",
		"/api/courses/ai-ml/modules/wrong-module/lessons/lesson.stable",
		"/api/courses/ai-ml/modules/python/lessons/missing",
		"/api/courses/other/modules/python/worksheets/worksheet",
		"/api/courses/ai-ml/modules/wrong-module/worksheets/worksheet",
		"/api/courses/ai-ml/modules/python/worksheets/missing",
	} {
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
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), curriculum.NewEmptyCatalog(), nil)
	for path, operationID := range map[string]string{
		"/api/courses":                               "listCourses",
		"/api/courses/{courseId}":                    "getCourse",
		"/api/courses/{courseId}/modules/{moduleId}": "getCourseModule",
		"/api/courses/{courseId}/modules/{moduleId}/lessons/{lessonId}":       "getCourseLesson",
		"/api/courses/{courseId}/modules/{moduleId}/worksheets/{worksheetId}": "getCourseModuleWorksheet",
	} {
		item, ok := app.Spec.Paths[path]
		if !ok || item.Get == nil || item.Get.OperationID != operationID {
			t.Fatalf("missing curriculum operation %s %s", path, operationID)
		}
	}
	for _, obsoletePath := range []string{
		"/api/modules",
		"/api/modules/{moduleId}",
		"/api/modules/{moduleId}/lessons/{lessonId}",
	} {
		if _, ok := app.Spec.Paths[obsoletePath]; ok {
			t.Fatalf("obsolete curriculum operation remains at %s", obsoletePath)
		}
	}
	listSchema := app.Spec.Components.Schemas.Map()["CourseSummaryList"]
	if listSchema == nil || listSchema.Type != "array" || listSchema.Items == nil || listSchema.Nullable {
		t.Fatalf("expected reusable non-null course list schema, got %#v", listSchema)
	}
	if summarySchema := app.Spec.Components.Schemas.Map()["CourseSummary"]; summarySchema == nil || summarySchema.AdditionalProperties != false {
		t.Fatalf("expected strict course summary schema, got %#v", summarySchema)
	}
	if schema := app.Spec.Paths["/api/courses"].Get.Responses["200"].Content["application/json"].Schema; schema.Ref != "#/components/schemas/CourseSummaryList" {
		t.Fatalf("expected list response to reference reusable array schema, got %#v", schema)
	}
	pageContextSchema := app.Spec.Components.Schemas.Map()["PageContext"]
	if pageContextSchema == nil || pageContextSchema.Properties["courseId"] == nil || pageContextSchema.Properties["courseTitle"] == nil {
		t.Fatalf("expected PageContext course identity, got %#v", pageContextSchema)
	}
	contentSchema := app.Spec.Components.Schemas.Map()["LessonResource"].Properties["content"]
	if contentSchema == nil || !strings.Contains(contentSchema.Description, "frontmatter removed") {
		t.Fatalf("expected lesson content documentation, got %#v", contentSchema)
	}
	worksheetSchema := app.Spec.Components.Schemas.Map()["WorksheetResource"]
	if worksheetSchema == nil || worksheetSchema.Properties["problems"] == nil {
		t.Fatalf("expected worksheet resource schema, got %#v", worksheetSchema)
	}
	problemSchema := app.Spec.Components.Schemas.Map()["WorksheetProblemResource"]
	if problemSchema == nil || problemSchema.Properties["expectedAnswer"] != nil || problemSchema.Properties["rubric"] != nil {
		t.Fatalf("student worksheet problem schema leaked solution fields: %#v", problemSchema)
	}
}

func testCatalog(t *testing.T) *curriculum.Catalog {
	t.Helper()
	catalog, err := curriculum.Load(fstest.MapFS{
		"sources.yaml":                                            &fstest.MapFile{Data: []byte("sources:\n  go-docs:\n    title: Go documentation\n    url: https://go.dev/doc/\n")},
		"courses/ai-ml/course.yaml":                               &fstest.MapFile{Data: []byte("id: ai-ml\ntitle: AI & Machine Learning\ndescription: Learn AI and machine learning.\norder: 0\n")},
		"courses/ai-ml/modules/first/module.yaml":                 &fstest.MapFile{Data: []byte("id: foundations\ntitle: Foundations\norder: 1\nobjectives: []\nvideos: []\nlessons: []\n")},
		"courses/other/course.yaml":                               &fstest.MapFile{Data: []byte("id: other\ntitle: Other course\ndescription: Another course.\norder: 1\n")},
		"courses/other/modules/extra/module.yaml":                 &fstest.MapFile{Data: []byte("id: extra\ntitle: Extra\norder: 0\nobjectives: []\nvideos: []\nlessons: []\n")},
		"courses/ai-ml/modules/storage/module.yaml":               &fstest.MapFile{Data: []byte("id: python\ntitle: Python\norder: 2\nobjectives:\n  - id: python.variables\n    title: Use variables\n    description: Bind names to values.\n    prerequisites: []\nvideos:\n  - id: python-video\n    title: Python video\n    url: https://example.com/python\n    objectiveIds:\n      - python.variables\nlessons:\n  - lesson.stable\n")},
		"courses/ai-ml/modules/storage/not-the-id.mdx":            &fstest.MapFile{Data: []byte("---\nid: lesson.stable\ntitle: Stable lesson\nobjectiveIds:\n  - python.variables\nsourceIds:\n  - go-docs\n---\n# Lesson\n")},
		"courses/ai-ml/modules/storage/worksheets/worksheet.yaml": &fstest.MapFile{Data: []byte("id: worksheet\ntitle: Worksheet\nlessonId: lesson.stable\norder: 0\nobjectiveIds:\n  - python.variables\ninstructions: Complete the worksheet.\nproblems:\n  - id: problem\n    prompt: Solve it.\n    objectiveIds:\n      - python.variables\n    expectedAnswer: Secret answer.\n    requiresWork: true\n    responseLines: 3\n    rubric:\n      - Secret criterion.\n")},
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
