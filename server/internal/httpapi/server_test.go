package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
)

func TestNewAPIRejectsNilCatalog(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered != "httpapi.NewAPI: nil curriculum catalog" {
			t.Fatalf("expected nil catalog panic, got %v", recovered)
		}
	}()

	NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), nil, nil, nil)
}

func TestHealthReturnsTypedRepresentation(t *testing.T) {
	app := newAuthenticatedTestAPI(t, tutor.NewService(tutor.NewUnavailableProvider()), curriculum.NewEmptyCatalog(), nil, nil)
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
	app := newAuthenticatedTestAPI(t, tutor.NewService(tutor.NewUnavailableProvider()), curriculum.NewEmptyCatalog(), nil, nil)
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
	body := response.Body.String()
	if !strings.Contains(body, `data: {"type":"text_delta"`) {
		t.Fatalf("expected raw text delta event payload, got %s", body)
	}
	if strings.Contains(body, `data: {"data":`) {
		t.Fatalf("unexpected nested SSE event payload, got %s", body)
	}
	if !strings.Contains(body, `"type":"completed"`) {
		t.Fatalf("expected completed event, got %s", body)
	}

	oldReq := httptest.NewRequest(http.MethodPost, "/api/tutor/turn", strings.NewReader(`{"message":"hello"}`))
	oldResponse := httptest.NewRecorder()
	app.Handler.ServeHTTP(oldResponse, oldReq)
	if oldResponse.Code != http.StatusNotFound {
		t.Fatalf("expected old tutor route to be gone, got %d", oldResponse.Code)
	}
}

func TestTutorTurnInvalidInputUsesCommonError(t *testing.T) {
	app := newAuthenticatedTestAPI(t, tutor.NewService(tutor.NewUnavailableProvider()), curriculum.NewEmptyCatalog(), nil, nil)
	for _, message := range []string{" ", "   "} {
		t.Run(fmt.Sprintf("message %q", message), func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/tutor/turns", strings.NewReader(fmt.Sprintf(`{"message":%q}`, message)))
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
			if _, ok := problem["error"]; ok {
				t.Fatalf("unexpected endpoint-specific error field: %#v", problem)
			}
		})
	}
}

func TestTutorTurnMalformedJSONUsesBadRequestProblem(t *testing.T) {
	app := newAuthenticatedTestAPI(t, tutor.NewService(tutor.NewUnavailableProvider()), curriculum.NewEmptyCatalog(), nil, nil)
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
	app := newAuthenticatedTestAPI(t, tutor.NewService(failingProvider{}), curriculum.NewEmptyCatalog(), nil, nil)
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
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), curriculum.NewEmptyCatalog(), nil, nil)
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
	turnRequestSchema := app.Spec.Components.Schemas.Map()["TurnRequest"]
	if turnRequestSchema == nil {
		t.Fatal("TurnRequest schema is not generated")
	}
	messageSchema := turnRequestSchema.Properties["message"]
	if messageSchema == nil || messageSchema.Pattern != "[^\\s]" {
		t.Fatalf("expected non-whitespace message pattern, got %#v", messageSchema)
	}
	streamResponse := tutorPath.Post.Responses["200"].Content["text/event-stream"]
	if streamResponse == nil {
		t.Fatal("tutor SSE response is not documented")
	}
	if streamResponse.Schema == nil || streamResponse.Schema.Type != huma.TypeArray {
		t.Fatalf("expected tutor SSE response to be an array stream schema, got %#v", streamResponse.Schema)
	}
	if streamResponse.Schema.Items == nil || streamResponse.Schema.Items.Ref != "#/components/schemas/Event" {
		t.Fatalf("expected tutor SSE items to reference Event directly, got %#v", streamResponse.Schema.Items)
	}
	if _, ok := app.Spec.Components.Schemas.Map()["TutorStreamEvent"]; ok {
		t.Fatal("fake TutorStreamEvent wrapper schema is still advertised")
	}
	eventSchema := app.Spec.Components.Schemas.Map()["Event"]
	if eventSchema == nil || eventSchema.Properties["conversationId"] == nil || eventSchema.Properties["toolCallId"] == nil {
		t.Fatalf("normalized tutor event correlation fields are missing: %#v", eventSchema)
	}
	usageSchema := app.Spec.Components.Schemas.Map()["Usage"]
	if usageSchema == nil || usageSchema.Properties["cachedTokens"] == nil || usageSchema.Properties["reasoningTokens"] == nil {
		t.Fatalf("normalized tutor usage fields are missing: %#v", usageSchema)
	}
	if tutorPath.Post.Responses["422"].Content["application/problem+json"] == nil {
		t.Fatal("common validation error response is not documented")
	}
}

type failingProvider struct{}

func (failingProvider) Stream(context.Context, tutor.ModelRequest) (<-chan tutor.ProviderEvent, error) {
	return nil, errors.New("provider secret should not be exposed")
}
