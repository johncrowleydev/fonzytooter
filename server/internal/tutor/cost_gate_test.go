package tutor

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/johncrowleydev/fonzytooter/server/internal/auth"
	"github.com/johncrowleydev/fonzytooter/server/internal/database"
)

var fixedUsageTime = time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)

func TestCostGateDefaultsToNotEntitled(t *testing.T) {
	gate := testCostGate(t, CostGateConfig{Now: func() time.Time { return fixedUsageTime }})
	if _, err := gate.ReserveTurn(context.Background(), testUserID); !errors.Is(err, ErrTutorNotEntitled) {
		t.Fatalf("reserve for non-entitled learner = %v, want %v", err, ErrTutorNotEntitled)
	}
	access, err := gate.Access(context.Background(), testUserID)
	if err != nil || access.Status != AccessNotEntitled {
		t.Fatalf("non-entitled access = %#v, %v", access, err)
	}
}

func TestCostGatePersistsHardMonthlyAllowancePerUser(t *testing.T) {
	gate, db := testCostGateWithDB(t, CostGateConfig{
		Entitled: true, MonthlyTurnLimit: 2, Now: func() time.Time { return fixedUsageTime },
	})
	secondUser := auth.UserID("00000000-0000-4000-8000-000000000002")
	insertTestTutorUser(t, db, secondUser)
	ctx := context.Background()

	for turn := 1; turn <= 2; turn++ {
		access, err := gate.ReserveTurn(ctx, testUserID)
		if err != nil || access.UsedTurns != turn {
			t.Fatalf("reserve turn %d = %#v, %v", turn, access, err)
		}
	}
	if _, err := gate.ReserveTurn(ctx, testUserID); !errors.Is(err, ErrTutorLimitExhausted) {
		t.Fatalf("third reserve = %v, want %v", err, ErrTutorLimitExhausted)
	}
	if access, err := gate.Access(ctx, secondUser); err != nil || access.Status != AccessAllowed || access.RemainingTurns != 2 {
		t.Fatalf("second learner access = %#v, %v", access, err)
	}
}

func TestCostGateConcurrentReservationsCannotCrossLimit(t *testing.T) {
	gate := testCostGate(t, CostGateConfig{
		Entitled: true, MonthlyTurnLimit: 3, Now: func() time.Time { return fixedUsageTime },
	})
	var allowed atomic.Int32
	var exhausted atomic.Int32
	var unexpected atomic.Value
	var wait sync.WaitGroup
	for request := 0; request < 12; request++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := gate.ReserveTurn(context.Background(), testUserID)
			switch {
			case err == nil:
				allowed.Add(1)
			case errors.Is(err, ErrTutorLimitExhausted):
				exhausted.Add(1)
			default:
				unexpected.Store(err)
			}
		}()
	}
	wait.Wait()
	if err, _ := unexpected.Load().(error); err != nil {
		t.Fatal(err)
	}
	if allowed.Load() != 3 || exhausted.Load() != 9 {
		t.Fatalf("allowed=%d exhausted=%d, want 3 and 9", allowed.Load(), exhausted.Load())
	}
}

func TestTutorServiceCountsProviderFailureAndBlocksRetryAtLimit(t *testing.T) {
	gate := testCostGate(t, CostGateConfig{
		Entitled: true, MonthlyTurnLimit: 1, Now: func() time.Time { return fixedUsageTime },
	})
	provider := &countingFailureProvider{}
	service := NewService(provider, gate)
	request := TurnRequest{Message: "hello"}

	if _, err := service.StreamTurn(context.Background(), testUserID, request); err == nil {
		t.Fatal("expected provider failure")
	}
	if provider.calls.Load() != 1 {
		t.Fatalf("provider calls = %d, want 1", provider.calls.Load())
	}
	if _, err := service.StreamTurn(context.Background(), testUserID, request); !errors.Is(err, ErrTutorLimitExhausted) {
		t.Fatalf("retry = %v, want %v", err, ErrTutorLimitExhausted)
	}
	if provider.calls.Load() != 1 {
		t.Fatalf("exhausted retry reached provider: calls=%d", provider.calls.Load())
	}
}

func TestTutorServiceRejectsMissingCostGateBeforeProviderInvocation(t *testing.T) {
	provider := &countingFailureProvider{}
	service := NewService(provider)

	_, err := service.StreamTurn(context.Background(), testUserID, TurnRequest{Message: "hello"})
	if !errors.Is(err, ErrTutorPolicyUnavailable) {
		t.Fatalf("missing cost gate = %v, want %v", err, ErrTutorPolicyUnavailable)
	}
	if provider.calls.Load() != 0 {
		t.Fatalf("missing cost gate reached provider: calls=%d", provider.calls.Load())
	}
}

type countingFailureProvider struct{ calls atomic.Int32 }

func (p *countingFailureProvider) Stream(context.Context, ModelRequest) (<-chan ProviderEvent, error) {
	p.calls.Add(1)
	return nil, errors.New("provider failed")
}

func testCostGate(t *testing.T, config CostGateConfig) *CostGate {
	t.Helper()
	gate, _ := testCostGateWithDB(t, config)
	return gate
}

func testCostGateWithDB(t *testing.T, config CostGateConfig) (*CostGate, *sql.DB) {
	t.Helper()
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	gate, err := NewCostGate(db, config)
	if err != nil {
		t.Fatal(err)
	}
	return gate, db
}

func insertTestTutorUser(t *testing.T, db *sql.DB, userID auth.UserID) {
	t.Helper()
	_, err := db.Exec(`
		INSERT INTO users (id, username, display_name, password_hash, created_at, updated_at)
		VALUES (?, ?, 'Second', 'unused-test-hash', ?, ?)`,
		userID, "second-"+string(userID), formatUsageTime(fixedUsageTime), formatUsageTime(fixedUsageTime),
	)
	if err != nil {
		t.Fatal(err)
	}
}
