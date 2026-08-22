package tutorlearning

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/learner"
	"github.com/johncrowleydev/fonzytooter/server/internal/review"
	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
)

type Services struct {
	Catalog *curriculum.Catalog
	Learner *learner.Service
	Review  *review.Service
}

func NewTools(services Services) ([]tutor.Tool, error) {
	if services.Catalog == nil || services.Learner == nil || services.Review == nil {
		return nil, errors.New("tutor learning tools require curriculum, learner, and review services")
	}
	constructors := []func() (tutor.Tool, error){
		func() (tutor.Tool, error) { return newSearchTool(services.Catalog) },
		func() (tutor.Tool, error) { return newContentTool(services.Catalog) },
		func() (tutor.Tool, error) { return newObjectiveTool(services.Catalog, services.Learner) },
		func() (tutor.Tool, error) { return newActivityTool(services.Catalog, services.Learner) },
		func() (tutor.Tool, error) { return newExerciseHistoryTool(services.Catalog, services.Learner) },
		func() (tutor.Tool, error) { return newReviewHistoryTool(services.Catalog, services.Review) },
	}
	tools := make([]tutor.Tool, 0, len(constructors))
	for _, construct := range constructors {
		tool, err := construct()
		if err != nil {
			return nil, err
		}
		tools = append(tools, tool)
	}
	return tools, nil
}

type SearchCurriculumArgs struct {
	CourseID string `json:"courseId" minLength:"1" maxLength:"200"`
	Query    string `json:"query" minLength:"1" maxLength:"500"`
	Limit    int    `json:"limit,omitempty" minimum:"1" maximum:"20"`
}

type SearchResult struct {
	Kind         string   `json:"kind"`
	CourseID     string   `json:"courseId"`
	ModuleID     string   `json:"moduleId,omitempty"`
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Excerpt      string   `json:"excerpt,omitempty"`
	ObjectiveIDs []string `json:"objectiveIds,omitempty"`
	SourceIDs    []string `json:"sourceIds,omitempty"`
	Score        int      `json:"score"`
}

type SearchCurriculumResult struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
}

type searchCandidate struct {
	result SearchResult
	text   string
}

func newSearchTool(catalog *curriculum.Catalog) (tutor.Tool, error) {
	return tutor.NewTypedToolWithProvenance(
		ToolSearchCurriculum,
		"Search authored curriculum within one course using deterministic text ranking. Use when fresh page context is insufficient.",
		func(args SearchCurriculumArgs) error {
			if strings.TrimSpace(args.CourseID) == "" || strings.TrimSpace(args.Query) == "" {
				return errors.New("courseId and query are required")
			}
			if normalize(args.Query) == "" {
				return errors.New("query must contain a letter or number")
			}
			return nil
		},
		func(_ context.Context, args SearchCurriculumArgs) (SearchCurriculumResult, error) {
			course, ok := catalog.CourseByID(args.CourseID)
			if !ok {
				return SearchCurriculumResult{}, fmt.Errorf("course %q not found", args.CourseID)
			}
			limit := args.Limit
			if limit <= 0 {
				limit = 8
			}
			candidates := make([]searchCandidate, 0)
			for _, module := range course.Modules {
				candidates = append(candidates, searchCandidate{result: SearchResult{Kind: "module", CourseID: course.ID, ModuleID: module.ID, ID: module.ID, Title: module.Title}, text: module.Title})
				for _, objective := range module.Objectives {
					candidates = append(candidates, searchCandidate{result: SearchResult{Kind: "objective", CourseID: course.ID, ModuleID: module.ID, ID: objective.ID, Title: objective.Title, Excerpt: truncate(objective.Description, 300)}, text: objective.Title + " " + objective.Description})
				}
				for _, lesson := range module.Lessons {
					candidates = append(candidates, searchCandidate{result: SearchResult{Kind: "lesson", CourseID: course.ID, ModuleID: module.ID, ID: lesson.ID, Title: lesson.Title, Excerpt: truncate(lesson.Content, 400), ObjectiveIDs: lesson.ObjectiveIDs, SourceIDs: lesson.SourceIDs}, text: lesson.Title + " " + lesson.Content})
				}
				for _, exercise := range module.Exercises {
					candidates = append(candidates, searchCandidate{result: SearchResult{Kind: "exercise", CourseID: course.ID, ModuleID: module.ID, ID: exercise.ID, Title: exercise.Title, Excerpt: truncate(exercise.Prompt, 300), ObjectiveIDs: exercise.ObjectiveIDs}, text: exercise.Title + " " + exercise.Prompt})
				}
				for _, item := range module.ReviewItems {
					candidates = append(candidates, searchCandidate{result: SearchResult{Kind: "review_item", CourseID: course.ID, ModuleID: module.ID, ID: item.ID, Title: item.Prompt, ObjectiveIDs: item.ObjectiveIDs}, text: item.Prompt})
				}
			}
			query := normalize(args.Query)
			tokens := strings.Fields(query)
			results := make([]SearchResult, 0)
			for _, candidate := range candidates {
				candidate.result.Score = searchScore(query, tokens, candidate.result.ID, candidate.result.Title, candidate.text)
				if candidate.result.Score > 0 {
					results = append(results, candidate.result)
				}
			}
			sort.Slice(results, func(i, j int) bool {
				if results[i].Score != results[j].Score {
					return results[i].Score > results[j].Score
				}
				if results[i].Kind != results[j].Kind {
					return results[i].Kind < results[j].Kind
				}
				if results[i].ModuleID != results[j].ModuleID {
					return results[i].ModuleID < results[j].ModuleID
				}
				return results[i].ID < results[j].ID
			})
			if len(results) > limit {
				results = results[:limit]
			}
			return SearchCurriculumResult{Query: args.Query, Results: results}, nil
		},
		func(result SearchCurriculumResult) []string {
			var ids []string
			for _, item := range result.Results {
				ids = append(ids, item.SourceIDs...)
			}
			return ids
		},
	)
}

func normalize(value string) string {
	return strings.Join(strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}), " ")
}

func searchScore(query string, tokens []string, id, title, body string) int {
	normalizedID, normalizedTitle, normalizedBody := normalize(id), normalize(title), normalize(body)
	score := 0
	if normalizedID == query {
		score += 40
	} else if strings.Contains(normalizedID, query) {
		score += 18
	}
	if normalizedTitle == query {
		score += 35
	} else if strings.Contains(normalizedTitle, query) {
		score += 20
	}
	for _, token := range tokens {
		if strings.Contains(normalizedTitle, token) {
			score += 6
		}
		if strings.Contains(normalizedBody, token) {
			score += 2
		}
	}
	return score
}

type GetCurriculumContentArgs struct {
	CourseID string `json:"courseId" minLength:"1" maxLength:"200"`
	ModuleID string `json:"moduleId,omitempty" maxLength:"200"`
	Kind     string `json:"kind" enum:"lesson,objective,exercise,source"`
	ID       string `json:"id" minLength:"1" maxLength:"200"`
	MaxChars int    `json:"maxChars,omitempty" minimum:"500" maximum:"8000"`
}

type CurriculumContentResult struct {
	Kind         string          `json:"kind"`
	CourseID     string          `json:"courseId"`
	ModuleID     string          `json:"moduleId,omitempty"`
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	Content      string          `json:"content"`
	ObjectiveIDs []string        `json:"objectiveIds,omitempty"`
	Sources      []sourceSummary `json:"sources,omitempty"`
	Truncated    bool            `json:"truncated"`
}

func newContentTool(catalog *curriculum.Catalog) (tutor.Tool, error) {
	return tutor.NewTypedToolWithProvenance(
		ToolGetCurriculumContent,
		"Retrieve bounded authoritative lesson, objective, exercise, or source content by stable course-qualified reference. Never returns worksheet answers or hidden tests.",
		func(args GetCurriculumContentArgs) error {
			if args.CourseID == "" || args.ID == "" {
				return errors.New("courseId and id are required")
			}
			if args.Kind != "source" && args.ModuleID == "" {
				return errors.New("moduleId is required for this curriculum reference")
			}
			return nil
		},
		func(_ context.Context, args GetCurriculumContentArgs) (CurriculumContentResult, error) {
			course, ok := catalog.CourseByID(args.CourseID)
			if !ok {
				return CurriculumContentResult{}, fmt.Errorf("course %q not found", args.CourseID)
			}
			maxChars := args.MaxChars
			if maxChars <= 0 {
				maxChars = 6_000
			}
			result := CurriculumContentResult{Kind: args.Kind, CourseID: course.ID, ModuleID: args.ModuleID, ID: args.ID}
			switch args.Kind {
			case "lesson":
				lesson, ok := catalog.LessonByCourse(course.ID, args.ModuleID, args.ID)
				if !ok {
					return result, fmt.Errorf("lesson %q not found in course %q module %q", args.ID, course.ID, args.ModuleID)
				}
				result.Title, result.Content, result.ObjectiveIDs = lesson.Title, lesson.Content, lesson.ObjectiveIDs
				for _, sourceID := range lesson.SourceIDs {
					if source, exists := catalog.SourceByID(sourceID); exists {
						result.Sources = append(result.Sources, sourceSummary{ID: source.ID, Title: source.Title, URL: source.URL})
					}
				}
			case "objective":
				objective, ok := catalog.ObjectiveByID(args.ID)
				if !ok || objective.CourseID != course.ID || objective.ModuleID != args.ModuleID {
					return result, fmt.Errorf("objective %q not found in course %q module %q", args.ID, course.ID, args.ModuleID)
				}
				result.Title, result.Content = objective.Title, objective.Description
			case "exercise":
				exercise, ok := catalog.ExerciseByCourse(course.ID, args.ModuleID, args.ID)
				if !ok {
					return result, fmt.Errorf("exercise %q not found in course %q module %q", args.ID, course.ID, args.ModuleID)
				}
				result.Title, result.Content, result.ObjectiveIDs = exercise.Title, exercise.Prompt, exercise.ObjectiveIDs
			case "source":
				source, exists := catalog.SourceByID(args.ID)
				if !exists || !courseUsesSource(course, source.ID) {
					return result, fmt.Errorf("source %q is not referenced by course %q", args.ID, course.ID)
				}
				result.ModuleID, result.Title, result.Content = "", source.Title, source.URL
				result.Sources = []sourceSummary{{ID: source.ID, Title: source.Title, URL: source.URL}}
			default:
				return result, fmt.Errorf("unsupported curriculum content kind %q", args.Kind)
			}
			runes := []rune(result.Content)
			if len(runes) > maxChars {
				result.Content, result.Truncated = string(runes[:maxChars])+"…", true
			}
			return result, nil
		},
		func(result CurriculumContentResult) []string {
			ids := make([]string, 0, len(result.Sources))
			for _, source := range result.Sources {
				ids = append(ids, source.ID)
			}
			return ids
		},
	)
}

func courseUsesSource(course curriculum.Course, sourceID string) bool {
	for _, module := range course.Modules {
		for _, lesson := range module.Lessons {
			for _, candidate := range lesson.SourceIDs {
				if candidate == sourceID {
					return true
				}
			}
		}
	}
	return false
}

type GetObjectiveStateArgs struct {
	CourseID     string   `json:"courseId" minLength:"1" maxLength:"200"`
	ObjectiveIDs []string `json:"objectiveIds" minItems:"1" maxItems:"10"`
}

type ObjectiveState struct {
	Progress      learner.ObjectiveProgress   `json:"progress"`
	Prerequisites []learner.ObjectiveProgress `json:"prerequisites"`
}

type ObjectiveStateResult struct {
	CourseID   string           `json:"courseId"`
	Objectives []ObjectiveState `json:"objectives"`
}

func newObjectiveTool(catalog *curriculum.Catalog, learnerService *learner.Service) (tutor.Tool, error) {
	return tutor.NewTypedTool(
		ToolGetObjectiveState,
		"Return factual introduced, recall, application, and prerequisite evidence for course-owned learning objectives. Does not infer mastery.",
		func(args GetObjectiveStateArgs) error {
			if args.CourseID == "" || len(args.ObjectiveIDs) == 0 {
				return errors.New("courseId and objectiveIds are required")
			}
			return nil
		},
		func(ctx context.Context, args GetObjectiveStateArgs) (ObjectiveStateResult, error) {
			progress, err := learnerService.CourseProgress(ctx, args.CourseID)
			if err != nil {
				return ObjectiveStateResult{}, err
			}
			byID := make(map[string]learner.ObjectiveProgress, len(progress.Objectives))
			for _, state := range progress.Objectives {
				byID[state.ID] = state
			}
			result := ObjectiveStateResult{CourseID: args.CourseID}
			for _, id := range uniqueStrings(args.ObjectiveIDs) {
				objective, ok := catalog.ObjectiveByID(id)
				state, hasState := byID[id]
				if !ok || objective.CourseID != args.CourseID || !hasState {
					return ObjectiveStateResult{}, fmt.Errorf("objective %q not found in course %q", id, args.CourseID)
				}
				item := ObjectiveState{Progress: state}
				for _, prerequisiteID := range objective.Prerequisites {
					if prerequisite, exists := byID[prerequisiteID]; exists {
						item.Prerequisites = append(item.Prerequisites, prerequisite)
					}
				}
				result.Objectives = append(result.Objectives, item)
			}
			return result, nil
		},
	)
}

type GetRecentActivityArgs struct {
	CourseID string `json:"courseId" minLength:"1" maxLength:"200"`
	Limit    int    `json:"limit,omitempty" minimum:"1" maximum:"30"`
}

type RecentActivityResult struct {
	CourseID   string             `json:"courseId"`
	Activities []learner.Activity `json:"activities"`
}

func newActivityTool(catalog *curriculum.Catalog, learnerService *learner.Service) (tutor.Tool, error) {
	return tutor.NewTypedTool(
		ToolGetRecentActivity,
		"Return a bounded newest-first window of factual lesson, exercise, and review activity for one course.",
		func(args GetRecentActivityArgs) error {
			if _, ok := catalog.CourseByID(args.CourseID); !ok {
				return fmt.Errorf("course %q not found", args.CourseID)
			}
			return nil
		},
		func(ctx context.Context, args GetRecentActivityArgs) (RecentActivityResult, error) {
			limit := args.Limit
			if limit <= 0 {
				limit = 10
			}
			activities, err := learnerService.Activities(ctx, args.CourseID, limit)
			return RecentActivityResult{CourseID: args.CourseID, Activities: activities}, err
		},
	)
}

type GetExerciseHistoryArgs struct {
	CourseID   string `json:"courseId" minLength:"1" maxLength:"200"`
	ModuleID   string `json:"moduleId" minLength:"1" maxLength:"200"`
	ExerciseID string `json:"exerciseId" minLength:"1" maxLength:"200"`
	Limit      int    `json:"limit,omitempty" minimum:"1" maximum:"20"`
}

type ExerciseHistoryResult struct {
	CourseID   string                    `json:"courseId"`
	ModuleID   string                    `json:"moduleId"`
	ExerciseID string                    `json:"exerciseId"`
	Title      string                    `json:"title"`
	Attempts   []learner.ExerciseAttempt `json:"attempts"`
}

func newExerciseHistoryTool(catalog *curriculum.Catalog, learnerService *learner.Service) (tutor.Tool, error) {
	return tutor.NewTypedTool(
		ToolGetExerciseHistory,
		"Return bounded saved check attempts and deterministic test failures for one course-qualified exercise. Never executes code.",
		nil,
		func(ctx context.Context, args GetExerciseHistoryArgs) (ExerciseHistoryResult, error) {
			exercise, ok := catalog.ExerciseByCourse(args.CourseID, args.ModuleID, args.ExerciseID)
			if !ok {
				return ExerciseHistoryResult{}, fmt.Errorf("exercise %q not found in course %q module %q", args.ExerciseID, args.CourseID, args.ModuleID)
			}
			limit := args.Limit
			if limit <= 0 {
				limit = 10
			}
			attempts, err := learnerService.ExerciseAttempts(ctx, args.CourseID, args.ModuleID, args.ExerciseID, limit)
			if err != nil {
				return ExerciseHistoryResult{}, err
			}
			visibleTests := make(map[string]struct{}, len(exercise.Tests))
			for _, test := range exercise.Tests {
				if test.Visibility == curriculum.ExerciseTestVisible {
					visibleTests[test.ID] = struct{}{}
				}
			}
			for index := range attempts {
				attempts[index].CodeSnapshot = truncate(attempts[index].CodeSnapshot, 4_000)
				visibleResults := make([]learner.ExerciseTestResult, 0, len(attempts[index].Results))
				for _, result := range attempts[index].Results {
					if _, visible := visibleTests[result.TestID]; !visible {
						continue
					}
					result.Message = truncate(result.Message, 1_000)
					visibleResults = append(visibleResults, result)
				}
				attempts[index].Results = visibleResults
			}
			return ExerciseHistoryResult{CourseID: args.CourseID, ModuleID: args.ModuleID, ExerciseID: args.ExerciseID, Title: exercise.Title, Attempts: attempts}, nil
		},
	)
}

type GetReviewHistoryArgs struct {
	CourseID     string   `json:"courseId" minLength:"1" maxLength:"200"`
	ModuleID     string   `json:"moduleId,omitempty" maxLength:"200"`
	ReviewItemID string   `json:"reviewItemId,omitempty" maxLength:"200"`
	ObjectiveIDs []string `json:"objectiveIds,omitempty" maxItems:"10"`
	Limit        int      `json:"limit,omitempty" minimum:"1" maximum:"20"`
}

type ReviewHistoryEntry struct {
	ItemTitle string              `json:"itemTitle"`
	Entry     review.HistoryEntry `json:"entry"`
}

type ReviewHistoryResult struct {
	CourseID string               `json:"courseId"`
	History  []ReviewHistoryEntry `json:"history"`
}

func newReviewHistoryTool(catalog *curriculum.Catalog, reviewService *review.Service) (tutor.Tool, error) {
	return tutor.NewTypedTool(
		ToolGetReviewHistory,
		"Return bounded factual FSRS review ratings and schedule evidence for course-owned review items or objectives. Does not invent mastery percentages.",
		func(args GetReviewHistoryArgs) error {
			if args.CourseID == "" || args.ReviewItemID == "" && len(args.ObjectiveIDs) == 0 {
				return errors.New("courseId and either reviewItemId or objectiveIds are required")
			}
			if args.ReviewItemID != "" && args.ModuleID == "" {
				return errors.New("moduleId is required with reviewItemId")
			}
			return nil
		},
		func(ctx context.Context, args GetReviewHistoryArgs) (ReviewHistoryResult, error) {
			if _, ok := catalog.CourseByID(args.CourseID); !ok {
				return ReviewHistoryResult{}, fmt.Errorf("course %q not found", args.CourseID)
			}
			items := make(map[string]curriculum.ReviewItem)
			if args.ReviewItemID != "" {
				item, ok := catalog.ReviewItemByCourse(args.CourseID, args.ModuleID, args.ReviewItemID)
				if !ok {
					return ReviewHistoryResult{}, fmt.Errorf("review item %q not found in course %q module %q", args.ReviewItemID, args.CourseID, args.ModuleID)
				}
				items[item.ModuleID+"\x00"+item.ID] = item
			}
			for _, objectiveID := range uniqueStrings(args.ObjectiveIDs) {
				objective, ok := catalog.ObjectiveByID(objectiveID)
				if !ok || objective.CourseID != args.CourseID {
					return ReviewHistoryResult{}, fmt.Errorf("objective %q not found in course %q", objectiveID, args.CourseID)
				}
				for _, item := range catalog.ReviewItemsByObjective(objectiveID) {
					if item.CourseID == args.CourseID {
						items[item.ModuleID+"\x00"+item.ID] = item
					}
				}
			}
			limit := args.Limit
			if limit <= 0 {
				limit = 10
			}
			result := ReviewHistoryResult{CourseID: args.CourseID}
			for _, item := range items {
				entries, err := reviewService.History(ctx, args.CourseID, item.ModuleID, item.ID, limit)
				if err != nil {
					return ReviewHistoryResult{}, err
				}
				for _, entry := range entries {
					result.History = append(result.History, ReviewHistoryEntry{ItemTitle: item.Prompt, Entry: entry})
				}
			}
			sort.Slice(result.History, func(i, j int) bool {
				left, right := result.History[i].Entry, result.History[j].Entry
				if !left.ReviewedAt.Equal(right.ReviewedAt) {
					return left.ReviewedAt.After(right.ReviewedAt)
				}
				return left.ID > right.ID
			})
			if len(result.History) > limit {
				result.History = result.History[:limit]
			}
			return result, nil
		},
	)
}
