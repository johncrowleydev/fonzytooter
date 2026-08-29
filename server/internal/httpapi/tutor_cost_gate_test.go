package httpapi

import (
	"context"
	"net/http"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/johncrowleydev/helix-academy/server/internal/auth"
	"github.com/johncrowleydev/helix-academy/server/internal/curriculum"
	"github.com/johncrowleydev/helix-academy/server/internal/database"
	"github.com/johncrowleydev/helix-academy/server/internal/tutor"
)

func TestTutorCostGateRejectsBeforeProviderInvocation(t *testing.T) {
	provider := &costGateProvider{}
	notEntitled := tutor.NewService(provider, httpCostGate(t, false, 0))
	anonymous := NewAPI(notEntitled, curriculum.NewEmptyCatalog(), nil, nil)
	response := authorizationRequest(t, anonymous.Handler, http.MethodPost, "/api/tutor/turns", `{"message":"hello"}`, nil)
	if response.Code != http.StatusUnauthorized || provider.calls.Load() != 0 {
		t.Fatalf("anonymous status=%d provider calls=%d", response.Code, provider.calls.Load())
	}

	notEntitledAPI, cookie := newTestAPIWithSession(t, notEntitled, curriculum.NewEmptyCatalog(), nil, nil)
	response = authorizationRequest(t, notEntitledAPI.Handler, http.MethodPost, "/api/tutor/turns", `{"message":"hello"}`, cookie)
	if response.Code != http.StatusForbidden || provider.calls.Load() != 0 {
		t.Fatalf("not-entitled status=%d provider calls=%d: %s", response.Code, provider.calls.Load(), response.Body.String())
	}

	exhaustedGate := httpCostGate(t, true, 1)
	if _, err := exhaustedGate.ReserveTurn(context.Background(), auth.BootstrapUserID); err != nil {
		t.Fatal(err)
	}
	exhaustedAPI, exhaustedCookie := newTestAPIWithSession(t, tutor.NewService(provider, exhaustedGate), curriculum.NewEmptyCatalog(), nil, nil)
	response = authorizationRequest(t, exhaustedAPI.Handler, http.MethodPost, "/api/tutor/turns", `{"message":"hello"}`, exhaustedCookie)
	if response.Code != http.StatusTooManyRequests || provider.calls.Load() != 0 {
		t.Fatalf("exhausted status=%d provider calls=%d: %s", response.Code, provider.calls.Load(), response.Body.String())
	}
}

func TestTutorCostGateAllowsAndReportsReservedTurn(t *testing.T) {
	provider := &costGateProvider{}
	service := tutor.NewService(provider, httpCostGate(t, true, 2))
	app, cookie := newTestAPIWithSession(t, service, curriculum.NewEmptyCatalog(), nil, nil)
	response := authorizationRequest(t, app.Handler, http.MethodPost, "/api/tutor/turns", `{"message":"hello"}`, cookie)
	if response.Code != http.StatusOK || provider.calls.Load() != 1 {
		t.Fatalf("allowed status=%d provider calls=%d: %s", response.Code, provider.calls.Load(), response.Body.String())
	}
	access := authorizationRequest(t, app.Handler, http.MethodGet, "/api/tutor-access", "", cookie)
	if access.Code != http.StatusOK || !containsJSONField(access.Body.String(), `"status":"allowed"`) || !containsJSONField(access.Body.String(), `"usedTurns":1`) {
		t.Fatalf("access status=%d: %s", access.Code, access.Body.String())
	}
}

func TestTutorTurnReturnsServiceUnavailableWhenCostGateIsMissing(t *testing.T) {
	provider := &costGateProvider{}
	app, cookie := newTestAPIWithSession(t, tutor.NewService(provider), curriculum.NewEmptyCatalog(), nil, nil)

	response := authorizationRequest(t, app.Handler, http.MethodPost, "/api/tutor/turns", `{"message":"hello"}`, cookie)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("missing cost gate status=%d, want 503: %s", response.Code, response.Body.String())
	}
	if provider.calls.Load() != 0 {
		t.Fatalf("missing cost gate reached provider: calls=%d", provider.calls.Load())
	}
}

type costGateProvider struct{ calls atomic.Int32 }

func (p *costGateProvider) Stream(ctx context.Context, _ tutor.ModelRequest) (<-chan tutor.ProviderEvent, error) {
	p.calls.Add(1)
	events := make(chan tutor.ProviderEvent, 2)
	events <- tutor.ProviderEvent{Type: tutor.ProviderEventTextDelta, Text: "Available tutor response."}
	events <- tutor.ProviderEvent{Type: tutor.ProviderEventCompleted}
	close(events)
	return events, nil
}

func httpCostGate(t *testing.T, entitled bool, limit int) *tutor.CostGate {
	t.Helper()
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "cost-gate.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	gate, err := tutor.NewCostGate(db, tutor.CostGateConfig{
		Entitled: entitled, MonthlyTurnLimit: limit,
		Now: func() time.Time { return time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatal(err)
	}
	return gate
}
