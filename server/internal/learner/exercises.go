package learner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const ActivityExerciseChecked = "exercise_checked"

const (
	DefaultExerciseHistoryLimit = 10
	MaxExerciseHistoryLimit     = 50
)

var (
	ErrExerciseNotFound       = errors.New("exercise not found")
	ErrInvalidExerciseAttempt = errors.New("invalid exercise attempt")
)

type ExerciseWorkspace struct {
	CourseID   string
	ModuleID   string
	ExerciseID string
	Code       string
	UpdatedAt  *time.Time
}

type ExerciseTestResult struct {
	TestID     string
	Status     string
	Message    string
	DurationMS int64
}

type ExerciseAttempt struct {
	ID           int64
	CourseID     string
	ModuleID     string
	ExerciseID   string
	CreatedAt    time.Time
	PassedCount  int
	FailedCount  int
	DurationMS   int64
	AllPassed    bool
	CodeSnapshot string
	Results      []ExerciseTestResult
}

func (s *Service) ExerciseWorkspace(ctx context.Context, courseID, moduleID, exerciseID string) (ExerciseWorkspace, error) {
	exercise, ok := s.catalog.ExerciseByCourse(courseID, moduleID, exerciseID)
	if !ok {
		return ExerciseWorkspace{}, ErrExerciseNotFound
	}
	workspace := ExerciseWorkspace{CourseID: courseID, ModuleID: moduleID, ExerciseID: exerciseID, Code: exercise.StarterCode}
	var updatedAt string
	err := s.db.QueryRowContext(ctx, `
		SELECT code, updated_at FROM exercise_workspaces
		WHERE course_id = ? AND module_id = ? AND exercise_id = ?
	`, courseID, moduleID, exerciseID).Scan(&workspace.Code, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return workspace, nil
	}
	if err != nil {
		return ExerciseWorkspace{}, fmt.Errorf("read exercise workspace: %w", err)
	}
	parsed, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return ExerciseWorkspace{}, fmt.Errorf("parse exercise workspace time: %w", err)
	}
	workspace.UpdatedAt = &parsed
	return workspace, nil
}

func (s *Service) SetExerciseWorkspace(ctx context.Context, courseID, moduleID, exerciseID, code string) (ExerciseWorkspace, error) {
	if _, ok := s.catalog.ExerciseByCourse(courseID, moduleID, exerciseID); !ok {
		return ExerciseWorkspace{}, ErrExerciseNotFound
	}
	now := s.now().UTC()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO exercise_workspaces (course_id, module_id, exercise_id, code, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (course_id, module_id, exercise_id) DO UPDATE SET
			code = excluded.code,
			updated_at = excluded.updated_at
	`, courseID, moduleID, exerciseID, code, formatLearnerTime(now))
	if err != nil {
		return ExerciseWorkspace{}, fmt.Errorf("write exercise workspace: %w", err)
	}
	return ExerciseWorkspace{CourseID: courseID, ModuleID: moduleID, ExerciseID: exerciseID, Code: code, UpdatedAt: &now}, nil
}

func (s *Service) CreateExerciseAttempt(ctx context.Context, courseID, moduleID, exerciseID, code string, durationMS int64, results []ExerciseTestResult) (ExerciseAttempt, error) {
	exercise, ok := s.catalog.ExerciseByCourse(courseID, moduleID, exerciseID)
	if !ok {
		return ExerciseAttempt{}, ErrExerciseNotFound
	}
	if durationMS < 0 || len(results) != len(exercise.Tests) {
		return ExerciseAttempt{}, ErrInvalidExerciseAttempt
	}
	known := make(map[string]struct{}, len(exercise.Tests))
	for _, test := range exercise.Tests {
		known[test.ID] = struct{}{}
	}
	seen := make(map[string]struct{}, len(results))
	byID := make(map[string]ExerciseTestResult, len(results))
	for _, result := range results {
		if _, ok := known[result.TestID]; !ok {
			return ExerciseAttempt{}, ErrInvalidExerciseAttempt
		}
		if _, duplicate := seen[result.TestID]; duplicate {
			return ExerciseAttempt{}, ErrInvalidExerciseAttempt
		}
		seen[result.TestID] = struct{}{}
		byID[result.TestID] = result
		if result.DurationMS < 0 || (result.Status != "passed" && result.Status != "failed" && result.Status != "error") {
			return ExerciseAttempt{}, ErrInvalidExerciseAttempt
		}
	}
	normalized := make([]ExerciseTestResult, 0, len(exercise.Tests))
	passed := 0
	for _, authoredTest := range exercise.Tests {
		result := byID[authoredTest.ID]
		normalized = append(normalized, result)
		if result.Status == "passed" {
			passed++
		}
	}

	now := s.now().UTC()
	attempt := ExerciseAttempt{
		CourseID: courseID, ModuleID: moduleID, ExerciseID: exerciseID, CreatedAt: now,
		PassedCount: passed, FailedCount: len(results) - passed, DurationMS: durationMS,
		AllPassed: passed == len(normalized), CodeSnapshot: code, Results: normalized,
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ExerciseAttempt{}, fmt.Errorf("begin exercise attempt: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	createdAt := formatLearnerTime(now)
	result, err := tx.ExecContext(ctx, `
		INSERT INTO exercise_attempts
			(course_id, module_id, exercise_id, created_at, passed_count, failed_count, duration_ms, all_passed, code_snapshot)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, courseID, moduleID, exerciseID, createdAt, attempt.PassedCount, attempt.FailedCount, durationMS, attempt.AllPassed, code)
	if err != nil {
		return ExerciseAttempt{}, fmt.Errorf("insert exercise attempt: %w", err)
	}
	attempt.ID, err = result.LastInsertId()
	if err != nil {
		return ExerciseAttempt{}, fmt.Errorf("read exercise attempt id: %w", err)
	}
	for _, testResult := range normalized {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO exercise_test_results (attempt_id, test_id, status, message, duration_ms)
			VALUES (?, ?, ?, ?, ?)
		`, attempt.ID, testResult.TestID, testResult.Status, testResult.Message, testResult.DurationMS)
		if err != nil {
			return ExerciseAttempt{}, fmt.Errorf("insert exercise test result: %w", err)
		}
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO activities (kind, course_id, module_id, exercise_id, occurred_at)
		VALUES (?, ?, ?, ?, ?)
	`, ActivityExerciseChecked, courseID, moduleID, exerciseID, createdAt)
	if err != nil {
		return ExerciseAttempt{}, fmt.Errorf("record exercise activity: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return ExerciseAttempt{}, fmt.Errorf("commit exercise attempt: %w", err)
	}
	return attempt, nil
}

func (s *Service) ExerciseAttempts(ctx context.Context, courseID, moduleID, exerciseID string, limit int) ([]ExerciseAttempt, error) {
	if _, ok := s.catalog.ExerciseByCourse(courseID, moduleID, exerciseID); !ok {
		return nil, ErrExerciseNotFound
	}
	if limit <= 0 {
		limit = DefaultExerciseHistoryLimit
	}
	if limit > MaxExerciseHistoryLimit {
		limit = MaxExerciseHistoryLimit
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, created_at, passed_count, failed_count, duration_ms, all_passed, code_snapshot
		FROM exercise_attempts
		WHERE course_id = ? AND module_id = ? AND exercise_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, courseID, moduleID, exerciseID, limit)
	if err != nil {
		return nil, fmt.Errorf("read exercise attempts: %w", err)
	}
	attempts := make([]ExerciseAttempt, 0)
	for rows.Next() {
		attempt := ExerciseAttempt{CourseID: courseID, ModuleID: moduleID, ExerciseID: exerciseID}
		var createdAt string
		if err := rows.Scan(&attempt.ID, &createdAt, &attempt.PassedCount, &attempt.FailedCount, &attempt.DurationMS, &attempt.AllPassed, &attempt.CodeSnapshot); err != nil {
			return nil, fmt.Errorf("scan exercise attempt: %w", err)
		}
		attempt.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse exercise attempt time: %w", err)
		}
		attempts = append(attempts, attempt)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("iterate exercise attempts: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close exercise attempts: %w", err)
	}
	for index := range attempts {
		resultRows, err := s.db.QueryContext(ctx, `
			SELECT test_id, status, message, duration_ms
			FROM exercise_test_results WHERE attempt_id = ? ORDER BY rowid
		`, attempts[index].ID)
		if err != nil {
			return nil, fmt.Errorf("read exercise attempt results: %w", err)
		}
		for resultRows.Next() {
			var result ExerciseTestResult
			if err := resultRows.Scan(&result.TestID, &result.Status, &result.Message, &result.DurationMS); err != nil {
				_ = resultRows.Close()
				return nil, fmt.Errorf("scan exercise attempt result: %w", err)
			}
			attempts[index].Results = append(attempts[index].Results, result)
		}
		if err := resultRows.Err(); err != nil {
			_ = resultRows.Close()
			return nil, fmt.Errorf("iterate exercise attempt results: %w", err)
		}
		_ = resultRows.Close()
		if attempts[index].Results == nil {
			attempts[index].Results = []ExerciseTestResult{}
		}
	}
	return attempts, nil
}
