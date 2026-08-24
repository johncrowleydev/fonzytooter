package httpapi

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/johncrowleydev/fonzytooter/server/internal/learner"
)

type LessonProgressUpdate struct {
	Completed bool `json:"completed"`
}

type PutLessonProgressInput struct {
	CourseID string `path:"courseId"`
	ModuleID string `path:"moduleId"`
	LessonID string `path:"lessonId"`
	Body     LessonProgressUpdate
}

type LessonProgressResource struct {
	CourseID    string     `json:"courseId"`
	ModuleID    string     `json:"moduleId"`
	LessonID    string     `json:"lessonId"`
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

type LessonProgressResponse struct {
	Body LessonProgressResource
}

type VideoProgressUpdate struct {
	Completed bool `json:"completed"`
}

type PutVideoProgressInput struct {
	CourseID string `path:"courseId"`
	ModuleID string `path:"moduleId"`
	VideoID  string `path:"videoId"`
	Body     VideoProgressUpdate
}

type CourseVideoPathInput struct {
	CourseID string `path:"courseId"`
	ModuleID string `path:"moduleId"`
	VideoID  string `path:"videoId"`
}

type VideoProgressResource struct {
	CourseID    string     `json:"courseId"`
	ModuleID    string     `json:"moduleId"`
	VideoID     string     `json:"videoId"`
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

type VideoProgressResponse struct {
	Body VideoProgressResource
}

type ObjectiveProgressResource struct {
	CourseID             string                      `json:"courseId"`
	ModuleID             string                      `json:"moduleId"`
	ID                   string                      `json:"id"`
	Title                string                      `json:"title"`
	Description          string                      `json:"description"`
	Introduced           bool                        `json:"introduced"`
	LinkedLessonCount    int                         `json:"linkedLessonCount"`
	CompletedLessonCount int                         `json:"completedLessonCount"`
	Recall               RecallEvidenceResource      `json:"recall"`
	Application          ApplicationEvidenceResource `json:"application"`
	TransferAssessed     bool                        `json:"transferAssessed"`
}

type RecallEvidenceResource struct {
	ReviewItemCount  int        `json:"reviewItemCount"`
	ReviewsCompleted int        `json:"reviewsCompleted"`
	DueReviewCount   int        `json:"dueReviewCount"`
	LastReviewedAt   *time.Time `json:"lastReviewedAt,omitempty"`
	NextDueAt        *time.Time `json:"nextDueAt,omitempty"`
}

type ApplicationEvidenceResource struct {
	ExerciseCount       int        `json:"exerciseCount"`
	Attempts            int        `json:"attempts"`
	FullyPassedAttempts int        `json:"fullyPassedAttempts"`
	LastCheckedAt       *time.Time `json:"lastCheckedAt,omitempty"`
}

type NextLessonResource struct {
	CourseID    string `json:"courseId"`
	ModuleID    string `json:"moduleId"`
	ModuleTitle string `json:"moduleTitle"`
	LessonID    string `json:"lessonId"`
	LessonTitle string `json:"lessonTitle"`
}

type PracticeExerciseResource struct {
	CourseID      string `json:"courseId"`
	ModuleID      string `json:"moduleId"`
	ModuleTitle   string `json:"moduleTitle"`
	ExerciseID    string `json:"exerciseId"`
	ExerciseTitle string `json:"exerciseTitle"`
}

type CourseProgressResource struct {
	CourseID             string                      `json:"courseId"`
	CompletedLessonCount int                         `json:"completedLessonCount"`
	TotalLessonCount     int                         `json:"totalLessonCount"`
	DueReviewCount       int                         `json:"dueReviewCount"`
	Objectives           []ObjectiveProgressResource `json:"objectives" nullable:"false"`
	NextLesson           *NextLessonResource         `json:"nextLesson,omitempty"`
	PracticeExercise     *PracticeExerciseResource   `json:"practiceExercise,omitempty"`
}

type CourseProgressResponse struct {
	Body CourseProgressResource
}

type ActivityQueryInput struct {
	CourseID string `query:"courseId" required:"true"`
	Limit    int    `query:"limit" default:"20" minimum:"1" maximum:"100"`
}

type ActivityResource struct {
	ID            int64     `json:"id"`
	Kind          string    `json:"kind"`
	CourseID      string    `json:"courseId"`
	CourseTitle   string    `json:"courseTitle"`
	ModuleID      *string   `json:"moduleId,omitempty"`
	ModuleTitle   *string   `json:"moduleTitle,omitempty"`
	LessonID      *string   `json:"lessonId,omitempty"`
	LessonTitle   *string   `json:"lessonTitle,omitempty"`
	ExerciseID    *string   `json:"exerciseId,omitempty"`
	ExerciseTitle *string   `json:"exerciseTitle,omitempty"`
	VideoID       *string   `json:"videoId,omitempty"`
	VideoTitle    *string   `json:"videoTitle,omitempty"`
	ReviewItemID  *string   `json:"reviewItemId,omitempty"`
	OccurredAt    time.Time `json:"occurredAt"`
}

type ActivityListResponse struct {
	Body []ActivityResource
}

func registerLearning(api huma.API, service *learner.Service) {
	huma.Register[CourseVideoPathInput, VideoProgressResponse](api, authenticatedOperation(huma.Operation{
		OperationID: "getVideoProgress",
		Method:      http.MethodGet,
		Path:        "/api/courses/{courseId}/modules/{moduleId}/videos/{videoId}/progress",
		Summary:     "Get video progress",
		Tags:        []string{"learner"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}), func(ctx context.Context, input *CourseVideoPathInput) (*VideoProgressResponse, error) {
		if service == nil {
			return nil, huma.Error500InternalServerError("learner service is unavailable")
		}
		userID, err := requireUserID(ctx)
		if err != nil {
			return nil, err
		}
		progress, err := service.VideoProgress(ctx, userID, input.CourseID, input.ModuleID, input.VideoID)
		if err != nil {
			return nil, learningError("get video progress", err)
		}
		return &VideoProgressResponse{Body: videoProgressResource(progress)}, nil
	})

	huma.Register[PutVideoProgressInput, VideoProgressResponse](api, authenticatedOperation(huma.Operation{
		OperationID: "putVideoProgress",
		Method:      http.MethodPut,
		Path:        "/api/courses/{courseId}/modules/{moduleId}/videos/{videoId}/progress",
		Summary:     "Replace video progress",
		Tags:        []string{"learner"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}), func(ctx context.Context, input *PutVideoProgressInput) (*VideoProgressResponse, error) {
		if service == nil {
			return nil, huma.Error500InternalServerError("learner service is unavailable")
		}
		userID, err := requireUserID(ctx)
		if err != nil {
			return nil, err
		}
		progress, err := service.SetVideoProgress(ctx, userID, input.CourseID, input.ModuleID, input.VideoID, input.Body.Completed)
		if err != nil {
			return nil, learningError("put video progress", err)
		}
		return &VideoProgressResponse{Body: videoProgressResource(progress)}, nil
	})

	huma.Register[CourseLessonPathInput, LessonProgressResponse](api, authenticatedOperation(huma.Operation{
		OperationID: "getLessonProgress",
		Method:      http.MethodGet,
		Path:        "/api/courses/{courseId}/modules/{moduleId}/lessons/{lessonId}/progress",
		Summary:     "Get lesson progress",
		Tags:        []string{"learner"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}), func(ctx context.Context, input *CourseLessonPathInput) (*LessonProgressResponse, error) {
		if service == nil {
			return nil, huma.Error500InternalServerError("learner service is unavailable")
		}
		userID, err := requireUserID(ctx)
		if err != nil {
			return nil, err
		}
		progress, err := service.LessonProgress(ctx, userID, input.CourseID, input.ModuleID, input.LessonID)
		if err != nil {
			return nil, learningError("get lesson progress", err)
		}
		return &LessonProgressResponse{Body: lessonProgressResource(progress)}, nil
	})

	huma.Register[PutLessonProgressInput, LessonProgressResponse](api, authenticatedOperation(huma.Operation{
		OperationID: "putLessonProgress",
		Method:      http.MethodPut,
		Path:        "/api/courses/{courseId}/modules/{moduleId}/lessons/{lessonId}/progress",
		Summary:     "Replace lesson progress",
		Tags:        []string{"learner"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}), func(ctx context.Context, input *PutLessonProgressInput) (*LessonProgressResponse, error) {
		if service == nil {
			return nil, huma.Error500InternalServerError("learner service is unavailable")
		}
		userID, err := requireUserID(ctx)
		if err != nil {
			return nil, err
		}
		progress, err := service.SetLessonProgress(ctx, userID, input.CourseID, input.ModuleID, input.LessonID, input.Body.Completed)
		if err != nil {
			return nil, learningError("put lesson progress", err)
		}
		return &LessonProgressResponse{Body: lessonProgressResource(progress)}, nil
	})

	huma.Register[CoursePathInput, CourseProgressResponse](api, authenticatedOperation(huma.Operation{
		OperationID: "getCourseProgress",
		Method:      http.MethodGet,
		Path:        "/api/courses/{courseId}/progress",
		Summary:     "Get course progress",
		Tags:        []string{"learner"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}), func(ctx context.Context, input *CoursePathInput) (*CourseProgressResponse, error) {
		if service == nil {
			return nil, huma.Error500InternalServerError("learner service is unavailable")
		}
		userID, err := requireUserID(ctx)
		if err != nil {
			return nil, err
		}
		progress, err := service.CourseProgress(ctx, userID, input.CourseID)
		if err != nil {
			return nil, learningError("get course progress", err)
		}
		objectives := make([]ObjectiveProgressResource, 0, len(progress.Objectives))
		for _, objective := range progress.Objectives {
			objectives = append(objectives, ObjectiveProgressResource{
				CourseID: objective.CourseID, ModuleID: objective.ModuleID, ID: objective.ID,
				Title: objective.Title, Description: objective.Description,
				Introduced: objective.Introduced, LinkedLessonCount: objective.LinkedLessonCount,
				CompletedLessonCount: objective.CompletedLessonCount, TransferAssessed: objective.TransferAssessed,
				Recall: RecallEvidenceResource{
					ReviewItemCount: objective.Recall.ReviewItemCount, ReviewsCompleted: objective.Recall.ReviewsCompleted,
					DueReviewCount: objective.Recall.DueReviewCount, LastReviewedAt: objective.Recall.LastReviewedAt,
					NextDueAt: objective.Recall.NextDueAt,
				},
				Application: ApplicationEvidenceResource{
					ExerciseCount: objective.Application.ExerciseCount, Attempts: objective.Application.Attempts,
					FullyPassedAttempts: objective.Application.FullyPassedAttempts, LastCheckedAt: objective.Application.LastCheckedAt,
				},
			})
		}
		var nextLesson *NextLessonResource
		if progress.NextLesson != nil {
			nextLesson = &NextLessonResource{
				CourseID: progress.NextLesson.CourseID, ModuleID: progress.NextLesson.ModuleID,
				ModuleTitle: progress.NextLesson.ModuleTitle, LessonID: progress.NextLesson.LessonID,
				LessonTitle: progress.NextLesson.LessonTitle,
			}
		}
		var practiceExercise *PracticeExerciseResource
		if progress.PracticeExercise != nil {
			practiceExercise = &PracticeExerciseResource{
				CourseID: progress.PracticeExercise.CourseID, ModuleID: progress.PracticeExercise.ModuleID,
				ModuleTitle: progress.PracticeExercise.ModuleTitle, ExerciseID: progress.PracticeExercise.ExerciseID,
				ExerciseTitle: progress.PracticeExercise.ExerciseTitle,
			}
		}
		return &CourseProgressResponse{Body: CourseProgressResource{
			CourseID:             progress.CourseID,
			CompletedLessonCount: progress.CompletedLessonCount, TotalLessonCount: progress.TotalLessonCount,
			DueReviewCount: progress.DueReviewCount, Objectives: objectives,
			NextLesson: nextLesson, PracticeExercise: practiceExercise,
		}}, nil
	})

	huma.Register[ActivityQueryInput, ActivityListResponse](api, authenticatedOperation(huma.Operation{
		OperationID: "listActivities",
		Method:      http.MethodGet,
		Path:        "/api/activities",
		Summary:     "List recent learner activity",
		Tags:        []string{"learner"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}), func(ctx context.Context, input *ActivityQueryInput) (*ActivityListResponse, error) {
		if service == nil {
			return nil, huma.Error500InternalServerError("learner service is unavailable")
		}
		userID, err := requireUserID(ctx)
		if err != nil {
			return nil, err
		}
		activities, err := service.Activities(ctx, userID, input.CourseID, input.Limit)
		if err != nil {
			return nil, learningError("list learner activities", err)
		}
		body := make([]ActivityResource, 0, len(activities))
		for _, activity := range activities {
			body = append(body, ActivityResource{
				ID: activity.ID, Kind: activity.Kind, CourseID: activity.CourseID,
				CourseTitle: activity.CourseTitle, ModuleID: activity.ModuleID,
				ModuleTitle: activity.ModuleTitle, LessonID: activity.LessonID,
				LessonTitle: activity.LessonTitle, ExerciseID: activity.ExerciseID,
				ExerciseTitle: activity.ExerciseTitle, ReviewItemID: activity.ReviewItemID,
				VideoID: activity.VideoID, VideoTitle: activity.VideoTitle,
				OccurredAt: activity.OccurredAt,
			})
		}
		return &ActivityListResponse{Body: body}, nil
	})
	api.OpenAPI().Components.Schemas.Map()["ActivityList"] = &huma.Schema{
		Type:  huma.TypeArray,
		Items: &huma.Schema{Ref: "#/components/schemas/ActivityResource"},
	}
	api.OpenAPI().Paths["/api/activities"].Get.Responses["200"].Content["application/json"].Schema = &huma.Schema{
		Ref: "#/components/schemas/ActivityList",
	}
}

func lessonProgressResource(progress learner.LessonProgress) LessonProgressResource {
	return LessonProgressResource{
		CourseID: progress.CourseID, ModuleID: progress.ModuleID, LessonID: progress.LessonID,
		Completed: progress.Completed, CompletedAt: progress.CompletedAt,
	}
}

func videoProgressResource(progress learner.VideoProgress) VideoProgressResource {
	return VideoProgressResource{
		CourseID: progress.CourseID, ModuleID: progress.ModuleID, VideoID: progress.VideoID,
		Completed: progress.Completed, CompletedAt: progress.CompletedAt,
	}
}

func learningError(operation string, err error) error {
	if errors.Is(err, learner.ErrCourseNotFound) || errors.Is(err, learner.ErrLessonNotFound) || errors.Is(err, learner.ErrVideoNotFound) {
		return huma.Error404NotFound(err.Error())
	}
	log.Printf("%s: %v", operation, err)
	return huma.Error500InternalServerError("learner state is unavailable")
}
