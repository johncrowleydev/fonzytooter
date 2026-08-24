package learner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/johncrowleydev/fonzytooter/server/internal/auth"
	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
)

const (
	ActivityLessonCompleted = "lesson_completed"
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
	DueReviewCount       int
	Objectives           []ObjectiveProgress
	NextLesson           *NextLesson
	PracticeExercise     *PracticeExercise
}

type ObjectiveProgress struct {
	CourseID             string
	ModuleID             string
	ID                   string
	Title                string
	Description          string
	Introduced           bool
	LinkedLessonCount    int
	CompletedLessonCount int
	Recall               RecallEvidence
	Application          ApplicationEvidence
	TransferAssessed     bool
}

type RecallEvidence struct {
	ReviewItemCount  int
	ReviewsCompleted int
	DueReviewCount   int
	LastReviewedAt   *time.Time
	NextDueAt        *time.Time
}

type ApplicationEvidence struct {
	ExerciseCount       int
	Attempts            int
	FullyPassedAttempts int
	LastCheckedAt       *time.Time
}

type NextLesson struct {
	CourseID    string
	ModuleID    string
	ModuleTitle string
	LessonID    string
	LessonTitle string
}

type PracticeExercise struct {
	CourseID      string
	ModuleID      string
	ModuleTitle   string
	ExerciseID    string
	ExerciseTitle string
}

type courseLessonRef struct {
	module curriculum.Module
	lesson curriculum.Lesson
}

type Activity struct {
	ID            int64
	Kind          string
	CourseID      string
	CourseTitle   string
	ModuleID      *string
	ModuleTitle   *string
	LessonID      *string
	LessonTitle   *string
	ExerciseID    *string
	ExerciseTitle *string
	VideoID       *string
	VideoTitle    *string
	ReviewItemID  *string
	OccurredAt    time.Time
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

func (s *Service) LessonProgress(ctx context.Context, userID auth.UserID, courseID, moduleID, lessonID string) (LessonProgress, error) {
	if _, ok := s.catalog.LessonByCourse(courseID, moduleID, lessonID); !ok {
		return LessonProgress{}, ErrLessonNotFound
	}

	progress := LessonProgress{CourseID: courseID, ModuleID: moduleID, LessonID: lessonID}
	var completed bool
	var completedAt sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT completed, completed_at
		FROM lesson_progress
		WHERE user_id = ? AND course_id = ? AND module_id = ? AND lesson_id = ?
	`, userID, courseID, moduleID, lessonID).Scan(&completed, &completedAt)
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

func (s *Service) SetLessonProgress(ctx context.Context, userID auth.UserID, courseID, moduleID, lessonID string, completed bool) (LessonProgress, error) {
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
		WHERE user_id = ? AND course_id = ? AND module_id = ? AND lesson_id = ?
	`, userID, courseID, moduleID, lessonID).Scan(&existingCompleted, &existingCompletedAt)
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
	nowText := formatLearnerTime(now)
	var completedAt any
	if completed {
		completedAt = nowText
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO lesson_progress (user_id, course_id, module_id, lesson_id, completed, completed_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (user_id, course_id, module_id, lesson_id) DO UPDATE SET
			completed = excluded.completed,
			completed_at = excluded.completed_at,
			updated_at = excluded.updated_at
	`, userID, courseID, moduleID, lessonID, completed, completedAt, nowText)
	if err != nil {
		return LessonProgress{}, fmt.Errorf("write lesson progress: %w", err)
	}
	if completed {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO activities (user_id, kind, course_id, module_id, lesson_id, occurred_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, userID, ActivityLessonCompleted, courseID, moduleID, lessonID, nowText); err != nil {
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

func (s *Service) CourseProgress(ctx context.Context, userID auth.UserID, courseID string) (CourseProgress, error) {
	course, ok := s.catalog.CourseByID(courseID)
	if !ok {
		return CourseProgress{}, ErrCourseNotFound
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT module_id, lesson_id, completed_at
		FROM lesson_progress
		WHERE user_id = ? AND course_id = ? AND completed = 1
	`, userID, courseID)
	if err != nil {
		return CourseProgress{}, fmt.Errorf("read completed lessons: %w", err)
	}
	defer rows.Close()
	completed := make(map[string]time.Time)
	for rows.Next() {
		var moduleID, lessonID string
		var completedAt sql.NullString
		if err := rows.Scan(&moduleID, &lessonID, &completedAt); err != nil {
			return CourseProgress{}, fmt.Errorf("scan completed lesson: %w", err)
		}
		when := time.Time{}
		if completedAt.Valid {
			when, err = time.Parse(time.RFC3339Nano, completedAt.String)
			if err != nil {
				return CourseProgress{}, fmt.Errorf("parse completed lesson time: %w", err)
			}
		}
		completed[lessonKey(moduleID, lessonID)] = when
	}
	if err := rows.Err(); err != nil {
		return CourseProgress{}, fmt.Errorf("iterate completed lessons: %w", err)
	}

	result := CourseProgress{
		CourseID:   courseID,
		Objectives: []ObjectiveProgress{},
	}
	introduced := make(map[string]bool)
	objectiveIndex := make(map[string]int)
	orderedLessons := make([]courseLessonRef, 0)
	for _, module := range course.Modules {
		for _, lesson := range module.Lessons {
			orderedLessons = append(orderedLessons, courseLessonRef{module: module, lesson: lesson})
			result.TotalLessonCount++
			if _, ok := completed[lessonKey(module.ID, lesson.ID)]; ok {
				result.CompletedLessonCount++
				for _, objectiveID := range lesson.ObjectiveIDs {
					introduced[objectiveID] = true
				}
			}
		}
	}
	for _, module := range course.Modules {
		for _, objective := range module.Objectives {
			objectiveIndex[objective.ID] = len(result.Objectives)
			result.Objectives = append(result.Objectives, ObjectiveProgress{
				CourseID: course.ID, ModuleID: module.ID, ID: objective.ID,
				Title: objective.Title, Description: objective.Description,
				Introduced: introduced[objective.ID],
			})
		}
	}
	for _, item := range orderedLessons {
		for _, objectiveID := range item.lesson.ObjectiveIDs {
			index, ok := objectiveIndex[objectiveID]
			if !ok {
				continue
			}
			result.Objectives[index].LinkedLessonCount++
			if _, ok := completed[lessonKey(item.module.ID, item.lesson.ID)]; ok {
				result.Objectives[index].CompletedLessonCount++
			}
		}
	}
	result.NextLesson = nextLesson(course, orderedLessons, completed)
	if err := s.addCourseEvidence(ctx, userID, course, objectiveIndex, &result); err != nil {
		return CourseProgress{}, err
	}
	return result, nil
}

func nextLesson(course curriculum.Course, lessons []courseLessonRef, completed map[string]time.Time) *NextLesson {
	latestIndex := -1
	latestAt := time.Time{}
	for index, item := range lessons {
		when, ok := completed[lessonKey(item.module.ID, item.lesson.ID)]
		if ok && (latestIndex < 0 || when.After(latestAt)) {
			latestIndex, latestAt = index, when
		}
	}
	if latestIndex >= 0 && latestIndex+1 < len(lessons) {
		candidate := lessons[latestIndex+1]
		if _, done := completed[lessonKey(candidate.module.ID, candidate.lesson.ID)]; !done {
			return nextLessonValue(course, candidate.module, candidate.lesson)
		}
	}
	for _, candidate := range lessons {
		if _, done := completed[lessonKey(candidate.module.ID, candidate.lesson.ID)]; !done {
			return nextLessonValue(course, candidate.module, candidate.lesson)
		}
	}
	return nil
}

func nextLessonValue(course curriculum.Course, module curriculum.Module, lesson curriculum.Lesson) *NextLesson {
	return &NextLesson{CourseID: course.ID, ModuleID: module.ID, ModuleTitle: module.Title, LessonID: lesson.ID, LessonTitle: lesson.Title}
}

func (s *Service) Activities(ctx context.Context, userID auth.UserID, courseID string, limit int) ([]Activity, error) {
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
		SELECT id, kind, course_id, module_id, lesson_id, exercise_id, video_id, review_item_id, occurred_at
		FROM activities
		WHERE user_id = ? AND course_id = ?
		ORDER BY occurred_at DESC, id DESC
		LIMIT ?
	`, userID, courseID, limit)
	if err != nil {
		return nil, fmt.Errorf("read learner activities: %w", err)
	}
	defer rows.Close()

	activities := make([]Activity, 0)
	for rows.Next() {
		var activity Activity
		var moduleID, lessonID, exerciseID, videoID, reviewItemID sql.NullString
		var occurredAt string
		if err := rows.Scan(&activity.ID, &activity.Kind, &activity.CourseID, &moduleID, &lessonID, &exerciseID, &videoID, &reviewItemID, &occurredAt); err != nil {
			return nil, fmt.Errorf("scan learner activity: %w", err)
		}
		activity.CourseTitle = course.Title
		activity.ModuleID = nullableString(moduleID)
		activity.LessonID = nullableString(lessonID)
		activity.ExerciseID = nullableString(exerciseID)
		activity.VideoID = nullableString(videoID)
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
		if activity.ModuleID != nil && activity.ExerciseID != nil {
			if exercise, ok := s.catalog.ExerciseByCourse(courseID, *activity.ModuleID, *activity.ExerciseID); ok {
				activity.ExerciseTitle = stringPointer(exercise.Title)
			}
		}
		if activity.ModuleID != nil && activity.VideoID != nil {
			if module, ok := s.catalog.ModuleByCourse(courseID, *activity.ModuleID); ok {
				for _, video := range module.Videos {
					if video.ID == *activity.VideoID {
						activity.VideoTitle = stringPointer(video.Title)
						break
					}
				}
			}
		}
		activities = append(activities, activity)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate learner activities: %w", err)
	}
	return activities, nil
}

func formatLearnerTime(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05.000000000Z")
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
