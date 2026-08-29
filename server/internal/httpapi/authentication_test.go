package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/johncrowleydev/helix-academy/server/internal/auth"
	"github.com/johncrowleydev/helix-academy/server/internal/curriculum"
	"github.com/johncrowleydev/helix-academy/server/internal/database"
	"github.com/johncrowleydev/helix-academy/server/internal/tutor"
)

func TestAuthenticationSessionLifecycle(t *testing.T) {
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	authService := auth.NewService(db, auth.SessionConfig{TTL: time.Hour})
	if err := authService.ProvisionBootstrap(context.Background(), auth.BootstrapConfig{
		Username: "owner", Password: "correct horse battery staple", DisplayName: "Owner",
	}); err != nil {
		t.Fatalf("provision user: %v", err)
	}
	app := NewAPIWithAuth(tutor.NewService(tutor.NewUnavailableProvider()), curriculum.NewEmptyCatalog(), nil, nil, authService)

	publicCurriculum := requestJSON(t, app.Handler, http.MethodGet, "/api/courses", "", nil)
	if publicCurriculum.Code != http.StatusOK || strings.TrimSpace(publicCurriculum.Body.String()) != "[]" {
		t.Fatalf("public curriculum was gated: %d %s", publicCurriculum.Code, publicCurriculum.Body.String())
	}
	var sessionCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM auth_sessions`).Scan(&sessionCount); err != nil || sessionCount != 0 {
		t.Fatalf("anonymous curriculum read created session state: count=%d err=%v", sessionCount, err)
	}

	anonymous := requestJSON(t, app.Handler, http.MethodGet, "/api/authentication-sessions/current", "", nil)
	if anonymous.Code != http.StatusOK {
		t.Fatalf("anonymous session returned %d: %s", anonymous.Code, anonymous.Body.String())
	}
	if anonymous.Header().Get("Cache-Control") != "no-store" || anonymous.Header().Get("Vary") != "Cookie" {
		t.Fatalf("current session response is cacheable: %#v", anonymous.Header())
	}
	assertSessionResponse(t, anonymous, false, "")

	invalid := requestJSON(t, app.Handler, http.MethodPost, "/api/authentication-sessions", `{"username":"owner","password":"wrong"}`, nil)
	if invalid.Code != http.StatusUnauthorized || invalid.Header().Get("Set-Cookie") != "" {
		t.Fatalf("invalid sign-in returned code=%d cookie=%q body=%s", invalid.Code, invalid.Header().Get("Set-Cookie"), invalid.Body.String())
	}

	signedIn := requestJSON(t, app.Handler, http.MethodPost, "/api/authentication-sessions", `{"username":"owner","password":"correct horse battery staple"}`, nil)
	if signedIn.Code != http.StatusCreated {
		t.Fatalf("sign-in returned %d: %s", signedIn.Code, signedIn.Body.String())
	}
	if signedIn.Header().Get("Location") != "/api/authentication-sessions/current" {
		t.Fatalf("sign-in location = %q", signedIn.Header().Get("Location"))
	}
	cookies := signedIn.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != auth.DefaultCookieName || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode {
		t.Fatalf("unexpected session cookies: %#v", cookies)
	}
	assertSessionResponse(t, signedIn, true, string(auth.BootstrapUserID))

	current := requestJSON(t, app.Handler, http.MethodGet, "/api/authentication-sessions/current", "", cookies[0])
	assertSessionResponse(t, current, true, string(auth.BootstrapUserID))

	signedOut := requestJSON(t, app.Handler, http.MethodDelete, "/api/authentication-sessions/current", "", cookies[0])
	if signedOut.Code != http.StatusNoContent {
		t.Fatalf("sign-out returned %d: %s", signedOut.Code, signedOut.Body.String())
	}
	if got := signedOut.Result().Cookies(); len(got) != 1 || got[0].MaxAge >= 0 {
		t.Fatalf("sign-out did not expire the session cookie: %#v", got)
	}
	afterSignOut := requestJSON(t, app.Handler, http.MethodGet, "/api/authentication-sessions/current", "", cookies[0])
	assertSessionResponse(t, afterSignOut, false, "")
}

func TestAuthenticationOpenAPIContract(t *testing.T) {
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), curriculum.NewEmptyCatalog(), nil, nil)
	collection := app.Spec.Paths["/api/authentication-sessions"]
	current := app.Spec.Paths["/api/authentication-sessions/current"]
	if collection == nil || collection.Post == nil || current == nil || current.Get == nil || current.Delete == nil {
		t.Fatalf("session resource operations are incomplete: collection=%#v current=%#v", collection, current)
	}
	if collection.Post.OperationID != "createAuthenticationSession" || current.Get.OperationID != "getCurrentAuthenticationSession" || current.Delete.OperationID != "deleteCurrentAuthenticationSession" {
		t.Fatalf("session operation IDs are unstable: collection=%#v current=%#v", collection, current)
	}
	if collection.Post.Responses["201"] == nil || collection.Post.Responses["201"].Headers["Location"] == nil {
		t.Fatalf("session creation response is missing 201/Location: %#v", collection.Post.Responses)
	}
	scheme := app.Spec.Components.SecuritySchemes[sessionSecurityScheme]
	if scheme == nil || scheme.Type != "apiKey" || scheme.In != "cookie" || scheme.Name != auth.DefaultCookieName {
		t.Fatalf("session cookie security scheme is missing: %#v", scheme)
	}
}

func requestJSON(t *testing.T, handler http.Handler, method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func assertSessionResponse(t *testing.T, response *httptest.ResponseRecorder, authenticated bool, userID string) {
	t.Helper()
	var session SessionResource
	if err := json.NewDecoder(response.Body).Decode(&session); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if session.Authenticated != authenticated {
		t.Fatalf("expected authenticated=%v, got %#v", authenticated, session)
	}
	if !authenticated && session.User != nil {
		t.Fatalf("anonymous session exposed a user: %#v", session.User)
	}
	if authenticated && (session.User == nil || session.User.ID != userID) {
		t.Fatalf("authenticated session missing expected user %q: %#v", userID, session.User)
	}
}
