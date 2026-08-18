package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
)

type API struct {
	Handler http.Handler
	Spec    *huma.OpenAPI
}

type Health struct {
	Status string `json:"status" enum:"ok"`
}

type HealthResponse struct {
	Body Health
}

type TutorTurnInput struct {
	Body tutor.TurnRequest
}

// NewAPI constructs the application handler and registers every documented
// operation on the same Huma API used by the OpenAPI command.
func NewAPI(tutorService *tutor.Service) *API {
	mux := http.NewServeMux()
	config := huma.DefaultConfig("Fonzytooter API", "0.1.0")
	config.OpenAPIPath = ""
	config.DocsPath = ""
	config.SchemasPath = ""

	humaAPI := humago.New(mux, config)
	registerHealth(humaAPI)
	registerTutorTurn(humaAPI, tutorService)

	return &API{Handler: mux, Spec: humaAPI.OpenAPI()}
}

func NewServer(address string, tutorService *tutor.Service) *http.Server {
	application := NewAPI(tutorService)

	return &http.Server{
		Addr:              address,
		Handler:           application.Handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func registerHealth(api huma.API) {
	huma.Register[struct{}, HealthResponse](api, huma.Operation{
		OperationID: "getHealth",
		Method:      http.MethodGet,
		Path:        "/api/health",
		Summary:     "Get API health",
		Tags:        []string{"health"},
	}, func(context.Context, *struct{}) (*HealthResponse, error) {
		return &HealthResponse{Body: Health{Status: "ok"}}, nil
	})
}

func registerTutorTurn(api huma.API, tutorService *tutor.Service) {
	eventSchema := api.OpenAPI().Components.Schemas.Schema(reflect.TypeOf(tutor.Event{}), true, "TutorEvent")
	api.OpenAPI().Components.Schemas.Schema(reflect.TypeOf(tutor.TurnRequest{}), true, "TutorTurnRequest")
	streamSchema := &huma.Schema{
		Type:        huma.TypeArray,
		Description: "Each item describes one possible Server-Sent Event message serialized as UTF-8 text.",
		Items:       eventSchema,
	}

	huma.Register[TutorTurnInput, huma.StreamResponse](api, huma.Operation{
		OperationID:   "createTutorTurn",
		Method:        http.MethodPost,
		Path:          "/api/tutor/turns",
		Summary:       "Create a tutor turn and stream its events",
		Tags:          []string{"tutor"},
		DefaultStatus: http.StatusOK,
		Errors:        []int{http.StatusBadRequest, http.StatusBadGateway},
		Responses: map[string]*huma.Response{
			"200": {
				Description: http.StatusText(http.StatusOK),
				Content: map[string]*huma.MediaType{
					"text/event-stream": {Schema: streamSchema},
				},
			},
		},
	}, func(ctx context.Context, input *TutorTurnInput) (*huma.StreamResponse, error) {
		if strings.TrimSpace(input.Body.Message) == "" {
			return nil, huma.Error422UnprocessableEntity(
				"message is required",
				&huma.ErrorDetail{Location: "body.message", Value: input.Body.Message},
			)
		}
		if tutorService == nil {
			return nil, huma.Error500InternalServerError("tutor service is unavailable")
		}

		events, err := tutorService.StreamTurn(ctx, input.Body)
		if err != nil {
			return nil, huma.Error502BadGateway("tutor provider unavailable")
		}

		return &huma.StreamResponse{Body: func(streamContext huma.Context) {
			streamTutorEvents(streamContext, events)
		}}, nil
	})
}

func streamTutorEvents(ctx huma.Context, events <-chan tutor.Event) {
	ctx.SetHeader("Content-Type", "text/event-stream")
	ctx.SetHeader("Cache-Control", "no-cache")
	ctx.SetHeader("Connection", "keep-alive")
	ctx.SetStatus(http.StatusOK)

	_, writer := humago.Unwrap(ctx)
	flusher, ok := writer.(http.Flusher)
	if !ok {
		return
	}

	encoder := json.NewEncoder(writer)
	for {
		select {
		case <-ctx.Context().Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			if _, err := fmt.Fprint(writer, "data: "); err != nil {
				return
			}
			if err := encoder.Encode(event); err != nil {
				return
			}
			if _, err := fmt.Fprintln(writer); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
