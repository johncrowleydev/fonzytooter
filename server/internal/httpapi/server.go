package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/learner"
	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
	"github.com/johncrowleydev/fonzytooter/server/internal/worksheetpdf"
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

type CoursePathInput struct {
	CourseID string `path:"courseId"`
}

type CourseModulePathInput struct {
	CourseID string `path:"courseId"`
	ModuleID string `path:"moduleId"`
}

type CourseLessonPathInput struct {
	CourseID string `path:"courseId"`
	ModuleID string `path:"moduleId"`
	LessonID string `path:"lessonId"`
}

type CourseWorksheetPathInput struct {
	CourseID    string `path:"courseId"`
	ModuleID    string `path:"moduleId"`
	WorksheetID string `path:"worksheetId"`
}

type CourseExercisePathInput struct {
	CourseID   string `path:"courseId"`
	ModuleID   string `path:"moduleId"`
	ExerciseID string `path:"exerciseId"`
}

type CourseReviewItemPathInput struct {
	CourseID     string `path:"courseId"`
	ModuleID     string `path:"moduleId"`
	ReviewItemID string `path:"reviewItemId"`
}

type CourseWorksheetDocumentPathInput struct {
	CourseID    string `path:"courseId"`
	ModuleID    string `path:"moduleId"`
	WorksheetID string `path:"worksheetId"`
	DocumentID  string `path:"documentId"`
}

type CourseModuleWorkbookPathInput struct {
	CourseID   string `path:"courseId"`
	ModuleID   string `path:"moduleId"`
	WorkbookID string `path:"workbookId"`
}

type CourseSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Order       int    `json:"order"`
}

type ModuleSummary struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Order int    `json:"order"`
}

type CourseResource struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Order       int             `json:"order"`
	Modules     []ModuleSummary `json:"modules" nullable:"false"`
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

type WorksheetSummary struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	LessonID     string   `json:"lessonId"`
	Order        int      `json:"order"`
	ObjectiveIDs []string `json:"objectiveIds" nullable:"false"`
	ProblemCount int      `json:"problemCount"`
}

type ExerciseSummary struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	LessonID     string   `json:"lessonId"`
	Order        int      `json:"order"`
	ObjectiveIDs []string `json:"objectiveIds" nullable:"false"`
}

type ModuleResource struct {
	CourseID   string              `json:"courseId"`
	ID         string              `json:"id"`
	Title      string              `json:"title"`
	Order      int                 `json:"order"`
	Objectives []ObjectiveResource `json:"objectives" nullable:"false"`
	Videos     []VideoResource     `json:"videos" nullable:"false"`
	Lessons    []LessonSummary     `json:"lessons" nullable:"false"`
	Worksheets []WorksheetSummary  `json:"worksheets" nullable:"false"`
	Exercises  []ExerciseSummary   `json:"exercises" nullable:"false"`
}

type SourceResource struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type LessonResource struct {
	CourseID     string             `json:"courseId"`
	ID           string             `json:"id"`
	ModuleID     string             `json:"moduleId"`
	Title        string             `json:"title"`
	ObjectiveIDs []string           `json:"objectiveIds" nullable:"false"`
	Sources      []SourceResource   `json:"sources" nullable:"false"`
	Content      string             `json:"content" doc:"Raw MDX source body with YAML frontmatter removed."`
	Worksheets   []WorksheetSummary `json:"worksheets" nullable:"false"`
	Exercises    []ExerciseSummary  `json:"exercises" nullable:"false"`
}

type WorksheetProblemResource struct {
	ID            string   `json:"id"`
	Prompt        string   `json:"prompt"`
	ObjectiveIDs  []string `json:"objectiveIds" nullable:"false"`
	RequiresWork  bool     `json:"requiresWork"`
	ResponseLines int      `json:"responseLines"`
}

type WorksheetResource struct {
	CourseID     string                     `json:"courseId"`
	ModuleID     string                     `json:"moduleId"`
	ID           string                     `json:"id"`
	Title        string                     `json:"title"`
	LessonID     string                     `json:"lessonId"`
	Order        int                        `json:"order"`
	ObjectiveIDs []string                   `json:"objectiveIds" nullable:"false"`
	Instructions string                     `json:"instructions"`
	Problems     []WorksheetProblemResource `json:"problems" nullable:"false"`
}

type VisibleExerciseTestResource struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Code  string `json:"code"`
}

type ExerciseResource struct {
	CourseID     string                        `json:"courseId"`
	ModuleID     string                        `json:"moduleId"`
	ID           string                        `json:"id"`
	Title        string                        `json:"title"`
	LessonID     string                        `json:"lessonId"`
	Order        int                           `json:"order"`
	ObjectiveIDs []string                      `json:"objectiveIds" nullable:"false"`
	Prompt       string                        `json:"prompt"`
	StarterCode  string                        `json:"starterCode"`
	VisibleTests []VisibleExerciseTestResource `json:"visibleTests" nullable:"false"`
}

type ReviewItemResource struct {
	CourseID       string   `json:"courseId"`
	ModuleID       string   `json:"moduleId"`
	ID             string   `json:"id"`
	Order          int      `json:"order"`
	ObjectiveIDs   []string `json:"objectiveIds" nullable:"false"`
	SourceLessonID string   `json:"sourceLessonId"`
	Prompt         string   `json:"prompt"`
	Answer         string   `json:"answer"`
	Hint           string   `json:"hint,omitempty"`
}

type ListCoursesResponse struct {
	Body []CourseSummary
}

type GetCourseResponse struct {
	Body CourseResource
}

type GetCourseModuleResponse struct {
	Body ModuleResource
}

type GetCourseLessonResponse struct {
	Body LessonResource
}

type GetCourseWorksheetResponse struct {
	Body WorksheetResource
}

type GetCourseExerciseResponse struct {
	Body ExerciseResource
}

type GetCourseReviewItemResponse struct {
	Body ReviewItemResource
}

// NewAPI constructs the application handler and registers every documented
// operation on the same Huma API used by the OpenAPI command.
func NewAPI(tutorService *tutor.Service, catalog *curriculum.Catalog, learnerService *learner.Service) *API {
	return newAPI(tutorService, catalog, learnerService, worksheetpdf.NewRenderer())
}

type worksheetDocumentRenderer interface {
	Render(context.Context, curriculum.Worksheet, worksheetpdf.Variant) ([]byte, error)
	RenderWorkbook(context.Context, curriculum.Course, curriculum.Module, worksheetpdf.Variant) ([]byte, error)
}

func newAPI(tutorService *tutor.Service, catalog *curriculum.Catalog, learnerService *learner.Service, documentRenderer worksheetDocumentRenderer) *API {
	if catalog == nil {
		panic("httpapi.NewAPI: nil curriculum catalog")
	}
	mux := http.NewServeMux()
	config := huma.DefaultConfig("Fonzytooter API", "0.1.0")
	config.OpenAPIPath = ""
	config.DocsPath = ""
	config.SchemasPath = ""

	humaAPI := humago.New(mux, config)
	registerHealth(humaAPI)
	registerTutorTurn(humaAPI, tutorService)
	registerCurriculum(humaAPI, catalog, documentRenderer)
	registerLearning(humaAPI, learnerService)

	return &API{Handler: mux, Spec: humaAPI.OpenAPI()}
}

func NewServer(address string, tutorService *tutor.Service, catalog *curriculum.Catalog, learnerService *learner.Service) *http.Server {
	application := NewAPI(tutorService, catalog, learnerService)

	return &http.Server{
		Addr:              address,
		Handler:           application.Handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func registerCurriculum(api huma.API, catalog *curriculum.Catalog, documentRenderer worksheetDocumentRenderer) {
	huma.Register[struct{}, ListCoursesResponse](api, huma.Operation{
		OperationID: "listCourses",
		Method:      http.MethodGet,
		Path:        "/api/courses",
		Summary:     "List curriculum courses",
		Tags:        []string{"curriculum"},
	}, func(context.Context, *struct{}) (*ListCoursesResponse, error) {
		courses := catalog.Courses()
		body := make([]CourseSummary, 0, len(courses))
		for _, course := range courses {
			body = append(body, CourseSummary{
				ID:          course.ID,
				Title:       course.Title,
				Description: course.Description,
				Order:       course.Order,
			})
		}
		return &ListCoursesResponse{Body: body}, nil
	})
	api.OpenAPI().Components.Schemas.Map()["CourseSummaryList"] = &huma.Schema{
		Type:  huma.TypeArray,
		Items: &huma.Schema{Ref: "#/components/schemas/CourseSummary"},
	}
	api.OpenAPI().Paths["/api/courses"].Get.Responses["200"].Content["application/json"].Schema = &huma.Schema{
		Ref: "#/components/schemas/CourseSummaryList",
	}

	huma.Register[CoursePathInput, GetCourseResponse](api, huma.Operation{
		OperationID: "getCourse",
		Method:      http.MethodGet,
		Path:        "/api/courses/{courseId}",
		Summary:     "Get a curriculum course",
		Tags:        []string{"curriculum"},
		Errors:      []int{http.StatusNotFound},
	}, func(_ context.Context, input *CoursePathInput) (*GetCourseResponse, error) {
		course, ok := catalog.CourseByID(input.CourseID)
		if !ok {
			return nil, huma.Error404NotFound("course not found")
		}
		modules := make([]ModuleSummary, 0, len(course.Modules))
		for _, module := range course.Modules {
			modules = append(modules, ModuleSummary{ID: module.ID, Title: module.Title, Order: module.Order})
		}
		return &GetCourseResponse{Body: CourseResource{
			ID:          course.ID,
			Title:       course.Title,
			Description: course.Description,
			Order:       course.Order,
			Modules:     modules,
		}}, nil
	})

	huma.Register[CourseModulePathInput, GetCourseModuleResponse](api, huma.Operation{
		OperationID: "getCourseModule",
		Method:      http.MethodGet,
		Path:        "/api/courses/{courseId}/modules/{moduleId}",
		Summary:     "Get a course module",
		Tags:        []string{"curriculum"},
		Errors:      []int{http.StatusNotFound},
	}, func(_ context.Context, input *CourseModulePathInput) (*GetCourseModuleResponse, error) {
		module, ok := catalog.ModuleByCourse(input.CourseID, input.ModuleID)
		if !ok {
			return nil, huma.Error404NotFound("module not found")
		}
		return &GetCourseModuleResponse{Body: moduleResource(module)}, nil
	})

	huma.Register[CourseLessonPathInput, GetCourseLessonResponse](api, huma.Operation{
		OperationID: "getCourseLesson",
		Method:      http.MethodGet,
		Path:        "/api/courses/{courseId}/modules/{moduleId}/lessons/{lessonId}",
		Summary:     "Get a course lesson",
		Tags:        []string{"curriculum"},
		Errors:      []int{http.StatusNotFound},
	}, func(_ context.Context, input *CourseLessonPathInput) (*GetCourseLessonResponse, error) {
		lesson, ok := catalog.LessonByCourse(input.CourseID, input.ModuleID, input.LessonID)
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
		return &GetCourseLessonResponse{Body: LessonResource{
			CourseID:     input.CourseID,
			ID:           lesson.ID,
			ModuleID:     input.ModuleID,
			Title:        lesson.Title,
			ObjectiveIDs: append([]string{}, lesson.ObjectiveIDs...),
			Sources:      sources,
			Content:      lesson.Content,
			Worksheets:   worksheetSummaries(catalog.WorksheetsByLesson(input.CourseID, input.ModuleID, input.LessonID)),
			Exercises:    exerciseSummaries(catalog.ExercisesByLesson(input.CourseID, input.ModuleID, input.LessonID)),
		}}, nil
	})

	huma.Register[CourseExercisePathInput, GetCourseExerciseResponse](api, huma.Operation{
		OperationID: "getCourseModuleExercise",
		Method:      http.MethodGet,
		Path:        "/api/courses/{courseId}/modules/{moduleId}/exercises/{exerciseId}",
		Summary:     "Get an embedded course exercise",
		Tags:        []string{"curriculum"},
		Errors:      []int{http.StatusNotFound},
	}, func(_ context.Context, input *CourseExercisePathInput) (*GetCourseExerciseResponse, error) {
		exercise, ok := catalog.ExerciseByCourse(input.CourseID, input.ModuleID, input.ExerciseID)
		if !ok {
			return nil, huma.Error404NotFound("exercise not found")
		}
		visibleTests := make([]VisibleExerciseTestResource, 0, len(exercise.Tests))
		for _, test := range exercise.Tests {
			if test.Visibility != curriculum.ExerciseTestVisible {
				continue
			}
			visibleTests = append(visibleTests, VisibleExerciseTestResource{ID: test.ID, Title: test.Title, Code: test.Code})
		}
		return &GetCourseExerciseResponse{Body: ExerciseResource{
			CourseID: exercise.CourseID, ModuleID: exercise.ModuleID,
			ID: exercise.ID, Title: exercise.Title, LessonID: exercise.LessonID, Order: exercise.Order,
			ObjectiveIDs: append([]string{}, exercise.ObjectiveIDs...), Prompt: exercise.Prompt,
			StarterCode: exercise.StarterCode, VisibleTests: visibleTests,
		}}, nil
	})

	huma.Register[CourseReviewItemPathInput, GetCourseReviewItemResponse](api, huma.Operation{
		OperationID: "getCourseModuleReviewItem",
		Method:      http.MethodGet,
		Path:        "/api/courses/{courseId}/modules/{moduleId}/review-items/{reviewItemId}",
		Summary:     "Get an authored review item",
		Tags:        []string{"curriculum"},
		Errors:      []int{http.StatusNotFound},
	}, func(_ context.Context, input *CourseReviewItemPathInput) (*GetCourseReviewItemResponse, error) {
		reviewItem, ok := catalog.ReviewItemByCourse(input.CourseID, input.ModuleID, input.ReviewItemID)
		if !ok {
			return nil, huma.Error404NotFound("review item not found")
		}
		return &GetCourseReviewItemResponse{Body: ReviewItemResource{
			CourseID:       reviewItem.CourseID,
			ModuleID:       reviewItem.ModuleID,
			ID:             reviewItem.ID,
			Order:          reviewItem.Order,
			ObjectiveIDs:   append([]string{}, reviewItem.ObjectiveIDs...),
			SourceLessonID: reviewItem.SourceLessonID,
			Prompt:         reviewItem.Prompt,
			Answer:         reviewItem.Answer,
			Hint:           reviewItem.Hint,
		}}, nil
	})

	huma.Register[CourseWorksheetPathInput, GetCourseWorksheetResponse](api, huma.Operation{
		OperationID: "getCourseModuleWorksheet",
		Method:      http.MethodGet,
		Path:        "/api/courses/{courseId}/modules/{moduleId}/worksheets/{worksheetId}",
		Summary:     "Get a course module worksheet",
		Tags:        []string{"curriculum"},
		Errors:      []int{http.StatusNotFound},
	}, func(_ context.Context, input *CourseWorksheetPathInput) (*GetCourseWorksheetResponse, error) {
		worksheet, ok := catalog.WorksheetByCourse(input.CourseID, input.ModuleID, input.WorksheetID)
		if !ok {
			return nil, huma.Error404NotFound("worksheet not found")
		}
		problems := make([]WorksheetProblemResource, 0, len(worksheet.Problems))
		for _, problem := range worksheet.Problems {
			problems = append(problems, WorksheetProblemResource{
				ID:            problem.ID,
				Prompt:        problem.Prompt,
				ObjectiveIDs:  append([]string{}, problem.ObjectiveIDs...),
				RequiresWork:  problem.RequiresWork,
				ResponseLines: problem.ResponseLines,
			})
		}
		return &GetCourseWorksheetResponse{Body: WorksheetResource{
			CourseID:     worksheet.CourseID,
			ModuleID:     worksheet.ModuleID,
			ID:           worksheet.ID,
			Title:        worksheet.Title,
			LessonID:     worksheet.LessonID,
			Order:        worksheet.Order,
			ObjectiveIDs: append([]string{}, worksheet.ObjectiveIDs...),
			Instructions: worksheet.Instructions,
			Problems:     problems,
		}}, nil
	})

	huma.Register[CourseWorksheetDocumentPathInput, huma.StreamResponse](api, huma.Operation{
		OperationID: "getCourseModuleWorksheetDocument",
		Method:      http.MethodGet,
		Path:        "/api/courses/{courseId}/modules/{moduleId}/worksheets/{worksheetId}/documents/{documentId}",
		Summary:     "Get a printable worksheet document",
		Description: "Returns either the student worksheet or its solutions as a generated PDF.",
		Tags:        []string{"curriculum"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError, http.StatusServiceUnavailable},
		Responses: map[string]*huma.Response{
			"200": {
				Description: http.StatusText(http.StatusOK),
				Headers: map[string]*huma.Param{
					"Content-Disposition": {
						Description: "Attachment disposition with a deterministic worksheet filename.",
						Schema:      &huma.Schema{Type: huma.TypeString},
					},
				},
				Content: map[string]*huma.MediaType{
					"application/pdf": {
						Schema: &huma.Schema{Type: huma.TypeString, Format: "binary"},
					},
				},
			},
		},
	}, func(ctx context.Context, input *CourseWorksheetDocumentPathInput) (*huma.StreamResponse, error) {
		variant, err := worksheetpdf.ParseVariant(input.DocumentID)
		if err != nil {
			return nil, huma.Error404NotFound("worksheet document not found")
		}
		worksheet, ok := catalog.WorksheetByCourse(input.CourseID, input.ModuleID, input.WorksheetID)
		if !ok {
			return nil, huma.Error404NotFound("worksheet not found")
		}
		if documentRenderer == nil {
			return nil, huma.Error503ServiceUnavailable("worksheet PDF rendering is unavailable")
		}

		pdf, err := documentRenderer.Render(ctx, worksheet, variant)
		if errors.Is(err, worksheetpdf.ErrToolUnavailable) {
			log.Printf("worksheet PDF tooling unavailable: %v", err)
			return nil, huma.Error503ServiceUnavailable("worksheet PDF rendering is unavailable")
		}
		if err != nil {
			log.Printf("render worksheet PDF %s/%s/%s/%s: %v", input.CourseID, input.ModuleID, input.WorksheetID, input.DocumentID, err)
			return nil, huma.Error500InternalServerError("worksheet PDF rendering failed")
		}
		filename, err := worksheetpdf.Filename(worksheet.ID, variant)
		if err != nil {
			log.Printf("choose worksheet PDF filename: %v", err)
			return nil, huma.Error500InternalServerError("worksheet PDF rendering failed")
		}

		return &huma.StreamResponse{Body: func(streamContext huma.Context) {
			streamContext.SetHeader("Content-Type", "application/pdf")
			streamContext.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
			streamContext.SetStatus(http.StatusOK)
			_, _ = streamContext.BodyWriter().Write(pdf)
		}}, nil
	})

	huma.Register[CourseModuleWorkbookPathInput, huma.StreamResponse](api, huma.Operation{
		OperationID: "getCourseModuleWorkbook",
		Method:      http.MethodGet,
		Path:        "/api/courses/{courseId}/modules/{moduleId}/workbooks/{workbookId}",
		Summary:     "Get a printable module worksheet workbook",
		Description: "Returns the module's student workbook or solutions workbook as one generated PDF.",
		Tags:        []string{"curriculum"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError, http.StatusServiceUnavailable},
		Responses: map[string]*huma.Response{
			"200": {
				Description: http.StatusText(http.StatusOK),
				Headers: map[string]*huma.Param{
					"Content-Disposition": {
						Description: "Attachment disposition with a deterministic module workbook filename.",
						Schema:      &huma.Schema{Type: huma.TypeString},
					},
				},
				Content: map[string]*huma.MediaType{
					"application/pdf": {
						Schema: &huma.Schema{Type: huma.TypeString, Format: "binary"},
					},
				},
			},
		},
	}, func(ctx context.Context, input *CourseModuleWorkbookPathInput) (*huma.StreamResponse, error) {
		variant, err := worksheetpdf.ParseVariant(input.WorkbookID)
		if err != nil {
			return nil, huma.Error404NotFound("module workbook not found")
		}
		course, ok := catalog.CourseByID(input.CourseID)
		if !ok {
			return nil, huma.Error404NotFound("course not found")
		}
		module, ok := catalog.ModuleByCourse(input.CourseID, input.ModuleID)
		if !ok || len(module.Worksheets) == 0 {
			return nil, huma.Error404NotFound("module workbook not found")
		}
		if documentRenderer == nil {
			return nil, huma.Error503ServiceUnavailable("worksheet PDF rendering is unavailable")
		}

		pdf, err := documentRenderer.RenderWorkbook(ctx, course, module, variant)
		if errors.Is(err, worksheetpdf.ErrToolUnavailable) {
			log.Printf("worksheet PDF tooling unavailable: %v", err)
			return nil, huma.Error503ServiceUnavailable("worksheet PDF rendering is unavailable")
		}
		if err != nil {
			log.Printf("render module workbook %s/%s/%s: %v", input.CourseID, input.ModuleID, input.WorkbookID, err)
			return nil, huma.Error500InternalServerError("module workbook rendering failed")
		}
		filename, err := worksheetpdf.WorkbookFilename(module.ID, variant)
		if err != nil {
			log.Printf("choose module workbook filename: %v", err)
			return nil, huma.Error500InternalServerError("module workbook rendering failed")
		}

		return &huma.StreamResponse{Body: func(streamContext huma.Context) {
			streamContext.SetHeader("Content-Type", "application/pdf")
			streamContext.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
			streamContext.SetStatus(http.StatusOK)
			_, _ = streamContext.BodyWriter().Write(pdf)
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
		CourseID:   module.CourseID,
		ID:         module.ID,
		Title:      module.Title,
		Order:      module.Order,
		Objectives: objectives,
		Videos:     videos,
		Lessons:    lessons,
		Worksheets: worksheetSummaries(module.Worksheets),
		Exercises:  exerciseSummaries(module.Exercises),
	}
}

func exerciseSummaries(exercises []curriculum.Exercise) []ExerciseSummary {
	summaries := make([]ExerciseSummary, 0, len(exercises))
	for _, exercise := range exercises {
		summaries = append(summaries, ExerciseSummary{
			ID: exercise.ID, Title: exercise.Title, LessonID: exercise.LessonID,
			Order: exercise.Order, ObjectiveIDs: append([]string{}, exercise.ObjectiveIDs...),
		})
	}
	return summaries
}

func worksheetSummaries(worksheets []curriculum.Worksheet) []WorksheetSummary {
	summaries := make([]WorksheetSummary, 0, len(worksheets))
	for _, worksheet := range worksheets {
		summaries = append(summaries, WorksheetSummary{
			ID:           worksheet.ID,
			Title:        worksheet.Title,
			LessonID:     worksheet.LessonID,
			Order:        worksheet.Order,
			ObjectiveIDs: append([]string{}, worksheet.ObjectiveIDs...),
			ProblemCount: len(worksheet.Problems),
		})
	}
	return summaries
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
