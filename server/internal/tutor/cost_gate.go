package tutor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/johncrowleydev/helix-academy/server/internal/auth"
)

var (
	ErrTutorNotEntitled       = errors.New("tutor is unavailable for this account")
	ErrTutorLimitExhausted    = errors.New("tutor usage limit reached")
	ErrTutorPolicyUnavailable = errors.New("tutor access policy is unavailable")
)

type AccessStatus string

const (
	AccessAllowed        AccessStatus = "allowed"
	AccessNotEntitled    AccessStatus = "not_entitled"
	AccessLimitExhausted AccessStatus = "limit_exhausted"
	AccessUnavailable    AccessStatus = "unavailable"
)

type Access struct {
	Status           AccessStatus
	MonthlyTurnLimit int
	UsedTurns        int
	RemainingTurns   int
	WindowEndsAt     time.Time
}

type CostGateConfig struct {
	Entitled         bool
	MonthlyTurnLimit int
	Now              func() time.Time
}

type CostGate struct {
	db               *sql.DB
	entitled         bool
	monthlyTurnLimit int
	now              func() time.Time
}

func NewCostGate(db *sql.DB, config CostGateConfig) (*CostGate, error) {
	if db == nil {
		return nil, errors.New("tutor cost gate database is nil")
	}
	if config.Entitled && config.MonthlyTurnLimit <= 0 {
		return nil, errors.New("entitled tutor cost gate requires a positive monthly turn limit")
	}
	if config.MonthlyTurnLimit < 0 {
		return nil, errors.New("tutor monthly turn limit cannot be negative")
	}
	now := config.Now
	if now == nil {
		now = time.Now
	}
	return &CostGate{
		db: db, entitled: config.Entitled, monthlyTurnLimit: config.MonthlyTurnLimit, now: now,
	}, nil
}

func (g *CostGate) Access(ctx context.Context, userID auth.UserID) (Access, error) {
	windowStart, windowEnd := monthlyWindow(g.now())
	access := Access{
		MonthlyTurnLimit: g.monthlyTurnLimit,
		WindowEndsAt:     windowEnd,
	}
	if !g.entitled {
		access.Status = AccessNotEntitled
		return access, nil
	}

	err := g.db.QueryRowContext(ctx,
		`SELECT reserved_turns FROM tutor_usage_windows WHERE user_id = ? AND window_start = ?`,
		userID, formatUsageTime(windowStart),
	).Scan(&access.UsedTurns)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Access{}, fmt.Errorf("%w: read tutor usage: %v", ErrTutorPolicyUnavailable, err)
	}
	access.RemainingTurns = max(g.monthlyTurnLimit-access.UsedTurns, 0)
	if access.RemainingTurns == 0 {
		access.Status = AccessLimitExhausted
	} else {
		access.Status = AccessAllowed
	}
	return access, nil
}

func (g *CostGate) ReserveTurn(ctx context.Context, userID auth.UserID) (Access, error) {
	if !g.entitled {
		return Access{Status: AccessNotEntitled}, ErrTutorNotEntitled
	}
	windowStart, windowEnd := monthlyWindow(g.now())
	var used int
	err := g.db.QueryRowContext(ctx, `
		INSERT INTO tutor_usage_windows (user_id, window_start, reserved_turns, updated_at)
		VALUES (?, ?, 1, ?)
		ON CONFLICT (user_id, window_start) DO UPDATE SET
			reserved_turns = tutor_usage_windows.reserved_turns + 1,
			updated_at = excluded.updated_at
		WHERE tutor_usage_windows.reserved_turns < ?
		RETURNING reserved_turns`,
		userID, formatUsageTime(windowStart), formatUsageTime(g.now()), g.monthlyTurnLimit,
	).Scan(&used)
	if errors.Is(err, sql.ErrNoRows) {
		return Access{
			Status: AccessLimitExhausted, MonthlyTurnLimit: g.monthlyTurnLimit,
			UsedTurns: g.monthlyTurnLimit, RemainingTurns: 0, WindowEndsAt: windowEnd,
		}, ErrTutorLimitExhausted
	}
	if err != nil {
		return Access{}, fmt.Errorf("%w: reserve tutor usage: %v", ErrTutorPolicyUnavailable, err)
	}
	return Access{
		Status: AccessAllowed, MonthlyTurnLimit: g.monthlyTurnLimit,
		UsedTurns: used, RemainingTurns: max(g.monthlyTurnLimit-used, 0), WindowEndsAt: windowEnd,
	}, nil
}

func monthlyWindow(now time.Time) (time.Time, time.Time) {
	utc := now.UTC()
	start := time.Date(utc.Year(), utc.Month(), 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 1, 0)
}

func formatUsageTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}
