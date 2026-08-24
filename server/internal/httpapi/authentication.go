package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/johncrowleydev/fonzytooter/server/internal/auth"
)

const sessionSecurityScheme = "sessionCookie"

type UserResource struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

type SessionResource struct {
	Authenticated bool          `json:"authenticated"`
	User          *UserResource `json:"user,omitempty"`
}

type CreateSessionRequest struct {
	Username string `json:"username" minLength:"1"`
	Password string `json:"password" minLength:"1"`
}

type CreateSessionInput struct {
	Body CreateSessionRequest
}

type GetSessionResponse struct {
	CacheControl string `header:"Cache-Control"`
	Vary         string `header:"Vary"`
	Body         SessionResource
}

type CreateSessionResponse struct {
	Status       int
	SetCookie    string `header:"Set-Cookie"`
	Location     string `header:"Location"`
	CacheControl string `header:"Cache-Control"`
	Body         SessionResource
}

type DeleteSessionResponse struct {
	Status       int
	SetCookie    string `header:"Set-Cookie"`
	CacheControl string `header:"Cache-Control"`
}

func registerAuthentication(api huma.API, service *auth.Service) {
	huma.Register[struct{}, GetSessionResponse](api, publicOperation(huma.Operation{
		OperationID: "getCurrentAuthenticationSession",
		Method:      http.MethodGet,
		Path:        "/api/authentication-sessions/current",
		Summary:     "Get the current authentication session",
		Tags:        []string{"authentication"},
	}), func(ctx context.Context, _ *struct{}) (*GetSessionResponse, error) {
		user, ok := auth.CurrentUser(ctx)
		if !ok {
			return &GetSessionResponse{CacheControl: "no-store", Vary: "Cookie", Body: SessionResource{Authenticated: false}}, nil
		}
		return &GetSessionResponse{CacheControl: "no-store", Vary: "Cookie", Body: authenticatedSession(user)}, nil
	})

	huma.Register[CreateSessionInput, CreateSessionResponse](api, publicOperation(huma.Operation{
		OperationID:   "createAuthenticationSession",
		Method:        http.MethodPost,
		Path:          "/api/authentication-sessions",
		Summary:       "Sign in and create an authentication session",
		Tags:          []string{"authentication"},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusUnauthorized, http.StatusServiceUnavailable},
	}), func(ctx context.Context, input *CreateSessionInput) (*CreateSessionResponse, error) {
		if service == nil {
			return nil, huma.Error503ServiceUnavailable("authentication is unavailable")
		}
		user, token, err := service.SignIn(ctx, input.Body.Username, input.Body.Password)
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, huma.Error401Unauthorized("invalid username or password")
		}
		if err != nil {
			return nil, huma.Error500InternalServerError("authentication failed")
		}
		return &CreateSessionResponse{
			Status:       http.StatusCreated,
			SetCookie:    service.SessionCookie(token).String(),
			Location:     "/api/authentication-sessions/current",
			CacheControl: "no-store",
			Body:         authenticatedSession(user),
		}, nil
	})

	huma.Register[struct{}, DeleteSessionResponse](api, publicOperation(huma.Operation{
		OperationID: "deleteCurrentAuthenticationSession",
		Method:      http.MethodDelete,
		Path:        "/api/authentication-sessions/current",
		Summary:     "Sign out and delete the current authentication session",
		Tags:        []string{"authentication"},
		Errors:      []int{http.StatusServiceUnavailable},
	}), func(ctx context.Context, _ *struct{}) (*DeleteSessionResponse, error) {
		if service == nil {
			return nil, huma.Error503ServiceUnavailable("authentication is unavailable")
		}
		if err := service.SignOut(ctx); err != nil {
			return nil, huma.Error500InternalServerError("sign out failed")
		}
		return &DeleteSessionResponse{Status: http.StatusNoContent, SetCookie: service.ExpiredSessionCookie().String(), CacheControl: "no-store"}, nil
	})
}

func authenticatedSession(user auth.User) SessionResource {
	return SessionResource{
		Authenticated: true,
		User: &UserResource{
			ID:          string(user.ID),
			Username:    user.Username,
			DisplayName: user.DisplayName,
		},
	}
}

func requireUserID(ctx context.Context) (auth.UserID, error) {
	user, err := auth.RequireUser(ctx)
	if err != nil {
		return "", huma.Error401Unauthorized("authentication required")
	}
	return user.ID, nil
}
