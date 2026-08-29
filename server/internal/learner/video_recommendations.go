package learner

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/johncrowleydev/helix-academy/server/internal/auth"
	"github.com/johncrowleydev/helix-academy/server/internal/curriculum"
)

const (
	VideoRecommendationCurrentLesson     = "current_lesson"
	VideoRecommendationNextPrerequisite  = "next_prerequisite"
	VideoRecommendationWeakEvidence      = "weak_evidence"
	VideoRecommendationRevisit           = "revisit"
	VideoRecommendationCurrentModule     = "current_module"
	MaxVideoRecommendations              = 3
	recentExerciseDifficultyWindow       = 30 * 24 * time.Hour
	minimumDifficultExerciseAttemptCount = 2
)

type VideoRecommendation struct {
	Video       curriculum.Video
	ReasonKind  string
	Reason      string
	Watched     bool
	LessonID    *string
	LessonTitle *string
}

type rankedVideoRecommendation struct {
	recommendation VideoRecommendation
	priority       int
	moduleOrder    int
	videoOrder     int
}

func (s *Service) VideoRecommendations(ctx context.Context, userID auth.UserID, courseID string, limit int) ([]VideoRecommendation, error) {
	course, ok := s.catalog.CourseByID(courseID)
	if !ok {
		return nil, ErrCourseNotFound
	}
	if limit <= 0 || limit > MaxVideoRecommendations {
		limit = MaxVideoRecommendations
	}

	progress, err := s.CourseProgress(ctx, userID, courseID)
	if err != nil {
		return nil, fmt.Errorf("derive video recommendation context: %w", err)
	}
	watched, err := s.completedVideoKeys(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}

	objectives := make(map[string]curriculum.Objective)
	for _, module := range course.Modules {
		for _, objective := range module.Objectives {
			objectives[objective.ID] = objective
		}
	}
	weakObjectives := recentDifficultObjectives(progress.Objectives, s.now().UTC())

	var nextLesson curriculum.Lesson
	var hasNextLesson bool
	prerequisiteIDs := make(map[string]bool)
	if progress.NextLesson != nil {
		nextLesson, hasNextLesson = s.catalog.LessonByCourse(courseID, progress.NextLesson.ModuleID, progress.NextLesson.LessonID)
		if hasNextLesson {
			for _, objectiveID := range nextLesson.ObjectiveIDs {
				if objective, exists := objectives[objectiveID]; exists {
					for _, prerequisiteID := range objective.Prerequisites {
						prerequisiteIDs[prerequisiteID] = true
					}
				}
			}
		}
	}

	candidates := make([]rankedVideoRecommendation, 0)
	for _, module := range course.Modules {
		for _, video := range module.Videos {
			isWatched := watched[videoKey(module.ID, video.ID)]
			priority, kind, reason, include := rankVideo(video, module, progress.NextLesson, nextLesson, hasNextLesson, prerequisiteIDs, weakObjectives, objectives, isWatched)
			if !include {
				continue
			}
			lessonID, lessonTitle := recommendationLesson(video, module, progress.NextLesson)
			candidates = append(candidates, rankedVideoRecommendation{
				recommendation: VideoRecommendation{
					Video: video, ReasonKind: kind, Reason: reason, Watched: isWatched,
					LessonID: lessonID, LessonTitle: lessonTitle,
				},
				priority: priority, moduleOrder: module.Order, videoOrder: video.Order,
			})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		left, right := candidates[i], candidates[j]
		if left.priority != right.priority {
			return left.priority < right.priority
		}
		if left.moduleOrder != right.moduleOrder {
			return left.moduleOrder < right.moduleOrder
		}
		if left.videoOrder != right.videoOrder {
			return left.videoOrder < right.videoOrder
		}
		return left.recommendation.Video.ID < right.recommendation.Video.ID
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	result := make([]VideoRecommendation, 0, len(candidates))
	for _, candidate := range candidates {
		result = append(result, candidate.recommendation)
	}
	return result, nil
}

func (s *Service) completedVideoKeys(ctx context.Context, userID auth.UserID, courseID string) (map[string]bool, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT module_id, video_id FROM video_progress
		WHERE user_id = ? AND course_id = ? AND completed = 1
	`, userID, courseID)
	if err != nil {
		return nil, fmt.Errorf("read completed videos for recommendations: %w", err)
	}
	defer rows.Close()
	result := make(map[string]bool)
	for rows.Next() {
		var moduleID, videoID string
		if err := rows.Scan(&moduleID, &videoID); err != nil {
			return nil, fmt.Errorf("scan completed video for recommendations: %w", err)
		}
		result[videoKey(moduleID, videoID)] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate completed videos for recommendations: %w", err)
	}
	return result, nil
}

func recentDifficultObjectives(objectives []ObjectiveProgress, now time.Time) map[string]ObjectiveProgress {
	result := make(map[string]ObjectiveProgress)
	for _, objective := range objectives {
		application := objective.Application
		if application.Attempts < minimumDifficultExerciseAttemptCount || application.FullyPassedAttempts != 0 || application.LastCheckedAt == nil {
			continue
		}
		age := now.Sub(application.LastCheckedAt.UTC())
		if age >= 0 && age <= recentExerciseDifficultyWindow {
			result[objective.ID] = objective
		}
	}
	return result
}

func rankVideo(video curriculum.Video, module curriculum.Module, next *NextLesson, nextLesson curriculum.Lesson, hasNextLesson bool, prerequisites map[string]bool, weak map[string]ObjectiveProgress, objectives map[string]curriculum.Objective, watched bool) (int, string, string, bool) {
	if !watched && hasNextLesson && module.ID == next.ModuleID && containsString(video.LessonIDs, nextLesson.ID) {
		return 0, VideoRecommendationCurrentLesson, fmt.Sprintf("Worth watching next — supports %s.", nextLesson.Title), true
	}
	if !watched {
		if objectiveID, ok := firstMatchingObjective(video.ObjectiveIDs, prerequisites); ok {
			return 1, VideoRecommendationNextPrerequisite, fmt.Sprintf("Prepare visually — supports the prerequisite %s.", objectives[objectiveID].Title), true
		}
	}
	if objectiveID, ok := firstMatchingWeakObjective(video.ObjectiveIDs, weak); ok {
		objective := weak[objectiveID]
		if watched {
			return 3, VideoRecommendationRevisit, fmt.Sprintf("Review this visually — recent practice on %s has been difficult.", objective.Title), true
		}
		return 2, VideoRecommendationWeakEvidence, fmt.Sprintf("Practice support — recent practice on %s has been difficult.", objective.Title), true
	}
	if !watched && next != nil && module.ID == next.ModuleID {
		return 4, VideoRecommendationCurrentModule, fmt.Sprintf("More from %s — an unwatched explanation in your current module.", module.Title), true
	}
	return 0, "", "", false
}

func recommendationLesson(video curriculum.Video, module curriculum.Module, next *NextLesson) (*string, *string) {
	if next != nil && module.ID == next.ModuleID && containsString(video.LessonIDs, next.LessonID) {
		return stringPointer(next.LessonID), stringPointer(next.LessonTitle)
	}
	for _, lessonID := range video.LessonIDs {
		for _, lesson := range module.Lessons {
			if lesson.ID == lessonID {
				return stringPointer(lesson.ID), stringPointer(lesson.Title)
			}
		}
	}
	return nil, nil
}

func firstMatchingObjective(ids []string, matches map[string]bool) (string, bool) {
	for _, id := range ids {
		if matches[id] {
			return id, true
		}
	}
	return "", false
}

func firstMatchingWeakObjective(ids []string, matches map[string]ObjectiveProgress) (string, bool) {
	for _, id := range ids {
		if _, ok := matches[id]; ok {
			return id, true
		}
	}
	return "", false
}

func videoKey(moduleID, videoID string) string { return moduleID + "\x00" + videoID }

func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
