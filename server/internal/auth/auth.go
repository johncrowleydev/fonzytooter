package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserID string

const (
	BootstrapUserID       UserID = "00000000-0000-4000-8000-000000000001"
	DefaultCookieName            = "fonzytooter_session"
	minimumPasswordLength        = 12
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthenticated    = errors.New("authentication required")
)

type User struct {
	ID          UserID
	Username    string
	DisplayName string
}

type BootstrapConfig struct {
	Username    string
	Password    string
	DisplayName string
}

type SessionConfig struct {
	CookieName string
	Secure     bool
	TTL        time.Duration
}

type Service struct {
	db         *sql.DB
	cookieName string
	secure     bool
	ttl        time.Duration
	now        func() time.Time
	random     io.Reader
}

type principal struct {
	user        User
	sessionHash string
}

type principalContextKey struct{}

func NewService(db *sql.DB, config SessionConfig) *Service {
	if db == nil {
		panic("auth.NewService: nil database")
	}
	if config.CookieName == "" {
		config.CookieName = DefaultCookieName
	}
	if config.TTL <= 0 {
		config.TTL = 24 * time.Hour
	}
	return &Service{
		db:         db,
		cookieName: config.CookieName,
		secure:     config.Secure,
		ttl:        config.TTL,
		now:        time.Now,
		random:     rand.Reader,
	}
}

// ProvisionBootstrap configures the personal deployment's durable owner. An
// empty configuration disables authentication while preserving the owner and
// public curriculum access.
func (s *Service) ProvisionBootstrap(ctx context.Context, config BootstrapConfig) error {
	username := normalizeUsername(config.Username)
	password := config.Password
	displayName := strings.TrimSpace(config.DisplayName)
	if username == "" && password == "" {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin bootstrap deprovisioning: %w", err)
		}
		defer func() { _ = tx.Rollback() }()
		if _, err := tx.ExecContext(ctx, `
			UPDATE users
			SET username = NULL, password_hash = NULL, updated_at = ?
			WHERE id = ?
		`, formatTime(s.now()), BootstrapUserID); err != nil {
			return fmt.Errorf("clear bootstrap credentials: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM auth_sessions WHERE user_id = ?`, BootstrapUserID); err != nil {
			return fmt.Errorf("revoke bootstrap sessions: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit bootstrap deprovisioning: %w", err)
		}
		return nil
	}
	if username == "" || password == "" {
		return errors.New("bootstrap username and password must be configured together")
	}
	if len(password) < minimumPasswordLength {
		return fmt.Errorf("bootstrap password must contain at least %d characters", minimumPasswordLength)
	}
	if displayName == "" {
		displayName = "Owner"
	}

	var currentUsername string
	var currentHash []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(username, ''), COALESCE(password_hash, X'')
		FROM users
		WHERE id = ?
	`, BootstrapUserID).Scan(&currentUsername, &currentHash)
	if err != nil {
		return fmt.Errorf("load bootstrap user: %w", err)
	}
	if currentUsername == username && bcrypt.CompareHashAndPassword(currentHash, []byte(password)) == nil {
		_, err = s.db.ExecContext(ctx, `UPDATE users SET display_name = ?, updated_at = ? WHERE id = ?`, displayName, formatTime(s.now()), BootstrapUserID)
		if err != nil {
			return fmt.Errorf("update bootstrap user display name: %w", err)
		}
		return nil
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash bootstrap password: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin bootstrap provisioning: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx, `
		UPDATE users
		SET username = ?, display_name = ?, password_hash = ?, updated_at = ?
		WHERE id = ?
	`, username, displayName, passwordHash, formatTime(s.now()), BootstrapUserID)
	if err != nil {
		return fmt.Errorf("provision bootstrap user: %w", err)
	}
	// Credential rotation is also a session revocation boundary.
	if _, err := tx.ExecContext(ctx, `DELETE FROM auth_sessions WHERE user_id = ?`, BootstrapUserID); err != nil {
		return fmt.Errorf("revoke bootstrap sessions: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit bootstrap provisioning: %w", err)
	}
	return nil
}

func (s *Service) SignIn(ctx context.Context, username, password string) (User, string, error) {
	var user User
	var passwordHash []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT id, username, display_name, password_hash
		FROM users
		WHERE username = ? AND password_hash IS NOT NULL
	`, normalizeUsername(username)).Scan(&user.ID, &user.Username, &user.DisplayName, &passwordHash)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && bcrypt.CompareHashAndPassword(passwordHash, []byte(password)) != nil) {
		return User{}, "", ErrInvalidCredentials
	}
	if err != nil {
		return User{}, "", fmt.Errorf("load authentication user: %w", err)
	}

	tokenBytes := make([]byte, 32)
	if _, err := io.ReadFull(s.random, tokenBytes); err != nil {
		return User{}, "", fmt.Errorf("generate session token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	now := s.now().UTC()
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO auth_sessions (token_hash, user_id, created_at, expires_at)
		VALUES (?, ?, ?, ?)
	`, hashToken(token), user.ID, formatTime(now), formatTime(now.Add(s.ttl))); err != nil {
		return User{}, "", fmt.Errorf("create authentication session: %w", err)
	}
	return user, token, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (User, string, bool, error) {
	if token == "" {
		return User{}, "", false, nil
	}
	sessionHash := hashToken(token)
	var user User
	var expiresAt string
	err := s.db.QueryRowContext(ctx, `
		SELECT users.id, users.username, users.display_name, auth_sessions.expires_at
		FROM auth_sessions
		JOIN users ON users.id = auth_sessions.user_id
		WHERE auth_sessions.token_hash = ?
	`, sessionHash).Scan(&user.ID, &user.Username, &user.DisplayName, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, "", false, nil
	}
	if err != nil {
		return User{}, "", false, fmt.Errorf("load authentication session: %w", err)
	}
	expires, err := time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil {
		return User{}, "", false, fmt.Errorf("parse authentication session expiry: %w", err)
	}
	if !expires.After(s.now().UTC()) {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE token_hash = ?`, sessionHash)
		return User{}, "", false, nil
	}
	return user, sessionHash, true, nil
}

func (s *Service) SignOut(ctx context.Context) error {
	value, ok := ctx.Value(principalContextKey{}).(principal)
	if !ok {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE token_hash = ?`, value.sessionHash); err != nil {
		return fmt.Errorf("delete authentication session: %w", err)
	}
	return nil
}

func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cookie, err := request.Cookie(s.cookieName)
		if errors.Is(err, http.ErrNoCookie) {
			next.ServeHTTP(writer, request)
			return
		}
		if err != nil {
			http.Error(writer, "invalid authentication cookie", http.StatusBadRequest)
			return
		}
		user, sessionHash, authenticated, err := s.Authenticate(request.Context(), cookie.Value)
		if err != nil {
			http.Error(writer, "authentication unavailable", http.StatusInternalServerError)
			return
		}
		if authenticated {
			ctx := context.WithValue(request.Context(), principalContextKey{}, principal{user: user, sessionHash: sessionHash})
			request = request.WithContext(ctx)
		}
		next.ServeHTTP(writer, request)
	})
}

func CurrentUser(ctx context.Context) (User, bool) {
	value, ok := ctx.Value(principalContextKey{}).(principal)
	return value.user, ok
}

func RequireUser(ctx context.Context) (User, error) {
	user, ok := CurrentUser(ctx)
	if !ok {
		return User{}, ErrUnauthenticated
	}
	return user, nil
}

func (s *Service) SessionCookie(token string) *http.Cookie {
	return &http.Cookie{
		Name:     s.cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(s.ttl.Seconds()),
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteStrictMode,
	}
}

func (s *Service) ExpiredSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     s.cookieName,
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0).UTC(),
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteStrictMode,
	}
}

func normalizeUsername(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func hashToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}
