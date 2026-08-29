package auth

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/johncrowleydev/helix-academy/server/internal/database"
)

func TestAnonymousRequestHasNoPrincipalAndCreatesNoGuest(t *testing.T) {
	service, db := testService(t)
	request := httptest.NewRequest(http.MethodGet, "/api/courses", nil)
	response := httptest.NewRecorder()
	service.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		if _, ok := CurrentUser(request.Context()); ok {
			t.Fatal("anonymous request unexpectedly has a user")
		}
	})).ServeHTTP(response, request)

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected only the durable bootstrap owner, got %d users", count)
	}
	var sessionCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM auth_sessions`).Scan(&sessionCount); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if sessionCount != 0 {
		t.Fatalf("anonymous request created %d sessions", sessionCount)
	}
}

func TestProvisionSignInAuthenticateAndSignOut(t *testing.T) {
	service, db := testService(t)
	ctx := context.Background()
	if err := service.ProvisionBootstrap(ctx, BootstrapConfig{
		Username: " Owner@Example.Test ", Password: "correct horse battery staple", DisplayName: "Fonzy Owner",
	}); err != nil {
		t.Fatalf("provision bootstrap user: %v", err)
	}

	user, token, err := service.SignIn(ctx, "owner@example.test", "correct horse battery staple")
	if err != nil {
		t.Fatalf("sign in: %v", err)
	}
	if user.ID != BootstrapUserID || user.DisplayName != "Fonzy Owner" || token == "" {
		t.Fatalf("unexpected authenticated identity: %#v token=%q", user, token)
	}

	var storedToken string
	if err := db.QueryRow(`SELECT token_hash FROM auth_sessions`).Scan(&storedToken); err != nil {
		t.Fatalf("load stored session: %v", err)
	}
	if storedToken == token || len(storedToken) != 64 {
		t.Fatalf("expected only a SHA-256 token hash in storage, got %q", storedToken)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	request.AddCookie(service.SessionCookie(token))
	service.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		current, ok := CurrentUser(request.Context())
		if !ok || current.ID != BootstrapUserID {
			t.Fatalf("session did not survive the next request: %#v, %v", current, ok)
		}
		if err := service.SignOut(request.Context()); err != nil {
			t.Fatalf("sign out: %v", err)
		}
	})).ServeHTTP(httptest.NewRecorder(), request)

	if _, _, ok, err := service.Authenticate(ctx, token); err != nil || ok {
		t.Fatalf("signed-out token remained valid: ok=%v err=%v", ok, err)
	}
}

func TestInvalidAndExpiredSessionsAreRejected(t *testing.T) {
	service, _ := testService(t)
	ctx := context.Background()
	if err := service.ProvisionBootstrap(ctx, BootstrapConfig{
		Username: "owner", Password: "correct horse battery staple",
	}); err != nil {
		t.Fatalf("provision bootstrap user: %v", err)
	}
	if _, _, err := service.SignIn(ctx, "owner", "wrong password"); err != ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}

	now := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	_, token, err := service.SignIn(ctx, "owner", "correct horse battery staple")
	if err != nil {
		t.Fatalf("sign in: %v", err)
	}
	now = now.Add(service.ttl + time.Second)
	if _, _, ok, err := service.Authenticate(ctx, token); err != nil || ok {
		t.Fatalf("expired session was accepted: ok=%v err=%v", ok, err)
	}
	if _, err := RequireUser(ctx); err != ErrUnauthenticated {
		t.Fatalf("expected unauthenticated error, got %v", err)
	}
}

func TestProvisionRequiresCompleteStrongCredentials(t *testing.T) {
	service, _ := testService(t)
	for name, config := range map[string]BootstrapConfig{
		"username only":  {Username: "owner"},
		"password only":  {Password: "correct horse battery staple"},
		"short password": {Username: "owner", Password: "too-short"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := service.ProvisionBootstrap(context.Background(), config); err == nil {
				t.Fatal("expected invalid bootstrap configuration to fail")
			}
		})
	}
}

func TestSessionCookieUsesSecurityProperties(t *testing.T) {
	service, _ := testService(t)
	service.secure = true
	cookie := service.SessionCookie("opaque-token")
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode || cookie.Path != "/" {
		t.Fatalf("session cookie is missing security properties: %#v", cookie)
	}
}

func TestCredentialRotationRevokesExistingSessions(t *testing.T) {
	service, _ := testService(t)
	ctx := context.Background()
	if err := service.ProvisionBootstrap(ctx, BootstrapConfig{Username: "owner", Password: "first secure password"}); err != nil {
		t.Fatalf("provision initial credentials: %v", err)
	}
	_, token, err := service.SignIn(ctx, "owner", "first secure password")
	if err != nil {
		t.Fatalf("sign in: %v", err)
	}
	if err := service.ProvisionBootstrap(ctx, BootstrapConfig{Username: "owner", Password: "second secure password"}); err != nil {
		t.Fatalf("rotate credentials: %v", err)
	}
	if _, _, ok, err := service.Authenticate(ctx, token); err != nil || ok {
		t.Fatalf("rotated credentials left an old session valid: ok=%v err=%v", ok, err)
	}
}

func TestRemovingBootstrapCredentialsDisablesAuthentication(t *testing.T) {
	service, db := testService(t)
	ctx := context.Background()
	const password = "correct horse battery staple"
	if err := service.ProvisionBootstrap(ctx, BootstrapConfig{
		Username: "owner", Password: password, DisplayName: "Fonzy Owner",
	}); err != nil {
		t.Fatalf("provision bootstrap credentials: %v", err)
	}
	user, token, err := service.SignIn(ctx, "owner", password)
	if err != nil {
		t.Fatalf("sign in: %v", err)
	}
	if _, _, ok, err := service.Authenticate(ctx, token); err != nil || !ok {
		t.Fatalf("issued session did not authenticate: ok=%v err=%v", ok, err)
	}

	if err := service.ProvisionBootstrap(ctx, BootstrapConfig{}); err != nil {
		t.Fatalf("remove bootstrap credentials: %v", err)
	}
	if _, _, err := service.SignIn(ctx, "owner", password); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("removed credentials still authenticated: %v", err)
	}
	if _, _, ok, err := service.Authenticate(ctx, token); err != nil || ok {
		t.Fatalf("removed credentials left old session valid: ok=%v err=%v", ok, err)
	}

	var storedID UserID
	var username sql.NullString
	var passwordHash []byte
	var displayName string
	if err := db.QueryRow(`
		SELECT id, username, password_hash, display_name
		FROM users
		WHERE id = ?
	`, BootstrapUserID).Scan(&storedID, &username, &passwordHash, &displayName); err != nil {
		t.Fatalf("load deprovisioned bootstrap user: %v", err)
	}
	if storedID != user.ID || storedID != BootstrapUserID {
		t.Fatalf("bootstrap user ID changed: got %q want %q", storedID, user.ID)
	}
	if username.Valid || passwordHash != nil {
		t.Fatalf("bootstrap credentials were not cleared: username=%#v hash=%x", username, passwordHash)
	}
	if displayName != "Fonzy Owner" {
		t.Fatalf("bootstrap display name changed: got %q", displayName)
	}

	if err := service.ProvisionBootstrap(ctx, BootstrapConfig{}); err != nil {
		t.Fatalf("repeat bootstrap deprovisioning: %v", err)
	}
}

func testService(t *testing.T) (*Service, *sql.DB) {
	t.Helper()
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	service := NewService(db, SessionConfig{TTL: time.Hour})
	service.random = bytes.NewReader(bytes.Repeat([]byte{0x42}, 32))
	return service, db
}
