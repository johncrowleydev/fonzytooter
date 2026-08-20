package learner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
)

const (
	ActivityLessonCompleted = "lesson_completed"
	EvidenceNotAssessed     = "not_assessed"
	DefaultActivityLimit    = 20
	MaxActivityLimit        = 100
)

var (
	ErrCourseNotFound = errors.New("course not found")
	ErrLessonNotFound = errors.New("lesson not found")
)

type Service struct {
	db      *sql.DB
	catalog *curriculum.Catalog
	now     func() time.Time
}

type LessonProgress struct {
	CourseID    string
	ModuleID    string
	LessonID    string
	Completed   bool
	CompletedAt *time.Time
}

type CourseProgress struct {
	CourseID             string
	CompletedLessonCount int
	TotalLessonCount     int
	Objectives           []ObjectiveProgress
	NextLesson           *NextLesson
}

type ObjectiveProgress struct {
	CourseID    string
	ModuleID    string
	ID          string
	Title       string
	Description string
	Introduced  bool
	Recall      string
	Application string
	Transfer    string
}

type NextLesson struct {
	CourseID    string
	ModuleID    string
	ModuleTitle string
	LessonID    string
	LessonTitle string
}

type Activity struct {
	ID           int64
	Kind         string
	CourseID     string
	CourseTitle  string
	ModuleID     *string
	ModuleTitle  *string
	LessonID     *string
	LessonTitle  *string
	ExerciseID   *string
	ReviewItemID *string
	OccurredAt   time.Time
}

func NewService(db *sql.DB, catalog *curriculum.Catalog) *Service {
	if db == nil {
		panic("learner.NewService: nil database")
	}
	if catalog == nil {
		panic("learner.NewService: nil curriculum catalog")
	}
	return &Service{db: db, catalog: catalog, now: time.Now}
}

func (s *Service) LessonProgress(ctx context.Context, courseID, moduleID, lessonID string) (LessonProgress, error) {
	if _, ok := s.catalog.LessonByCourse(courseID, moduleID, lessonID); !ok {
		return LessonProgress{}, ErrLessonNotFound
	}

	progress := LessonProgress{CourseID: courseID, ModuleID: moduleID, LessonID: lessonID}
	var completed bool
	var completedAt sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT completed, completed_at
		FROM lesson_progress
		WHERE course_id = ? AND module_id = ? AND lesson_id = ?
	`, courseID, moduleID, lessonID).Scan(&completed, &completedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return progress, nil
	}
	if err != nil {
		return LessonProgress{}, fmt.Errorf("read lesson progress: %w", err)
	}
	progress.Completed = completed
	if completedAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, completedAt.String)
		if err != nil {
			return LessonProgress{}, fmt.Errorf("parse lesson completion time: %w", err)
		}
		progress.CompletedAt = &parsed
	}
	return progress, nil
}

func (s *Service) SetLessonProgress(ctx context.Context, courseID, moduleID, lessonID string, completed bool) (LessonProgress, error) {
	if _, ok := s.catalog.LessonByCourse(courseID, moduleID, lessonID); !ok {
		return LessonProgress{}, ErrLessonNotFound
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return LessonProgress{}, fmt.Errorf("begin lesson progress update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var existingCompleted bool
	var existingCompletedAt sql.NullString
	err = tx.QueryRowContext(ctx, `
		SELECT completed, completed_at
		FROM lesson_progress
		WHERE course_id = ? AND module_id = ? AND lesson_id = ?
	`, courseID, moduleID, lessonID).Scan(&existingCompleted, &existingCompletedAt)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return LessonProgress{}, fmt.Errorf("read lesson progress for update: %w", err)
	}

	if err == nil && existingCompleted == completed {
		if err := tx.Commit(); err != nil {
			return LessonProgress{}, fmt.Errorf("commit unchanged lesson progress: %w", err)
		}
		return lessonProgressFromStored(courseID, moduleID, lessonID, existingCompleted, existingCompletedAt)
	}
	if errors.Is(err, sql.ErrNoRows) && !completed {
		if err := tx.Commit(); err != nil {
			return LessonProgress{}, fmt.Errorf("commit default lesson progress: %w", err)
		}
		return LessonProgress{CourseID: courseID, ModuleID: moduleID, LessonID: lessonID}, nil
	}

	now := s.now().UTC()
	nowText := now.Format(time.RFC3339Nano)
	var completedAt any
	if completed {
		completedAt = nowText
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO lesson_progress (course_id, module_id, lesson_id, completed, completed_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (course_id, module_id, lesson_id) DO UPDATE SET
			completed = excluded.completed,
			completed_at = excluded.completed_at,
			updated_at = excluded.updated_at
	`, courseID, moduleID, lessonID, completed, completedAt, nowText)
	if err != nil {
		return LessonProgress{}, fmt.Errorf("write lesson progress: %w", err)
	}
	if completed {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO activities (kind, course_id, module_id, lesson_id, occurred_at)
			VALUES (?, ?, ?, ?, ?)
		`, ActivityLessonCompleted, courseID, moduleID, lessonID, nowText); err != nil {
			return LessonProgress{}, fmt.Errorf("record lesson completion activity: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return LessonProgress{}, fmt.Errorf("commit lesson progress update: %w", err)
	}

	progress := LessonProgress{CourseID: courseID, ModuleID: moduleID, LessonID: lessonID, Completed: completed}
	if completed {
		progress.CompletedAt = &now
	}
	return progress, nil
}

func (s *Service) CourseProgress(ctx context.Context, courseID string) (CourseProgress, error) {
	course, ok := s.catalog.CourseByID(courseID)
	if !ok {
		return CourseProgress{}, ErrCourseNotFound
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT module_id, lesson_id
		FROM lesson_progress
		WHERE course_id = ? AND completed = 1
	`, courseID)
	if err != nil {
		return CourseProgress{}, fmt.Errorf("read completed lessons: %w", err)
	}
	defer rows.Close()
	completed := make(map[string]struct{})
	for rows.Next() {
		var moduleID, lessonID string
		if err := rows.Scan(&moduleID, &lessonID); err != nil {
			return CourseProgress{}, fmt.Errorf("scan completed lesson: %w", err)
		}
		completed[lessonKey(moduleID, lessonID)] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return CourseProgress{}, fmt.Errorf("iterate completed lessons: %w", err)
	}

	result := CourseProgress{
		CourseID:   courseID,
		Objectives: []ObjectiveProgress{},
	}
	introduced := make(map[string]bool)
	for _, module := range course.Modules {
		for _, lesson := range module.Lessons {
			result.TotalLessonCount++
			if _, ok := completed[lessonKey(module.ID, lesson.ID)]; ok {
				result.CompletedLessonCount++
				for _, objectiveID := range lesson.ObjectiveIDs {
					introduced[objectiveID] = true
				}
			} else if result.NextLesson == nil {
				result.NextLesson = &NextLesson{
					CourseID: course.ID, ModuleID: module.ID, ModuleTitle: module.Title,
					LessonID: lesson.ID, LessonTitle: lesson.Title,
				}
			}
		}
	}
	for _, module := range course.Modules {
		for _, objective := range module.Objectives {
			result.Objectives = append(result.Objectives, ObjectiveProgress{
				CourseID: course.ID, ModuleID: module.ID, ID: objective.ID,
				Title: objective.Title, Description: objective.Description,
				Introduced: introduced[objective.ID], Recall: EvidenceNotAssessed,
				Application: EvidenceNotAssessed, Transfer: EvidenceNotAssessed,
			})
		}
	}
	return result, nil
}

func (s *Service) Activities(ctx context.Context, courseID string, limit int) ([]Activity, error) {
	course, ok := s.catalog.CourseByID(courseID)
	if !ok {
		return nil, ErrCourseNotFound
	}
	if limit <= 0 {
		limit = DefaultActivityLimit
	}
	if limit > MaxActivityLimit {
		limit = MaxActivityLimit
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, kind, course_id, module_id, lesson_id, exercise_id, review_item_id, occurred_at
		FROM activities
		WHERE course_id = ?
		ORDER BY occurred_at DESC, id DESC
		LIMIT ?
	`, courseID, limit)
	if err != nil {
		return nil, fmt.Errorf("read learner activities: %w", err)
	}
	defer rows.Close()

	activities := make([]Activity, 0)
	for rows.Next() {
		var activity Activity
		var moduleID, lessonID, exerciseID, reviewItemID sql.NullString
		var occurredAt string
		if err := rows.Scan(&activity.ID, &activity.Kind, &activity.CourseID, &moduleID, &lessonID, &exerciseID, &reviewItemID, &occurredAt); err != nil {
			return nil, fmt.Errorf("scan learner activity: %w", err)
		}
		activity.CourseTitle = course.Title
		activity.ModuleID = nullableString(moduleID)
		activity.LessonID = nullableString(lessonID)
		activity.ExerciseID = nullableString(exerciseID)
		activity.ReviewItemID = nullableString(reviewItemID)
		activity.OccurredAt, err = time.Parse(time.RFC3339Nano, occurredAt)
		if err != nil {
			return nil, fmt.Errorf("parse learner activity time: %w", err)
		}
		if activity.ModuleID != nil {
			if module, ok := s.catalog.ModuleByCourse(courseID, *activity.ModuleID); ok {
				activity.ModuleTitle = stringPointer(module.Title)
			}
		}
		if activity.ModuleID != nil && activity.LessonID != nil {
			if lesson, ok := s.catalog.LessonByCourse(courseID, *activity.ModuleID, *activity.LessonID); ok {
				activity.LessonTitle = stringPointer(lesson.Title)
			}
		}
		activities = append(activities, activity)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate learner activities: %w", err)
	}
	return activities, nil
}

func lessonProgressFromStored(courseID, moduleID, lessonID string, completed bool, completedAt sql.NullString) (LessonProgress, error) {
	progress := LessonProgress{CourseID: courseID, ModuleID: moduleID, LessonID: lessonID, Completed: completed}
	if completedAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, completedAt.String)
		if err != nil {
			return LessonProgress{}, fmt.Errorf("parse lesson completion time: %w", err)
		}
		progress.CompletedAt = &parsed
	}
	return progress, nil
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return stringPointer(value.String)
}

func stringPointer(value string) *string {
	return &value
}

func lessonKey(moduleID, lessonID string) string {
	return moduleID + "\x00" + lessonID
}
