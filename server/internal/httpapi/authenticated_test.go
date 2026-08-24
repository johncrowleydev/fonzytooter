package httpapi

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/johncrowleydev/fonzytooter/server/internal/auth"
	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/database"
	"github.com/johncrowleydev/fonzytooter/server/internal/learner"
	"github.com/johncrowleydev/fonzytooter/server/internal/review"
	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
)

func newAuthenticatedTestAPI(t *testing.T, tutorService *tutor.Service, catalog *curriculum.Catalog, learnerService *learner.Service, reviewService *review.Service) *API {
	t.Helper()
	app, cookie := newTestAPIWithSession(t, tutorService, catalog, learnerService, reviewService)
	underlying := app.Handler
	app.Handler = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		request.AddCookie(cookie)
		underlying.ServeHTTP(response, request)
	})
	return app
}

func newTestAPIWithSession(t *testing.T, tutorService *tutor.Service, catalog *curriculum.Catalog, learnerService *learner.Service, reviewService *review.Service) (*API, *http.Cookie) {
	t.Helper()
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatalf("open authentication database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	authService := auth.NewService(db, auth.SessionConfig{TTL: time.Hour})
	if err := authService.ProvisionBootstrap(context.Background(), auth.BootstrapConfig{
		Username: "owner", Password: "test-password", DisplayName: "Owner",
	}); err != nil {
		t.Fatalf("provision test owner: %v", err)
	}
	_, token, err := authService.SignIn(context.Background(), "owner", "test-password")
	if err != nil {
		t.Fatalf("sign in test owner: %v", err)
	}
	app := NewAPIWithAuth(tutorService, catalog, learnerService, reviewService, authService)
	return app, authService.SessionCookie(token)
}
