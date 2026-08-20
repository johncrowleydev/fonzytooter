package httpapi

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/learner"
)

type ExerciseCheckTestResource struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Visibility string `json:"visibility" enum:"visible,hidden"`
	Code       string `json:"code"`
}

type ExerciseCheckDefinitionResource struct {
	CourseID   string                      `json:"courseId"`
	ModuleID   string                      `json:"moduleId"`
	ExerciseID string                      `json:"exerciseId"`
	Tests      []ExerciseCheckTestResource `json:"tests" nullable:"false"`
}

type ExerciseCheckDefinitionResponse struct {
	Body ExerciseCheckDefinitionResource
}

type ExerciseWorkspaceResource struct {
	CourseID   string     `json:"courseId"`
	ModuleID   string     `json:"moduleId"`
	ExerciseID string     `json:"exerciseId"`
	Code       string     `json:"code"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`
}

type ExerciseWorkspaceResponse struct {
	Body ExerciseWorkspaceResource
}

type PutExerciseWorkspaceInput struct {
	CourseExercisePathInput
	Body struct {
		Code string `json:"code"`
	}
}

type ExerciseAttemptTestResultResource struct {
	TestID     string `json:"testId"`
	Status     string `json:"status" enum:"passed,failed,error"`
	Message    string `json:"message"`
	DurationMS int64  `json:"durationMs" minimum:"0"`
}

type CreateExerciseAttemptInput struct {
	CourseExercisePathInput
	Body struct {
		CodeSnapshot string                              `json:"codeSnapshot"`
		DurationMS   int64                               `json:"durationMs" minimum:"0"`
		Results      []ExerciseAttemptTestResultResource `json:"results" minItems:"1" nullable:"false"`
	}
}

type ExerciseAttemptResource struct {
	ID           int64                               `json:"id"`
	CourseID     string                              `json:"courseId"`
	ModuleID     string                              `json:"moduleId"`
	ExerciseID   string                              `json:"exerciseId"`
	CreatedAt    time.Time                           `json:"createdAt"`
	PassedCount  int                                 `json:"passedCount"`
	FailedCount  int                                 `json:"failedCount"`
	DurationMS   int64                               `json:"durationMs"`
	AllPassed    bool                                `json:"allPassed"`
	CodeSnapshot string                              `json:"codeSnapshot"`
	Results      []ExerciseAttemptTestResultResource `json:"results" nullable:"false"`
}

type CreateExerciseAttemptResponse struct {
	Location string `header:"Location"`
	Body     ExerciseAttemptResource
}

func registerExercises(api huma.API, catalog *curriculum.Catalog, service *learner.Service) {
	huma.Register[CourseExercisePathInput, ExerciseCheckDefinitionResponse](api, huma.Operation{
		OperationID: "getExerciseCheckDefinition",
		Method:      http.MethodGet,
		Path:        "/api/courses/{courseId}/modules/{moduleId}/exercises/{exerciseId}/check-definition",
		Summary:     "Get an exercise check definition",
		Tags:        []string{"learner"},
		Errors:      []int{http.StatusNotFound},
	}, func(_ context.Context, input *CourseExercisePathInput) (*ExerciseCheckDefinitionResponse, error) {
		exercise, ok := catalog.ExerciseByCourse(input.CourseID, input.ModuleID, input.ExerciseID)
		if !ok {
			return nil, huma.Error404NotFound("exercise not found")
		}
		tests := make([]ExerciseCheckTestResource, 0, len(exercise.Tests))
		for _, test := range exercise.Tests {
			tests = append(tests, ExerciseCheckTestResource{ID: test.ID, Title: test.Title, Visibility: test.Visibility, Code: test.Code})
		}
		return &ExerciseCheckDefinitionResponse{Body: ExerciseCheckDefinitionResource{
			CourseID: input.CourseID, ModuleID: input.ModuleID, ExerciseID: input.ExerciseID, Tests: tests,
		}}, nil
	})

	huma.Register[CourseExercisePathInput, ExerciseWorkspaceResponse](api, huma.Operation{
		OperationID: "getExerciseWorkspace",
		Method:      http.MethodGet,
		Path:        "/api/courses/{courseId}/modules/{moduleId}/exercises/{exerciseId}/workspace",
		Summary:     "Get an exercise workspace",
		Tags:        []string{"learner"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}, func(ctx context.Context, input *CourseExercisePathInput) (*ExerciseWorkspaceResponse, error) {
		if service == nil {
			return nil, huma.Error500InternalServerError("learner service is unavailable")
		}
		workspace, err := service.ExerciseWorkspace(ctx, input.CourseID, input.ModuleID, input.ExerciseID)
		if err != nil {
			return nil, exerciseError("get exercise workspace", err)
		}
		return &ExerciseWorkspaceResponse{Body: exerciseWorkspaceResource(workspace)}, nil
	})

	huma.Register[PutExerciseWorkspaceInput, ExerciseWorkspaceResponse](api, huma.Operation{
		OperationID: "putExerciseWorkspace",
		Method:      http.MethodPut,
		Path:        "/api/courses/{courseId}/modules/{moduleId}/exercises/{exerciseId}/workspace",
		Summary:     "Replace an exercise workspace",
		Tags:        []string{"learner"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}, func(ctx context.Context, input *PutExerciseWorkspaceInput) (*ExerciseWorkspaceResponse, error) {
		if service == nil {
			return nil, huma.Error500InternalServerError("learner service is unavailable")
		}
		workspace, err := service.SetExerciseWorkspace(ctx, input.CourseID, input.ModuleID, input.ExerciseID, input.Body.Code)
		if err != nil {
			return nil, exerciseError("put exercise workspace", err)
		}
		return &ExerciseWorkspaceResponse{Body: exerciseWorkspaceResource(workspace)}, nil
	})

	huma.Register[CreateExerciseAttemptInput, CreateExerciseAttemptResponse](api, huma.Operation{
		OperationID:   "createExerciseAttempt",
		Method:        http.MethodPost,
		Path:          "/api/courses/{courseId}/modules/{moduleId}/exercises/{exerciseId}/attempts",
		Summary:       "Create an exercise attempt",
		Tags:          []string{"learner"},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, func(ctx context.Context, input *CreateExerciseAttemptInput) (*CreateExerciseAttemptResponse, error) {
		if service == nil {
			return nil, huma.Error500InternalServerError("learner service is unavailable")
		}
		results := make([]learner.ExerciseTestResult, 0, len(input.Body.Results))
		for _, result := range input.Body.Results {
			results = append(results, learner.ExerciseTestResult{
				TestID: result.TestID, Status: result.Status, Message: result.Message, DurationMS: result.DurationMS,
			})
		}
		attempt, err := service.CreateExerciseAttempt(ctx, input.CourseID, input.ModuleID, input.ExerciseID, input.Body.CodeSnapshot, input.Body.DurationMS, results)
		if err != nil {
			return nil, exerciseError("create exercise attempt", err)
		}
		return &CreateExerciseAttemptResponse{
			Location: fmt.Sprintf("/api/courses/%s/modules/%s/exercises/%s/attempts/%d", input.CourseID, input.ModuleID, input.ExerciseID, attempt.ID),
			Body:     exerciseAttemptResource(attempt),
		}, nil
	})
}

func exerciseWorkspaceResource(workspace learner.ExerciseWorkspace) ExerciseWorkspaceResource {
	return ExerciseWorkspaceResource{
		CourseID: workspace.CourseID, ModuleID: workspace.ModuleID, ExerciseID: workspace.ExerciseID,
		Code: workspace.Code, UpdatedAt: workspace.UpdatedAt,
	}
}

func exerciseAttemptResource(attempt learner.ExerciseAttempt) ExerciseAttemptResource {
	results := make([]ExerciseAttemptTestResultResource, 0, len(attempt.Results))
	for _, result := range attempt.Results {
		results = append(results, ExerciseAttemptTestResultResource{
			TestID: result.TestID, Status: result.Status, Message: result.Message, DurationMS: result.DurationMS,
		})
	}
	return ExerciseAttemptResource{
		ID: attempt.ID, CourseID: attempt.CourseID, ModuleID: attempt.ModuleID, ExerciseID: attempt.ExerciseID,
		CreatedAt: attempt.CreatedAt, PassedCount: attempt.PassedCount, FailedCount: attempt.FailedCount,
		DurationMS: attempt.DurationMS, AllPassed: attempt.AllPassed, CodeSnapshot: attempt.CodeSnapshot, Results: results,
	}
}

func exerciseError(operation string, err error) error {
	if errors.Is(err, learner.ErrExerciseNotFound) {
		return huma.Error404NotFound(err.Error())
	}
	if errors.Is(err, learner.ErrInvalidExerciseAttempt) {
		return huma.Error422UnprocessableEntity("exercise attempt results do not match the authored tests")
	}
	log.Printf("%s: %v", operation, err)
	return huma.Error500InternalServerError("exercise learner state is unavailable")
}
