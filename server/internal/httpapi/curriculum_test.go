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
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), testCatalog(t), nil, nil)

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
	if module.CourseID != "ai-ml" || module.ID != "python" || len(module.Objectives) != 1 || len(module.Videos) != 1 || len(module.Lessons) != 1 || len(module.Worksheets) != 1 || len(module.Exercises) != 1 {
		t.Fatalf("unexpected module resource: %#v", module)
	}
	if module.Lessons[0].ID != "lesson.stable" || len(module.Lessons[0].ObjectiveIDs) != 1 {
		t.Fatalf("unexpected lesson summary: %#v", module.Lessons[0])
	}
	if video := module.Videos[0]; video.CourseID != "ai-ml" || video.ModuleID != "python" || video.ID != "python-video" || video.YouTubeID != "dQw4w9WgXcQ" || video.Channel != "Python Creator" || video.DurationMinutes != 8 || video.Order != 0 || len(video.ObjectiveIDs) != 1 || len(video.LessonIDs) != 1 || video.LessonIDs[0] != "lesson.stable" {
		t.Fatalf("unexpected video resource: %#v", video)
	}
	if module.Worksheets[0].ID != "worksheet" || module.Worksheets[0].ProblemCount != 1 {
		t.Fatalf("unexpected worksheet summary: %#v", module.Worksheets[0])
	}
	if module.Exercises[0].ID != "python.example" || module.Exercises[0].LessonID != "lesson.stable" {
		t.Fatalf("unexpected exercise summary: %#v", module.Exercises[0])
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
	if len(lesson.Exercises) != 1 || lesson.Exercises[0].ID != "python.example" {
		t.Fatalf("lesson did not expose exercise summaries: %#v", lesson.Exercises)
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

	exerciseResponse := serve(t, app.Handler, http.MethodGet, "/api/courses/ai-ml/modules/python/exercises/python.example")
	if exerciseResponse.Code != http.StatusOK {
		t.Fatalf("get exercise status = %d: %s", exerciseResponse.Code, exerciseResponse.Body.String())
	}
	exerciseJSON := exerciseResponse.Body.String()
	if strings.Contains(exerciseJSON, "hidden-case") || strings.Contains(exerciseJSON, "Hidden check") || strings.Contains(exerciseJSON, "secret_solution") || strings.Contains(exerciseJSON, "visibility") || strings.Contains(exerciseJSON, "tests\"") {
		t.Fatalf("student exercise leaked hidden test data or internal test representation: %s", exerciseJSON)
	}
	var exercise ExerciseResource
	if err := json.Unmarshal([]byte(exerciseJSON), &exercise); err != nil {
		t.Fatalf("decode exercise JSON: %v", err)
	}
	if exercise.CourseID != "ai-ml" || exercise.ModuleID != "python" || exercise.ID != "python.example" || exercise.LessonID != "lesson.stable" || len(exercise.VisibleTests) != 1 || exercise.VisibleTests[0].ID != "visible-case" || exercise.VisibleTests[0].Code != "assert solution()" {
		t.Fatalf("unexpected exercise resource: %#v", exercise)
	}

	reviewItemResponse := serve(t, app.Handler, http.MethodGet, "/api/courses/ai-ml/modules/python/review-items/functions.definition")
	if reviewItemResponse.Code != http.StatusOK {
		t.Fatalf("get review item status = %d: %s", reviewItemResponse.Code, reviewItemResponse.Body.String())
	}
	reviewItemJSON := reviewItemResponse.Body.String()
	if strings.Contains(reviewItemJSON, "due") || strings.Contains(reviewItemJSON, "stability") || strings.Contains(reviewItemJSON, "difficulty") {
		t.Fatalf("authored review item included scheduling state: %s", reviewItemJSON)
	}
	var reviewItem ReviewItemResource
	if err := json.Unmarshal([]byte(reviewItemJSON), &reviewItem); err != nil {
		t.Fatalf("decode review item JSON: %v", err)
	}
	if reviewItem.CourseID != "ai-ml" || reviewItem.ModuleID != "python" || reviewItem.ID != "functions.definition" || reviewItem.SourceLessonID != "lesson.stable" || reviewItem.Answer != "Every input maps to one output." || reviewItem.Hint != "Focus on each input." {
		t.Fatalf("unexpected review item resource: %#v", reviewItem)
	}
}

func TestCurriculumReadAPIMissingResourcesUseProblemResponses(t *testing.T) {
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), testCatalog(t), nil, nil)
	for _, path := range []string{
		"/api/courses/missing",
		"/api/courses/other/modules/python",
		"/api/courses/other/modules/python/lessons/lesson.stable",
		"/api/courses/ai-ml/modules/wrong-module/lessons/lesson.stable",
		"/api/courses/ai-ml/modules/python/lessons/missing",
		"/api/courses/other/modules/python/worksheets/worksheet",
		"/api/courses/ai-ml/modules/wrong-module/worksheets/worksheet",
		"/api/courses/ai-ml/modules/python/worksheets/missing",
		"/api/courses/other/modules/python/exercises/python.example",
		"/api/courses/ai-ml/modules/wrong-module/exercises/python.example",
		"/api/courses/ai-ml/modules/python/exercises/missing",
		"/api/courses/other/modules/python/review-items/functions.definition",
		"/api/courses/ai-ml/modules/wrong-module/review-items/functions.definition",
		"/api/courses/ai-ml/modules/python/review-items/missing",
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
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), curriculum.NewEmptyCatalog(), nil, nil)
	for path, operationID := range map[string]string{
		"/api/courses":                               "listCourses",
		"/api/courses/{courseId}":                    "getCourse",
		"/api/courses/{courseId}/modules/{moduleId}": "getCourseModule",
		"/api/courses/{courseId}/modules/{moduleId}/lessons/{lessonId}":          "getCourseLesson",
		"/api/courses/{courseId}/modules/{moduleId}/worksheets/{worksheetId}":    "getCourseModuleWorksheet",
		"/api/courses/{courseId}/modules/{moduleId}/exercises/{exerciseId}":      "getCourseModuleExercise",
		"/api/courses/{courseId}/modules/{moduleId}/review-items/{reviewItemId}": "getCourseModuleReviewItem",
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
	exerciseSchema := app.Spec.Components.Schemas.Map()["ExerciseResource"]
	if exerciseSchema == nil || exerciseSchema.Properties["visibleTests"] == nil || exerciseSchema.Properties["tests"] != nil {
		t.Fatalf("expected student-safe exercise schema, got %#v", exerciseSchema)
	}
	visibleTestSchema := app.Spec.Components.Schemas.Map()["VisibleExerciseTestResource"]
	if visibleTestSchema == nil || visibleTestSchema.Properties["visibility"] != nil {
		t.Fatalf("visible exercise test schema exposed internal visibility: %#v", visibleTestSchema)
	}
	reviewItemSchema := app.Spec.Components.Schemas.Map()["ReviewItemResource"]
	if reviewItemSchema == nil || reviewItemSchema.AdditionalProperties != false || reviewItemSchema.Properties["answer"] == nil || reviewItemSchema.Properties["sourceLessonId"] == nil {
		t.Fatalf("expected strict authored review item schema, got %#v", reviewItemSchema)
	}
	for _, required := range reviewItemSchema.Required {
		if required == "hint" {
			t.Fatalf("optional review item hint is required in schema: %#v", reviewItemSchema)
		}
	}
	for _, schedulingField := range []string{"due", "stability", "difficulty", "state", "scheduledDays"} {
		if reviewItemSchema.Properties[schedulingField] != nil {
			t.Fatalf("review item schema contains scheduling field %q: %#v", schedulingField, reviewItemSchema)
		}
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
		"courses/ai-ml/modules/storage/module.yaml":               &fstest.MapFile{Data: []byte("id: python\ntitle: Python\norder: 2\nobjectives:\n  - id: python.variables\n    title: Use variables\n    description: Bind names to values.\n    prerequisites: []\nvideos:\n  - id: python-video\n    youtubeId: dQw4w9WgXcQ\n    title: Python video\n    channel: Python Creator\n    durationMinutes: 8\n    order: 0\n    objectiveIds:\n      - python.variables\n    lessonIds:\n      - lesson.stable\nlessons:\n  - lesson.stable\n")},
		"courses/ai-ml/modules/storage/not-the-id.mdx":            &fstest.MapFile{Data: []byte("---\nid: lesson.stable\ntitle: Stable lesson\nobjectiveIds:\n  - python.variables\nsourceIds:\n  - go-docs\n---\n# Lesson\n")},
		"courses/ai-ml/modules/storage/worksheets/worksheet.yaml": &fstest.MapFile{Data: []byte("id: worksheet\ntitle: Worksheet\nlessonId: lesson.stable\norder: 0\nobjectiveIds:\n  - python.variables\ninstructions: Complete the worksheet.\nproblems:\n  - id: problem\n    prompt: Solve it.\n    objectiveIds:\n      - python.variables\n    expectedAnswer: Secret answer.\n    requiresWork: true\n    responseLines: 3\n    rubric:\n      - Secret criterion.\n")},
		"courses/ai-ml/modules/storage/exercises/example.yaml":    &fstest.MapFile{Data: []byte("id: python.example\ntitle: Example exercise\nlessonId: lesson.stable\norder: 0\nobjectiveIds:\n  - python.variables\nprompt: Implement it.\nstarterCode: pass\ntests:\n  - id: visible-case\n    title: Visible check\n    visibility: visible\n    code: assert solution()\n  - id: hidden-case\n    title: Hidden check\n    visibility: hidden\n    code: assert secret_solution()\n")},
		"courses/ai-ml/modules/storage/reviews/functions.yaml":    &fstest.MapFile{Data: []byte("id: functions.definition\norder: 0\nobjectiveIds:\n  - python.variables\nsourceLessonId: lesson.stable\nprompt: What condition defines a function?\nanswer: Every input maps to one output.\nhint: Focus on each input.\n")},
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
