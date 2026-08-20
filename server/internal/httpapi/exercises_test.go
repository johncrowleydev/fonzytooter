package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/johncrowleydev/fonzytooter/server/internal/learner"
)

const exercisePath = "/api/courses/ai-ml/modules/python/exercises/python.example"

func TestExerciseCheckDefinitionIsPurposeSpecific(t *testing.T) {
	app := testLearningAPI(t)
	normal := serve(t, app.Handler, http.MethodGet, exercisePath)
	if normal.Code != http.StatusOK || strings.Contains(normal.Body.String(), "hidden-case") || strings.Contains(normal.Body.String(), "secret_solution") {
		t.Fatalf("normal detail leaked hidden test: %d %s", normal.Code, normal.Body.String())
	}
	response := serve(t, app.Handler, http.MethodGet, exercisePath+"/check-definition")
	if response.Code != http.StatusOK {
		t.Fatalf("check definition status = %d: %s", response.Code, response.Body.String())
	}
	var definition ExerciseCheckDefinitionResource
	decodeJSON(t, response, &definition)
	if len(definition.Tests) != 2 || definition.Tests[0].ID != "visible-case" || definition.Tests[1].ID != "hidden-case" || definition.Tests[1].Visibility != "hidden" {
		t.Fatalf("unexpected check definition: %#v", definition)
	}
	wrong := serve(t, app.Handler, http.MethodGet, "/api/courses/ai-ml/modules/foundations/exercises/python.example/check-definition")
	if wrong.Code != http.StatusNotFound {
		t.Fatalf("expected ownership 404, got %d", wrong.Code)
	}
}

func TestExerciseWorkspaceAndAttemptAPI(t *testing.T) {
	app := testLearningAPI(t)
	workspacePath := exercisePath + "/workspace"
	response := serve(t, app.Handler, http.MethodGet, workspacePath)
	if response.Code != http.StatusOK {
		t.Fatalf("default workspace status = %d: %s", response.Code, response.Body.String())
	}
	var workspace ExerciseWorkspaceResource
	decodeJSON(t, response, &workspace)
	if workspace.Code != "pass" || workspace.UpdatedAt != nil {
		t.Fatalf("unexpected default workspace: %#v", workspace)
	}
	response = serveJSON(t, app.Handler, http.MethodPut, workspacePath, `{"code":"def solution(): return True"}`)
	if response.Code != http.StatusOK {
		t.Fatalf("save workspace status = %d: %s", response.Code, response.Body.String())
	}
	decodeJSON(t, response, &workspace)
	if workspace.UpdatedAt == nil || workspace.Code != "def solution(): return True" {
		t.Fatalf("unexpected saved workspace: %#v", workspace)
	}

	attemptJSON := `{"codeSnapshot":"code","durationMs":12,"results":[{"testId":"visible-case","status":"passed","message":"","durationMs":4},{"testId":"hidden-case","status":"failed","message":"assertion failed","durationMs":5}]}`
	response = serveJSON(t, app.Handler, http.MethodPost, exercisePath+"/attempts", attemptJSON)
	if response.Code != http.StatusCreated || response.Header().Get("Location") == "" {
		t.Fatalf("attempt status = %d, location %q: %s", response.Code, response.Header().Get("Location"), response.Body.String())
	}
	var attempt ExerciseAttemptResource
	decodeJSON(t, response, &attempt)
	if attempt.PassedCount != 1 || attempt.FailedCount != 1 || attempt.AllPassed {
		t.Fatalf("server aggregate mismatch: %#v", attempt)
	}
	activities := serve(t, app.Handler, http.MethodGet, "/api/activities?courseId=ai-ml")
	var activityResources []ActivityResource
	decodeJSON(t, activities, &activityResources)
	if len(activityResources) != 1 || activityResources[0].Kind != learner.ActivityExerciseChecked || activityResources[0].ExerciseID == nil || *activityResources[0].ExerciseID != "python.example" {
		t.Fatalf("missing exercise activity: %#v", activityResources)
	}
	if activityResources[0].ExerciseTitle == nil || *activityResources[0].ExerciseTitle == "" {
		t.Fatalf("exercise activity was not curriculum-enriched: %#v", activityResources[0])
	}
}

func TestExerciseAttemptAPIRejectsUnknownAndDuplicateTests(t *testing.T) {
	app := testLearningAPI(t)
	bodies := []map[string]any{
		{"codeSnapshot": "code", "durationMs": 1, "results": []map[string]any{{"testId": "visible-case", "status": "passed", "message": "", "durationMs": 1}, {"testId": "unknown", "status": "passed", "message": "", "durationMs": 1}}},
		{"codeSnapshot": "code", "durationMs": 1, "results": []map[string]any{{"testId": "visible-case", "status": "passed", "message": "", "durationMs": 1}, {"testId": "visible-case", "status": "passed", "message": "", "durationMs": 1}}},
	}
	for _, body := range bodies {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		response := serveJSON(t, app.Handler, http.MethodPost, exercisePath+"/attempts", string(encoded))
		if response.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d: %s", response.Code, response.Body.String())
		}
	}
}

func TestExerciseOpenAPIContract(t *testing.T) {
	app := testLearningAPI(t)
	definition := app.Spec.Paths["/api/courses/{courseId}/modules/{moduleId}/exercises/{exerciseId}/check-definition"]
	workspace := app.Spec.Paths["/api/courses/{courseId}/modules/{moduleId}/exercises/{exerciseId}/workspace"]
	attempts := app.Spec.Paths["/api/courses/{courseId}/modules/{moduleId}/exercises/{exerciseId}/attempts"]
	if definition == nil || definition.Get == nil || definition.Get.OperationID != "getExerciseCheckDefinition" {
		t.Fatalf("missing check definition operation: %#v", definition)
	}
	if workspace == nil || workspace.Get == nil || workspace.Put == nil {
		t.Fatalf("missing workspace operations: %#v", workspace)
	}
	if attempts == nil || attempts.Post == nil || attempts.Post.OperationID != "createExerciseAttempt" {
		t.Fatalf("missing attempt creation: %#v", attempts)
	}
}
