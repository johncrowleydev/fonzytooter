package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johncrowleydev/fonzytooter/server/internal/database"
	"github.com/johncrowleydev/fonzytooter/server/internal/learner"
	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
)

func TestLearningAPICompletionProgressAndActivity(t *testing.T) {
	app := testLearningAPI(t)
	progressPath := "/api/courses/ai-ml/modules/python/lessons/lesson.stable/progress"

	response := serve(t, app.Handler, http.MethodGet, progressPath)
	if response.Code != http.StatusOK {
		t.Fatalf("expected default progress 200, got %d: %s", response.Code, response.Body.String())
	}
	var initial LessonProgressResource
	decodeJSON(t, response, &initial)
	if initial.Completed || initial.CompletedAt != nil {
		t.Fatalf("expected incomplete default progress, got %#v", initial)
	}

	for range 2 {
		response = serveJSON(t, app.Handler, http.MethodPut, progressPath, `{"completed":true}`)
		if response.Code != http.StatusOK {
			t.Fatalf("expected completion 200, got %d: %s", response.Code, response.Body.String())
		}
	}

	activityResponse := serve(t, app.Handler, http.MethodGet, "/api/activities?courseId=ai-ml&limit=20")
	if activityResponse.Code != http.StatusOK {
		t.Fatalf("expected activity 200, got %d: %s", activityResponse.Code, activityResponse.Body.String())
	}
	var activities []ActivityResource
	decodeJSON(t, activityResponse, &activities)
	if len(activities) != 1 || activities[0].Kind != learner.ActivityLessonCompleted {
		t.Fatalf("expected one completion activity, got %#v", activities)
	}
	if activities[0].LessonTitle == nil || *activities[0].LessonTitle != "Stable lesson" {
		t.Fatalf("expected enriched lesson title, got %#v", activities[0])
	}

	courseResponse := serve(t, app.Handler, http.MethodGet, "/api/courses/ai-ml/progress")
	if courseResponse.Code != http.StatusOK {
		t.Fatalf("expected course progress 200, got %d: %s", courseResponse.Code, courseResponse.Body.String())
	}
	var courseProgress CourseProgressResource
	decodeJSON(t, courseResponse, &courseProgress)
	if len(courseProgress.Objectives) != 1 || !courseProgress.Objectives[0].Introduced {
		t.Fatalf("expected introduced objective, got %#v", courseProgress.Objectives)
	}
	objective := courseProgress.Objectives[0]
	if objective.TransferAssessed || objective.Recall.ReviewsCompleted != 0 || objective.Application.Attempts != 0 {
		t.Fatalf("expected unassessed evidence dimensions, got %#v", objective)
	}

	response = serveJSON(t, app.Handler, http.MethodPut, progressPath, `{"completed":false}`)
	if response.Code != http.StatusOK {
		t.Fatalf("expected uncomplete 200, got %d: %s", response.Code, response.Body.String())
	}
	activityResponse = serve(t, app.Handler, http.MethodGet, "/api/activities?courseId=ai-ml")
	decodeJSON(t, activityResponse, &activities)
	if len(activities) != 1 {
		t.Fatalf("expected completion history to remain, got %#v", activities)
	}
}

func TestLearningAPIVideoCompletionIsIdempotentExposureOnly(t *testing.T) {
	app := testLearningAPI(t)
	path := "/api/courses/ai-ml/modules/python/videos/python-video/progress"

	response := serve(t, app.Handler, http.MethodGet, path)
	if response.Code != http.StatusOK {
		t.Fatalf("default video progress = %d: %s", response.Code, response.Body.String())
	}
	var initial VideoProgressResource
	decodeJSON(t, response, &initial)
	if initial.Completed || initial.CompletedAt != nil {
		t.Fatalf("expected incomplete video, got %#v", initial)
	}

	for range 2 {
		response = serveJSON(t, app.Handler, http.MethodPut, path, `{"completed":true}`)
		if response.Code != http.StatusOK {
			t.Fatalf("complete video = %d: %s", response.Code, response.Body.String())
		}
	}

	activityResponse := serve(t, app.Handler, http.MethodGet, "/api/activities?courseId=ai-ml")
	var activities []ActivityResource
	decodeJSON(t, activityResponse, &activities)
	if len(activities) != 1 || activities[0].Kind != learner.ActivityVideoCompleted || activities[0].VideoID == nil || *activities[0].VideoID != "python-video" {
		t.Fatalf("expected one video completion activity, got %#v", activities)
	}

	courseResponse := serve(t, app.Handler, http.MethodGet, "/api/courses/ai-ml/progress")
	var courseProgress CourseProgressResource
	decodeJSON(t, courseResponse, &courseProgress)
	if courseProgress.CompletedLessonCount != 0 || len(courseProgress.Objectives) != 1 || courseProgress.Objectives[0].Introduced {
		t.Fatalf("video exposure changed learning progress: %#v", courseProgress)
	}
}

func TestLearningAPIMissingIdentityReturnsNotFound(t *testing.T) {
	app := testLearningAPI(t)
	paths := []string{
		"/api/courses/other/modules/python/lessons/lesson.stable/progress",
		"/api/courses/ai-ml/modules/extra/lessons/lesson.stable/progress",
		"/api/courses/ai-ml/modules/python/lessons/missing/progress",
		"/api/courses/ai-ml/modules/python/videos/missing/progress",
		"/api/courses/missing/progress",
		"/api/activities?courseId=missing",
	}
	for _, path := range paths {
		response := serve(t, app.Handler, http.MethodGet, path)
		if response.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for %s, got %d: %s", path, response.Code, response.Body.String())
		}
	}

	response := serve(t, app.Handler, http.MethodGet, "/api/activities")
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected missing course query validation, got %d: %s", response.Code, response.Body.String())
	}
}

func TestLearningOpenAPIContract(t *testing.T) {
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), testCatalog(t), nil, nil)

	lessonPath := app.Spec.Paths["/api/courses/{courseId}/modules/{moduleId}/lessons/{lessonId}/progress"]
	if lessonPath == nil || lessonPath.Get == nil || lessonPath.Get.OperationID != "getLessonProgress" {
		t.Fatalf("missing lesson progress GET operation: %#v", lessonPath)
	}
	if lessonPath.Put == nil || lessonPath.Put.OperationID != "putLessonProgress" {
		t.Fatalf("missing lesson progress PUT operation: %#v", lessonPath)
	}
	coursePath := app.Spec.Paths["/api/courses/{courseId}/progress"]
	if coursePath == nil || coursePath.Get == nil || coursePath.Get.OperationID != "getCourseProgress" {
		t.Fatalf("missing course progress operation: %#v", coursePath)
	}
	activityPath := app.Spec.Paths["/api/activities"]
	if activityPath == nil || activityPath.Get == nil || activityPath.Get.OperationID != "listActivities" {
		t.Fatalf("missing activity collection operation: %#v", activityPath)
	}
	if activityPath.Post != nil {
		t.Fatal("activity API exposes an arbitrary create operation")
	}
	videoPath := app.Spec.Paths["/api/courses/{courseId}/modules/{moduleId}/videos/{videoId}/progress"]
	if videoPath == nil || videoPath.Get == nil || videoPath.Get.OperationID != "getVideoProgress" || videoPath.Put == nil || videoPath.Put.OperationID != "putVideoProgress" {
		t.Fatalf("missing video progress resource operations: %#v", videoPath)
	}
}

func testLearningAPI(t *testing.T) *API {
	t.Helper()
	catalog := testCatalog(t)
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "fonzytooter.db"))
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return newAuthenticatedTestAPI(t, tutor.NewService(tutor.NewUnavailableProvider()), catalog, learner.NewService(db, catalog), nil)
}

func serveJSON(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(response, request)
	return response
}
