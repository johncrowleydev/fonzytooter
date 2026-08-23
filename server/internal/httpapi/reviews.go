package httpapi

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/johncrowleydev/fonzytooter/server/internal/review"
)

type ReviewCardQueryInput struct {
	CourseID string `path:"courseId"`
	Due      bool   `query:"due" default:"true"`
}

type ReviewCardReviewPathInput struct {
	CourseID     string `path:"courseId"`
	ModuleID     string `path:"moduleId"`
	ReviewItemID string `path:"reviewItemId"`
	Body         ReviewSubmission
}

type ReviewSubmission struct {
	Rating review.Rating `json:"rating" enum:"again,hard,good,easy"`
}

type RatingPreviewResource struct {
	Rating          review.Rating `json:"rating" enum:"again,hard,good,easy"`
	DueAt           time.Time     `json:"dueAt"`
	IntervalSeconds int64         `json:"intervalSeconds" minimum:"0"`
	IntervalDays    float64       `json:"intervalDays" minimum:"0"`
}

type ReviewCardResource struct {
	CourseID       string                  `json:"courseId"`
	ModuleID       string                  `json:"moduleId"`
	ID             string                  `json:"id"`
	Order          int                     `json:"order"`
	ObjectiveIDs   []string                `json:"objectiveIds" nullable:"false"`
	SourceLessonID string                  `json:"sourceLessonId"`
	Prompt         string                  `json:"prompt"`
	Answer         string                  `json:"answer"`
	Hint           string                  `json:"hint"`
	State          string                  `json:"state" enum:"new,learning,review,relearning"`
	DueAt          time.Time               `json:"dueAt"`
	LastReviewedAt *time.Time              `json:"lastReviewedAt,omitempty"`
	Virtual        bool                    `json:"virtual"`
	Due            bool                    `json:"due"`
	Previews       []RatingPreviewResource `json:"previews" nullable:"false"`
}

type ReviewCardListResponse struct {
	Body []ReviewCardResource
}

type CreateCardReviewResponse struct {
	Location string `header:"Location"`
	Body     ReviewCardResource
}

func registerReviews(api huma.API, service *review.Service) {
	huma.Register[ReviewCardQueryInput, ReviewCardListResponse](api, authenticatedOperation(huma.Operation{
		OperationID: "listReviewCards",
		Method:      http.MethodGet,
		Path:        "/api/courses/{courseId}/review-cards",
		Summary:     "List learner review cards",
		Description: "Combines authored review items with persistent or virtual FSRS scheduling state.",
		Tags:        []string{"learner"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}), func(ctx context.Context, input *ReviewCardQueryInput) (*ReviewCardListResponse, error) {
		if service == nil {
			return nil, huma.Error500InternalServerError("review service is unavailable")
		}
		userID, err := requireUserID(ctx)
		if err != nil {
			return nil, err
		}
		cards, err := service.Cards(ctx, userID, input.CourseID, input.Due)
		if err != nil {
			return nil, reviewError("list review cards", err)
		}
		body := make([]ReviewCardResource, 0, len(cards))
		for _, card := range cards {
			body = append(body, reviewCardResource(card))
		}
		return &ReviewCardListResponse{Body: body}, nil
	})
	api.OpenAPI().Components.Schemas.Map()["ReviewCardList"] = &huma.Schema{
		Type:  huma.TypeArray,
		Items: &huma.Schema{Ref: "#/components/schemas/ReviewCardResource"},
	}
	api.OpenAPI().Paths["/api/courses/{courseId}/review-cards"].Get.Responses["200"].Content["application/json"].Schema = &huma.Schema{
		Ref: "#/components/schemas/ReviewCardList",
	}

	huma.Register[ReviewCardReviewPathInput, CreateCardReviewResponse](api, authenticatedOperation(huma.Operation{
		OperationID:   "createReviewCardReview",
		Method:        http.MethodPost,
		Path:          "/api/courses/{courseId}/modules/{moduleId}/review-cards/{reviewItemId}/reviews",
		Summary:       "Create a review-card rating",
		Description:   "Applies one rating and atomically records the updated FSRS card, immutable log, and learner activity.",
		Tags:          []string{"learner"},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}), func(ctx context.Context, input *ReviewCardReviewPathInput) (*CreateCardReviewResponse, error) {
		if service == nil {
			return nil, huma.Error500InternalServerError("review service is unavailable")
		}
		userID, err := requireUserID(ctx)
		if err != nil {
			return nil, err
		}
		result, err := service.Submit(ctx, userID, input.CourseID, input.ModuleID, input.ReviewItemID, input.Body.Rating)
		if err != nil {
			return nil, reviewError("create review", err)
		}
		return &CreateCardReviewResponse{
			Location: fmt.Sprintf("/api/courses/%s/modules/%s/review-cards/%s/reviews/%d", input.CourseID, input.ModuleID, input.ReviewItemID, result.ID),
			Body:     reviewCardResource(result.Card),
		}, nil
	})
}

func reviewCardResource(card review.Card) ReviewCardResource {
	previews := make([]RatingPreviewResource, 0, len(card.Previews))
	for _, preview := range card.Previews {
		previews = append(previews, RatingPreviewResource{
			Rating: preview.Rating, DueAt: preview.DueAt,
			IntervalSeconds: preview.IntervalSeconds, IntervalDays: preview.IntervalDays,
		})
	}
	return ReviewCardResource{
		CourseID: card.Item.CourseID, ModuleID: card.Item.ModuleID, ID: card.Item.ID,
		Order: card.Item.Order, ObjectiveIDs: append([]string{}, card.Item.ObjectiveIDs...),
		SourceLessonID: card.Item.SourceLessonID, Prompt: card.Item.Prompt,
		Answer: card.Item.Answer, Hint: card.Item.Hint, State: strings.ToLower(card.Schedule.State.String()),
		DueAt: card.Schedule.Due.UTC(), LastReviewedAt: card.LastRated,
		Virtual: card.Virtual, Due: card.IsDue, Previews: previews,
	}
}

func reviewError(operation string, err error) error {
	switch {
	case errors.Is(err, review.ErrCourseNotFound), errors.Is(err, review.ErrReviewItemNotFound):
		return huma.Error404NotFound(err.Error())
	case errors.Is(err, review.ErrInvalidRating):
		return huma.Error422UnprocessableEntity(err.Error())
	case errors.Is(err, review.ErrReviewItemNotEligible):
		return huma.Error409Conflict(err.Error())
	default:
		log.Printf("%s: %v", operation, err)
		return huma.Error500InternalServerError("review scheduling is unavailable")
	}
}
