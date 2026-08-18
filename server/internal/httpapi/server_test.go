package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
)

func TestHealthReturnsTypedRepresentation(t *testing.T) {
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()))
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	response := httptest.NewRecorder()

	app.Handler.ServeHTTP(response, req)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}

	var health Health
	if err := json.NewDecoder(response.Body).Decode(&health); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if health.Status != "ok" {
		t.Fatalf("expected health status ok, got %q", health.Status)
	}
}

func TestTutorTurnRouteUsesResourcePathAndStreamsEvents(t *testing.T) {
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()))
	req := httptest.NewRequest(http.MethodPost, "/api/tutor/turns", strings.NewReader(`{"message":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	app.Handler.ServeHTTP(response, req)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("expected SSE content type, got %q", got)
	}
	if !strings.Contains(response.Body.String(), `"type":"text_delta"`) {
		t.Fatalf("expected text delta event, got %s", response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"type":"completed"`) {
		t.Fatalf("expected completed event, got %s", response.Body.String())
	}

	oldReq := httptest.NewRequest(http.MethodPost, "/api/tutor/turn", strings.NewReader(`{"message":"hello"}`))
	oldResponse := httptest.NewRecorder()
	app.Handler.ServeHTTP(oldResponse, oldReq)
	if oldResponse.Code != http.StatusNotFound {
		t.Fatalf("expected old tutor route to be gone, got %d", oldResponse.Code)
	}
}

func TestTutorTurnInvalidInputUsesCommonError(t *testing.T) {
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()))
	req := httptest.NewRequest(http.MethodPost, "/api/tutor/turns", strings.NewReader(`{"message":"   "}`))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	app.Handler.ServeHTTP(response, req)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("expected common problem content type, got %q", got)
	}

	var problem map[string]any
	if err := json.NewDecoder(response.Body).Decode(&problem); err != nil {
		t.Fatalf("decode problem response: %v", err)
	}
	if problem["status"] != float64(http.StatusUnprocessableEntity) {
		t.Fatalf("expected problem status 422, got %#v", problem["status"])
	}
	if problem["detail"] != "message is required" {
		t.Fatalf("expected stable detail, got %#v", problem["detail"])
	}
	if _, ok := problem["error"]; ok {
		t.Fatalf("unexpected endpoint-specific error field: %#v", problem)
	}
}

func TestTutorTurnMalformedJSONUsesBadRequestProblem(t *testing.T) {
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()))
	req := httptest.NewRequest(http.MethodPost, "/api/tutor/turns", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	app.Handler.ServeHTTP(response, req)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("expected common problem content type, got %q", got)
	}
}

func TestTutorTurnProviderFailureUsesBadGatewayProblem(t *testing.T) {
	app := NewAPI(tutor.NewService(failingProvider{}))
	req := httptest.NewRequest(http.MethodPost, "/api/tutor/turns", strings.NewReader(`{"message":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	app.Handler.ServeHTTP(response, req)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("expected common problem content type, got %q", got)
	}
	if strings.Contains(response.Body.String(), "provider secret") {
		t.Fatalf("provider error leaked into response: %s", response.Body.String())
	}
}

func TestOpenAPIContract(t *testing.T) {
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()))
	if app.Spec.OpenAPI != "3.1.0" {
		t.Fatalf("expected OpenAPI 3.1.0, got %q", app.Spec.OpenAPI)
	}

	health, ok := app.Spec.Paths["/api/health"]
	if !ok || health.Get == nil || health.Get.OperationID != "getHealth" {
		t.Fatalf("health operation is missing or unstable: %#v", health)
	}

	tutorPath, ok := app.Spec.Paths["/api/tutor/turns"]
	if !ok || tutorPath.Post == nil || tutorPath.Post.OperationID != "createTutorTurn" {
		t.Fatalf("tutor operation is missing or unstable: %#v", tutorPath)
	}
	if _, ok := app.Spec.Paths["/api/tutor/turn"]; ok {
		t.Fatal("old tutor operation is still present")
	}
	if tutorPath.Post.RequestBody == nil || tutorPath.Post.RequestBody.Content["application/json"] == nil {
		t.Fatal("tutor request body is not documented as JSON")
	}
	if tutorPath.Post.Responses["200"].Content["text/event-stream"] == nil {
		t.Fatal("tutor SSE response is not documented")
	}
	if tutorPath.Post.Responses["422"].Content["application/problem+json"] == nil {
		t.Fatal("common validation error response is not documented")
	}
}

type failingProvider struct{}

func (failingProvider) StreamTurn(context.Context, tutor.TurnRequest) (<-chan tutor.Event, error) {
	return nil, errors.New("provider secret should not be exposed")
}
