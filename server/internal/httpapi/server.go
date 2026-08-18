package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
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

type ModulePathInput struct {
	ModuleID string `path:"moduleId"`
}

type LessonPathInput struct {
	ModuleID string `path:"moduleId"`
	LessonID string `path:"lessonId"`
}

type ModuleSummary struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Order int    `json:"order"`
}

type ObjectiveResource struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Prerequisites []string `json:"prerequisites" nullable:"false"`
}

type VideoResource struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	URL          string   `json:"url"`
	ObjectiveIDs []string `json:"objectiveIds" nullable:"false"`
}

type LessonSummary struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	ObjectiveIDs []string `json:"objectiveIds" nullable:"false"`
}

type ModuleResource struct {
	ID         string              `json:"id"`
	Title      string              `json:"title"`
	Order      int                 `json:"order"`
	Objectives []ObjectiveResource `json:"objectives" nullable:"false"`
	Videos     []VideoResource     `json:"videos" nullable:"false"`
	Lessons    []LessonSummary     `json:"lessons" nullable:"false"`
}

type SourceResource struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type LessonResource struct {
	ID           string           `json:"id"`
	ModuleID     string           `json:"moduleId"`
	Title        string           `json:"title"`
	ObjectiveIDs []string         `json:"objectiveIds" nullable:"false"`
	Sources      []SourceResource `json:"sources" nullable:"false"`
	Content      string           `json:"content" doc:"Raw MDX source body with YAML frontmatter removed."`
}

type ListModulesResponse struct {
	Body []ModuleSummary
}

type GetModuleResponse struct {
	Body ModuleResource
}

type GetLessonResponse struct {
	Body LessonResource
}

// NewAPI constructs the application handler and registers every documented
// operation on the same Huma API used by the OpenAPI command.
func NewAPI(tutorService *tutor.Service, catalogs ...*curriculum.Catalog) *API {
	catalog := curriculum.NewEmptyCatalog()
	if len(catalogs) > 0 && catalogs[0] != nil {
		catalog = catalogs[0]
	}
	mux := http.NewServeMux()
	config := huma.DefaultConfig("Fonzytooter API", "0.1.0")
	config.OpenAPIPath = ""
	config.DocsPath = ""
	config.SchemasPath = ""

	humaAPI := humago.New(mux, config)
	registerHealth(humaAPI)
	registerTutorTurn(humaAPI, tutorService)
	registerCurriculum(humaAPI, catalog)

	return &API{Handler: mux, Spec: humaAPI.OpenAPI()}
}

func NewServer(address string, tutorService *tutor.Service, catalogs ...*curriculum.Catalog) *http.Server {
	application := NewAPI(tutorService, catalogs...)

	return &http.Server{
		Addr:              address,
		Handler:           application.Handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func registerCurriculum(api huma.API, catalog *curriculum.Catalog) {
	huma.Register[struct{}, ListModulesResponse](api, huma.Operation{
		OperationID: "listModules",
		Method:      http.MethodGet,
		Path:        "/api/modules",
		Summary:     "List curriculum modules",
		Tags:        []string{"curriculum"},
	}, func(context.Context, *struct{}) (*ListModulesResponse, error) {
		modules := catalog.Modules()
		body := make([]ModuleSummary, 0, len(modules))
		for _, module := range modules {
			body = append(body, ModuleSummary{ID: module.ID, Title: module.Title, Order: module.Order})
		}
		return &ListModulesResponse{Body: body}, nil
	})
	api.OpenAPI().Components.Schemas.Map()["ModuleSummaryList"] = &huma.Schema{
		Type:  huma.TypeArray,
		Items: &huma.Schema{Ref: "#/components/schemas/ModuleSummary"},
	}
	api.OpenAPI().Paths["/api/modules"].Get.Responses["200"].Content["application/json"].Schema = &huma.Schema{
		Ref: "#/components/schemas/ModuleSummaryList",
	}

	huma.Register[ModulePathInput, GetModuleResponse](api, huma.Operation{
		OperationID: "getModule",
		Method:      http.MethodGet,
		Path:        "/api/modules/{moduleId}",
		Summary:     "Get a curriculum module",
		Tags:        []string{"curriculum"},
		Errors:      []int{http.StatusNotFound},
	}, func(_ context.Context, input *ModulePathInput) (*GetModuleResponse, error) {
		module, ok := catalog.ModuleByID(input.ModuleID)
		if !ok {
			return nil, huma.Error404NotFound("module not found")
		}
		return &GetModuleResponse{Body: moduleResource(module)}, nil
	})

	huma.Register[LessonPathInput, GetLessonResponse](api, huma.Operation{
		OperationID: "getLesson",
		Method:      http.MethodGet,
		Path:        "/api/modules/{moduleId}/lessons/{lessonId}",
		Summary:     "Get a curriculum lesson",
		Tags:        []string{"curriculum"},
		Errors:      []int{http.StatusNotFound},
	}, func(_ context.Context, input *LessonPathInput) (*GetLessonResponse, error) {
		lesson, ok := catalog.LessonByID(input.ModuleID, input.LessonID)
		if !ok {
			return nil, huma.Error404NotFound("lesson not found")
		}

		sources := make([]SourceResource, 0, len(lesson.SourceIDs))
		for _, sourceID := range lesson.SourceIDs {
			source, sourceOK := catalog.SourceByID(sourceID)
			if !sourceOK {
				return nil, huma.Error500InternalServerError("curriculum source unavailable")
			}
			sources = append(sources, SourceResource{ID: source.ID, Title: source.Title, URL: source.URL})
		}
		return &GetLessonResponse{Body: LessonResource{
			ID:           lesson.ID,
			ModuleID:     input.ModuleID,
			Title:        lesson.Title,
			ObjectiveIDs: append([]string{}, lesson.ObjectiveIDs...),
			Sources:      sources,
			Content:      lesson.Content,
		}}, nil
	})
}

func moduleResource(module curriculum.Module) ModuleResource {
	objectives := make([]ObjectiveResource, 0, len(module.Objectives))
	for _, objective := range module.Objectives {
		objectives = append(objectives, ObjectiveResource{
			ID:            objective.ID,
			Title:         objective.Title,
			Description:   objective.Description,
			Prerequisites: append([]string{}, objective.Prerequisites...),
		})
	}

	videos := make([]VideoResource, 0, len(module.Videos))
	for _, video := range module.Videos {
		videos = append(videos, VideoResource{
			ID:           video.ID,
			Title:        video.Title,
			URL:          video.URL,
			ObjectiveIDs: append([]string{}, video.ObjectiveIDs...),
		})
	}

	lessons := make([]LessonSummary, 0, len(module.Lessons))
	for _, lesson := range module.Lessons {
		lessons = append(lessons, LessonSummary{
			ID:           lesson.ID,
			Title:        lesson.Title,
			ObjectiveIDs: append([]string{}, lesson.ObjectiveIDs...),
		})
	}

	return ModuleResource{
		ID:         module.ID,
		Title:      module.Title,
		Order:      module.Order,
		Objectives: objectives,
		Videos:     videos,
		Lessons:    lessons,
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
