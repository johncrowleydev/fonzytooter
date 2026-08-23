package learner

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/johncrowleydev/fonzytooter/server/internal/auth"
	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
)

type queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

// SourceLessonEligibility is the learner-state policy for introducing virtual
// review cards. Persisted cards remain governed by their stored schedule.
type SourceLessonEligibility struct {
	courseID         string
	completedLessons map[string]struct{}
}

func LoadSourceLessonEligibility(ctx context.Context, db queryer, userID auth.UserID, courseID string) (SourceLessonEligibility, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT module_id, lesson_id
		FROM lesson_progress
		WHERE user_id = ? AND course_id = ? AND completed = 1
	`, userID, courseID)
	if err != nil {
		return SourceLessonEligibility{}, fmt.Errorf("read source lesson eligibility: %w", err)
	}
	defer rows.Close()

	eligibility := SourceLessonEligibility{
		courseID:         courseID,
		completedLessons: make(map[string]struct{}),
	}
	for rows.Next() {
		var moduleID, lessonID string
		if err := rows.Scan(&moduleID, &lessonID); err != nil {
			return SourceLessonEligibility{}, fmt.Errorf("scan source lesson eligibility: %w", err)
		}
		eligibility.completedLessons[lessonKey(moduleID, lessonID)] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return SourceLessonEligibility{}, fmt.Errorf("iterate source lesson eligibility: %w", err)
	}
	return eligibility, nil
}

func (eligibility SourceLessonEligibility) Allows(item curriculum.ReviewItem) bool {
	if item.CourseID != eligibility.courseID {
		return false
	}
	_, completed := eligibility.completedLessons[lessonKey(item.ModuleID, item.SourceLessonID)]
	return completed
}
