package learner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/johncrowleydev/fonzytooter/server/internal/auth"
)

const ActivityVideoCompleted = "video_completed"

var ErrVideoNotFound = errors.New("video not found")

type VideoProgress struct {
	CourseID    string
	ModuleID    string
	VideoID     string
	Completed   bool
	CompletedAt *time.Time
}

func (s *Service) VideoProgress(ctx context.Context, userID auth.UserID, courseID, moduleID, videoID string) (VideoProgress, error) {
	if !s.videoExists(courseID, moduleID, videoID) {
		return VideoProgress{}, ErrVideoNotFound
	}

	progress := VideoProgress{CourseID: courseID, ModuleID: moduleID, VideoID: videoID}
	var completed bool
	var completedAt sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT completed, completed_at
		FROM video_progress
		WHERE user_id = ? AND course_id = ? AND module_id = ? AND video_id = ?
	`, userID, courseID, moduleID, videoID).Scan(&completed, &completedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return progress, nil
	}
	if err != nil {
		return VideoProgress{}, fmt.Errorf("read video progress: %w", err)
	}
	return videoProgressFromStored(courseID, moduleID, videoID, completed, completedAt)
}

func (s *Service) SetVideoProgress(ctx context.Context, userID auth.UserID, courseID, moduleID, videoID string, completed bool) (VideoProgress, error) {
	if !s.videoExists(courseID, moduleID, videoID) {
		return VideoProgress{}, ErrVideoNotFound
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return VideoProgress{}, fmt.Errorf("begin video progress update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var existingCompleted bool
	var existingCompletedAt sql.NullString
	err = tx.QueryRowContext(ctx, `
		SELECT completed, completed_at
		FROM video_progress
		WHERE user_id = ? AND course_id = ? AND module_id = ? AND video_id = ?
	`, userID, courseID, moduleID, videoID).Scan(&existingCompleted, &existingCompletedAt)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return VideoProgress{}, fmt.Errorf("read video progress for update: %w", err)
	}
	if err == nil && existingCompleted == completed {
		if err := tx.Commit(); err != nil {
			return VideoProgress{}, fmt.Errorf("commit unchanged video progress: %w", err)
		}
		return videoProgressFromStored(courseID, moduleID, videoID, existingCompleted, existingCompletedAt)
	}
	if errors.Is(err, sql.ErrNoRows) && !completed {
		if err := tx.Commit(); err != nil {
			return VideoProgress{}, fmt.Errorf("commit default video progress: %w", err)
		}
		return VideoProgress{CourseID: courseID, ModuleID: moduleID, VideoID: videoID}, nil
	}

	now := s.now().UTC()
	nowText := formatLearnerTime(now)
	var completedAt any
	if completed {
		completedAt = nowText
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO video_progress (user_id, course_id, module_id, video_id, completed, completed_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (user_id, course_id, module_id, video_id) DO UPDATE SET
			completed = excluded.completed,
			completed_at = excluded.completed_at,
			updated_at = excluded.updated_at
	`, userID, courseID, moduleID, videoID, completed, completedAt, nowText); err != nil {
		return VideoProgress{}, fmt.Errorf("write video progress: %w", err)
	}
	if completed {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO activities (user_id, kind, course_id, module_id, video_id, occurred_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, userID, ActivityVideoCompleted, courseID, moduleID, videoID, nowText); err != nil {
			return VideoProgress{}, fmt.Errorf("record video completion activity: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return VideoProgress{}, fmt.Errorf("commit video progress update: %w", err)
	}

	progress := VideoProgress{CourseID: courseID, ModuleID: moduleID, VideoID: videoID, Completed: completed}
	if completed {
		progress.CompletedAt = &now
	}
	return progress, nil
}

func (s *Service) videoExists(courseID, moduleID, videoID string) bool {
	module, ok := s.catalog.ModuleByCourse(courseID, moduleID)
	if !ok {
		return false
	}
	for _, video := range module.Videos {
		if video.ID == videoID {
			return true
		}
	}
	return false
}

func videoProgressFromStored(courseID, moduleID, videoID string, completed bool, completedAt sql.NullString) (VideoProgress, error) {
	progress := VideoProgress{CourseID: courseID, ModuleID: moduleID, VideoID: videoID, Completed: completed}
	if completedAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, completedAt.String)
		if err != nil {
			return VideoProgress{}, fmt.Errorf("parse video completion time: %w", err)
		}
		progress.CompletedAt = &parsed
	}
	return progress, nil
}
