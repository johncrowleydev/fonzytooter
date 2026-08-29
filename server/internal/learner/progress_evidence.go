package learner

import (
	"context"
	"fmt"
	"time"

	"github.com/johncrowleydev/helix-academy/server/internal/auth"
	"github.com/johncrowleydev/helix-academy/server/internal/curriculum"
)

func (s *Service) addCourseEvidence(ctx context.Context, userID auth.UserID, course curriculum.Course, objectiveIndex map[string]int, result *CourseProgress) error {
	now := s.now().UTC()
	reviewItems := make(map[string]curriculum.ReviewItem)
	exercises := make(map[string]curriculum.Exercise)
	for _, module := range course.Modules {
		for _, item := range module.ReviewItems {
			reviewItems[evidenceKey(module.ID, item.ID)] = item
			for _, objectiveID := range item.ObjectiveIDs {
				if index, ok := objectiveIndex[objectiveID]; ok {
					result.Objectives[index].Recall.ReviewItemCount++
				}
			}
		}
		for _, exercise := range module.Exercises {
			exercises[evidenceKey(module.ID, exercise.ID)] = exercise
			for _, objectiveID := range exercise.ObjectiveIDs {
				if index, ok := objectiveIndex[objectiveID]; ok {
					result.Objectives[index].Application.ExerciseCount++
				}
			}
		}
	}

	storedDue, err := s.reviewDueTimes(ctx, userID, course.ID)
	if err != nil {
		return err
	}
	eligibility, err := LoadSourceLessonEligibility(ctx, s.db, userID, course.ID)
	if err != nil {
		return err
	}
	for key, item := range reviewItems {
		dueAt, stored := storedDue[key]
		if !stored {
			if !eligibility.Allows(item) {
				continue
			}
			dueAt = now
		}
		isDue := !dueAt.After(now)
		if isDue {
			result.DueReviewCount++
		}
		for _, objectiveID := range item.ObjectiveIDs {
			index, ok := objectiveIndex[objectiveID]
			if !ok {
				continue
			}
			recall := &result.Objectives[index].Recall
			if isDue {
				recall.DueReviewCount++
			}
			setEarlier(&recall.NextDueAt, dueAt)
		}
	}

	if err := s.addReviewHistory(ctx, userID, course.ID, reviewItems, objectiveIndex, result); err != nil {
		return err
	}
	passedExercises, err := s.addExerciseHistory(ctx, userID, course.ID, exercises, objectiveIndex, result)
	if err != nil {
		return err
	}
	for _, module := range course.Modules {
		for _, exercise := range module.Exercises {
			if passedExercises[evidenceKey(module.ID, exercise.ID)] || !hasIntroducedObjective(exercise.ObjectiveIDs, objectiveIndex, result.Objectives) {
				continue
			}
			result.PracticeExercise = &PracticeExercise{
				CourseID: course.ID, ModuleID: module.ID, ModuleTitle: module.Title,
				ExerciseID: exercise.ID, ExerciseTitle: exercise.Title,
			}
			return nil
		}
	}
	return nil
}

func (s *Service) reviewDueTimes(ctx context.Context, userID auth.UserID, courseID string) (map[string]time.Time, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT module_id, review_item_id, due_at FROM review_cards WHERE user_id = ? AND course_id = ?`, userID, courseID)
	if err != nil {
		return nil, fmt.Errorf("read review evidence: %w", err)
	}
	defer rows.Close()
	result := make(map[string]time.Time)
	for rows.Next() {
		var moduleID, itemID, dueText string
		if err := rows.Scan(&moduleID, &itemID, &dueText); err != nil {
			return nil, fmt.Errorf("scan review evidence: %w", err)
		}
		dueAt, err := time.Parse(time.RFC3339Nano, dueText)
		if err != nil {
			return nil, fmt.Errorf("parse review due time: %w", err)
		}
		result[evidenceKey(moduleID, itemID)] = dueAt.UTC()
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate review evidence: %w", err)
	}
	return result, nil
}

func (s *Service) addReviewHistory(ctx context.Context, userID auth.UserID, courseID string, items map[string]curriculum.ReviewItem, objectiveIndex map[string]int, result *CourseProgress) error {
	rows, err := s.db.QueryContext(ctx, `SELECT module_id, review_item_id, reviewed_at FROM review_logs WHERE user_id = ? AND course_id = ?`, userID, courseID)
	if err != nil {
		return fmt.Errorf("read review history evidence: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var moduleID, itemID, reviewedText string
		if err := rows.Scan(&moduleID, &itemID, &reviewedText); err != nil {
			return fmt.Errorf("scan review history evidence: %w", err)
		}
		item, ok := items[evidenceKey(moduleID, itemID)]
		if !ok {
			continue
		}
		reviewedAt, err := time.Parse(time.RFC3339Nano, reviewedText)
		if err != nil {
			return fmt.Errorf("parse review history time: %w", err)
		}
		for _, objectiveID := range item.ObjectiveIDs {
			if index, ok := objectiveIndex[objectiveID]; ok {
				recall := &result.Objectives[index].Recall
				recall.ReviewsCompleted++
				setLater(&recall.LastReviewedAt, reviewedAt.UTC())
			}
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate review history evidence: %w", err)
	}
	return nil
}

func (s *Service) addExerciseHistory(ctx context.Context, userID auth.UserID, courseID string, exercises map[string]curriculum.Exercise, objectiveIndex map[string]int, result *CourseProgress) (map[string]bool, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT module_id, exercise_id, created_at, all_passed FROM exercise_attempts WHERE user_id = ? AND course_id = ?`, userID, courseID)
	if err != nil {
		return nil, fmt.Errorf("read exercise evidence: %w", err)
	}
	defer rows.Close()
	passedExercises := make(map[string]bool)
	for rows.Next() {
		var moduleID, exerciseID, checkedText string
		var allPassed bool
		if err := rows.Scan(&moduleID, &exerciseID, &checkedText, &allPassed); err != nil {
			return nil, fmt.Errorf("scan exercise evidence: %w", err)
		}
		key := evidenceKey(moduleID, exerciseID)
		exercise, ok := exercises[key]
		if !ok {
			continue
		}
		checkedAt, err := time.Parse(time.RFC3339Nano, checkedText)
		if err != nil {
			return nil, fmt.Errorf("parse exercise evidence time: %w", err)
		}
		if allPassed {
			passedExercises[key] = true
		}
		for _, objectiveID := range exercise.ObjectiveIDs {
			if index, ok := objectiveIndex[objectiveID]; ok {
				application := &result.Objectives[index].Application
				application.Attempts++
				if allPassed {
					application.FullyPassedAttempts++
				}
				setLater(&application.LastCheckedAt, checkedAt.UTC())
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exercise evidence: %w", err)
	}
	return passedExercises, nil
}

func hasIntroducedObjective(ids []string, objectiveIndex map[string]int, objectives []ObjectiveProgress) bool {
	for _, id := range ids {
		if index, ok := objectiveIndex[id]; ok && objectives[index].Introduced {
			return true
		}
	}
	return false
}

func setEarlier(target **time.Time, value time.Time) {
	if *target == nil || value.Before(**target) {
		copy := value
		*target = &copy
	}
}

func setLater(target **time.Time, value time.Time) {
	if *target == nil || value.After(**target) {
		copy := value
		*target = &copy
	}
}

func evidenceKey(moduleID, itemID string) string { return moduleID + "\x00" + itemID }
