package curriculum

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var stableIDPattern = regexp.MustCompile(StableIDPattern)
var youtubeIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)
var youtubeVideoElementPattern = regexp.MustCompile(`<YouTubeVideo\b([^>]*)>`)
var youtubeVideoIDAttributePattern = regexp.MustCompile(`(?:^|\s)id\s*=\s*["']([^"']+)["']`)
var arbitraryEmbedElementPattern = regexp.MustCompile(`(?i)<\s*(iframe|embed|object)\b`)

type errorCollector struct {
	items []error
}

func (c *errorCollector) add(pathName string, err error) {
	if err == nil {
		return
	}
	c.items = append(c.items, formatPath(pathName, err.Error()))
}

func (c *errorCollector) addf(pathName, format string, args ...any) {
	c.add(pathName, fmt.Errorf(format, args...))
}

func (c *errorCollector) Err() error {
	if len(c.items) == 0 {
		return nil
	}
	sort.SliceStable(c.items, func(i, j int) bool { return c.items[i].Error() < c.items[j].Error() })
	lines := make([]string, len(c.items))
	for index, item := range c.items {
		lines[index] = item.Error()
	}
	return fmt.Errorf("curriculum validation failed:\n- %s", strings.Join(lines, "\n- "))
}

func validateSources(sources map[string]sourceAuthoring, errors *errorCollector) {
	ids := sortedKeys(sources)
	for _, id := range ids {
		source := sources[id]
		if !stableIDPattern.MatchString(id) {
			errors.addf("sources.yaml", "invalid source id %q", id)
		}
		if strings.TrimSpace(source.Title) == "" {
			errors.addf("sources.yaml", "source %q has empty title", id)
		}
		if err := validateHTTPURL(source.URL); err != nil {
			errors.addf("sources.yaml", "source %q: %v", id, err)
		}
	}
}

func validateCourses(courses []courseFile, sources map[string]sourceAuthoring, errors *errorCollector) {
	courseIDs := map[string]string{}
	courseOrders := map[int]string{}
	objectiveIDs := map[string]string{}
	objectivePaths := map[string]string{}
	modules := []moduleFile{}

	for _, course := range courses {
		moduleIDs := map[string]string{}
		if course.metadataOK {
			metadata := course.metadata
			if !stableIDPattern.MatchString(metadata.ID) {
				errors.addf(course.path, "invalid course id %q", metadata.ID)
			}
			if previous, ok := courseIDs[metadata.ID]; ok {
				errors.addf(course.path, "duplicate course id %q (already declared in %s)", metadata.ID, previous)
			} else {
				courseIDs[metadata.ID] = course.path
			}
			if strings.TrimSpace(metadata.Title) == "" {
				errors.addf(course.path, "course %q has empty title", metadata.ID)
			}
			if strings.TrimSpace(metadata.Description) == "" {
				errors.addf(course.path, "course %q has empty description", metadata.ID)
			}
			if metadata.Order == nil {
				errors.addf(course.path, "course %q is missing required order", metadata.ID)
			} else {
				if *metadata.Order < 0 {
					errors.addf(course.path, "course %q order must be non-negative", metadata.ID)
				}
				if previous, ok := courseOrders[*metadata.Order]; ok {
					errors.addf(course.path, "duplicate course order %d (already declared in %s)", *metadata.Order, previous)
				} else {
					courseOrders[*metadata.Order] = course.path
				}
			}
		}

		moduleOrders := map[int]string{}
		for _, module := range course.modules {
			modules = append(modules, module)
			if !module.metadataOK {
				continue
			}
			validateModule(module, moduleIDs, moduleOrders, objectiveIDs, objectivePaths, errors)
		}
	}

	for _, module := range modules {
		validateLessonFiles(module, objectiveIDs, sources, errors)
		validateWorksheetFiles(module, objectiveIDs, errors)
		validateExerciseFiles(module, objectiveIDs, errors)
		validateReviewItemFiles(module, objectiveIDs, errors)
	}
	validateObjectiveReferences(modules, objectiveIDs, errors)
	validatePrerequisiteCycles(modules, objectiveIDs, objectivePaths, errors)
}

func validateReviewItemFiles(module moduleFile, objectiveIDs map[string]string, errors *errorCollector) {
	lessonIDs := make(map[string]struct{}, len(module.lessons))
	for _, lesson := range module.lessons {
		if lesson.metadataOK {
			lessonIDs[lesson.metadata.ID] = struct{}{}
		}
	}

	reviewItemIDs := map[string]string{}
	for _, reviewItem := range module.reviewItems {
		if !reviewItem.metadataOK {
			continue
		}
		metadata := reviewItem.metadata
		if !stableIDPattern.MatchString(metadata.ID) {
			errors.addf(reviewItem.path, "invalid review item id %q", metadata.ID)
		}
		if previous, ok := reviewItemIDs[metadata.ID]; ok {
			errors.addf(reviewItem.path, "duplicate review item id %q (already declared in %s)", metadata.ID, previous)
		} else {
			reviewItemIDs[metadata.ID] = reviewItem.path
		}
		if metadata.Order == nil {
			errors.addf(reviewItem.path, "review item %q is missing required order", metadata.ID)
		} else if *metadata.Order < 0 {
			errors.addf(reviewItem.path, "review item %q order must be non-negative", metadata.ID)
		}
		validateReviewItemObjectiveIDs(reviewItem.path, metadata.ID, metadata.ObjectiveIDs, objectiveIDs, errors)
		if strings.TrimSpace(metadata.SourceLessonID) == "" {
			errors.addf(reviewItem.path, "review item %q has empty sourceLessonId", metadata.ID)
		} else if _, ok := lessonIDs[metadata.SourceLessonID]; !ok {
			errors.addf(reviewItem.path, "review item %q references unknown lesson id %q in this module", metadata.ID, metadata.SourceLessonID)
		}
		if strings.TrimSpace(metadata.Prompt) == "" {
			errors.addf(reviewItem.path, "review item %q has empty prompt", metadata.ID)
		}
		if strings.TrimSpace(metadata.Answer) == "" {
			errors.addf(reviewItem.path, "review item %q has empty answer", metadata.ID)
		}
	}
}

func validateReviewItemObjectiveIDs(pathName, reviewItemID string, ids []string, objectiveIDs map[string]string, errors *errorCollector) {
	if len(ids) == 0 {
		errors.addf(pathName, "review item %q has no objectiveIds", reviewItemID)
		return
	}
	validateUniqueStrings(pathName, fmt.Sprintf("review item %q", reviewItemID), "objectiveIds", ids, errors)
	for _, objectiveID := range ids {
		if _, ok := objectiveIDs[objectiveID]; !ok {
			errors.addf(pathName, "review item %q has unknown objective id %q", reviewItemID, objectiveID)
		}
	}
}

func validateExerciseFiles(module moduleFile, objectiveIDs map[string]string, errors *errorCollector) {
	lessonIDs := make(map[string]struct{}, len(module.lessons))
	for _, lesson := range module.lessons {
		if lesson.metadataOK {
			lessonIDs[lesson.metadata.ID] = struct{}{}
		}
	}

	exerciseIDs := map[string]string{}
	ordersByLesson := map[string]map[int]string{}
	for _, exercise := range module.exercises {
		if !exercise.metadataOK {
			continue
		}
		metadata := exercise.metadata
		if !stableIDPattern.MatchString(metadata.ID) {
			errors.addf(exercise.path, "invalid exercise id %q", metadata.ID)
		}
		if previous, ok := exerciseIDs[metadata.ID]; ok {
			errors.addf(exercise.path, "duplicate exercise id %q (already declared in %s)", metadata.ID, previous)
		} else {
			exerciseIDs[metadata.ID] = exercise.path
		}
		if strings.TrimSpace(metadata.Title) == "" {
			errors.addf(exercise.path, "exercise %q has empty title", metadata.ID)
		}
		if strings.TrimSpace(metadata.LessonID) == "" {
			errors.addf(exercise.path, "exercise %q has empty lessonId", metadata.ID)
		} else if _, ok := lessonIDs[metadata.LessonID]; !ok {
			errors.addf(exercise.path, "exercise %q references unknown lesson id %q in this module", metadata.ID, metadata.LessonID)
		}
		if metadata.Order == nil {
			errors.addf(exercise.path, "exercise %q is missing required order", metadata.ID)
		} else {
			if *metadata.Order < 0 {
				errors.addf(exercise.path, "exercise %q order must be non-negative", metadata.ID)
			}
			lessonOrders := ordersByLesson[metadata.LessonID]
			if lessonOrders == nil {
				lessonOrders = map[int]string{}
				ordersByLesson[metadata.LessonID] = lessonOrders
			}
			if previous, ok := lessonOrders[*metadata.Order]; ok {
				errors.addf(exercise.path, "duplicate exercise order %d for lesson %q (already declared in %s)", *metadata.Order, metadata.LessonID, previous)
			} else {
				lessonOrders[*metadata.Order] = exercise.path
			}
		}
		validateExerciseObjectiveIDs(exercise.path, fmt.Sprintf("exercise %q", metadata.ID), metadata.ObjectiveIDs, objectiveIDs, errors)
		if strings.TrimSpace(metadata.Prompt) == "" {
			errors.addf(exercise.path, "exercise %q has empty prompt", metadata.ID)
		}
		if strings.TrimSpace(metadata.StarterCode) == "" {
			errors.addf(exercise.path, "exercise %q has empty starterCode", metadata.ID)
		}
		if len(metadata.Tests) == 0 {
			errors.addf(exercise.path, "exercise %q has no tests", metadata.ID)
		}
		validateExerciseTests(exercise, errors)
	}
}

func validateExerciseObjectiveIDs(pathName, owner string, ids []string, objectiveIDs map[string]string, errors *errorCollector) {
	if len(ids) == 0 {
		errors.addf(pathName, "%s has no objectiveIds", owner)
		return
	}
	validateUniqueStrings(pathName, owner, "objectiveIds", ids, errors)
	for _, objectiveID := range ids {
		if _, ok := objectiveIDs[objectiveID]; !ok {
			errors.addf(pathName, "%s has unknown objective id %q", owner, objectiveID)
		}
	}
}

func validateExerciseTests(exercise exerciseFile, errors *errorCollector) {
	testIDs := map[string]struct{}{}
	for _, test := range exercise.metadata.Tests {
		if !stableIDPattern.MatchString(test.ID) {
			errors.addf(exercise.path, "exercise %q has invalid test id %q", exercise.metadata.ID, test.ID)
		}
		if _, ok := testIDs[test.ID]; ok {
			errors.addf(exercise.path, "exercise %q has duplicate test id %q", exercise.metadata.ID, test.ID)
		} else {
			testIDs[test.ID] = struct{}{}
		}
		testLabel := fmt.Sprintf("exercise %q test %q", exercise.metadata.ID, test.ID)
		if strings.TrimSpace(test.Title) == "" {
			errors.addf(exercise.path, "%s has empty title", testLabel)
		}
		if test.Visibility != ExerciseTestVisible && test.Visibility != ExerciseTestHidden {
			errors.addf(exercise.path, "%s has invalid visibility %q", testLabel, test.Visibility)
		}
		if strings.TrimSpace(test.Code) == "" {
			errors.addf(exercise.path, "%s has empty code", testLabel)
		}
	}
}

func validateWorksheetFiles(module moduleFile, objectiveIDs map[string]string, errors *errorCollector) {
	lessonIDs := make(map[string]struct{}, len(module.lessons))
	for _, lesson := range module.lessons {
		if lesson.metadataOK {
			lessonIDs[lesson.metadata.ID] = struct{}{}
		}
	}

	worksheetIDs := map[string]string{}
	ordersByLesson := map[string]map[int]string{}
	for _, worksheet := range module.worksheets {
		if !worksheet.metadataOK {
			continue
		}
		metadata := worksheet.metadata
		if !stableIDPattern.MatchString(metadata.ID) {
			errors.addf(worksheet.path, "invalid worksheet id %q", metadata.ID)
		}
		if previous, ok := worksheetIDs[metadata.ID]; ok {
			errors.addf(worksheet.path, "duplicate worksheet id %q (already declared in %s)", metadata.ID, previous)
		} else {
			worksheetIDs[metadata.ID] = worksheet.path
		}
		if strings.TrimSpace(metadata.Title) == "" {
			errors.addf(worksheet.path, "worksheet %q has empty title", metadata.ID)
		}
		if strings.TrimSpace(metadata.LessonID) == "" {
			errors.addf(worksheet.path, "worksheet %q has empty lessonId", metadata.ID)
		} else if _, ok := lessonIDs[metadata.LessonID]; !ok {
			errors.addf(worksheet.path, "worksheet %q references unknown lesson id %q in this module", metadata.ID, metadata.LessonID)
		}
		if metadata.Order == nil {
			errors.addf(worksheet.path, "worksheet %q is missing required order", metadata.ID)
		} else {
			if *metadata.Order < 0 {
				errors.addf(worksheet.path, "worksheet %q order must be non-negative", metadata.ID)
			}
			lessonOrders := ordersByLesson[metadata.LessonID]
			if lessonOrders == nil {
				lessonOrders = map[int]string{}
				ordersByLesson[metadata.LessonID] = lessonOrders
			}
			if previous, ok := lessonOrders[*metadata.Order]; ok {
				errors.addf(worksheet.path, "duplicate worksheet order %d for lesson %q (already declared in %s)", *metadata.Order, metadata.LessonID, previous)
			} else {
				lessonOrders[*metadata.Order] = worksheet.path
			}
		}
		validateWorksheetObjectiveIDs(worksheet.path, fmt.Sprintf("worksheet %q", metadata.ID), metadata.ObjectiveIDs, objectiveIDs, errors)
		if strings.TrimSpace(metadata.Instructions) == "" {
			errors.addf(worksheet.path, "worksheet %q has empty instructions", metadata.ID)
		}
		if len(metadata.Problems) == 0 {
			errors.addf(worksheet.path, "worksheet %q has no problems", metadata.ID)
		}
		validateWorksheetProblems(worksheet, objectiveIDs, errors)
	}
}

func validateWorksheetProblems(worksheet worksheetFile, objectiveIDs map[string]string, errors *errorCollector) {
	problemIDs := map[string]struct{}{}
	for _, problem := range worksheet.metadata.Problems {
		if !stableIDPattern.MatchString(problem.ID) {
			errors.addf(worksheet.path, "worksheet %q has invalid problem id %q", worksheet.metadata.ID, problem.ID)
		}
		if _, ok := problemIDs[problem.ID]; ok {
			errors.addf(worksheet.path, "worksheet %q has duplicate problem id %q", worksheet.metadata.ID, problem.ID)
		} else {
			problemIDs[problem.ID] = struct{}{}
		}
		problemLabel := fmt.Sprintf("worksheet %q problem %q", worksheet.metadata.ID, problem.ID)
		if strings.TrimSpace(problem.Prompt) == "" {
			errors.addf(worksheet.path, "%s has empty prompt", problemLabel)
		}
		validateWorksheetObjectiveIDs(worksheet.path, problemLabel, problem.ObjectiveIDs, objectiveIDs, errors)
		if strings.TrimSpace(problem.ExpectedAnswer) == "" {
			errors.addf(worksheet.path, "%s has empty expectedAnswer", problemLabel)
		}
		if problem.RequiresWork == nil {
			errors.addf(worksheet.path, "%s is missing required requiresWork", problemLabel)
		}
		if problem.ResponseLines == nil {
			errors.addf(worksheet.path, "%s is missing required responseLines", problemLabel)
		} else if *problem.ResponseLines <= 0 {
			errors.addf(worksheet.path, "%s responseLines must be greater than zero", problemLabel)
		}
		if len(problem.Rubric) == 0 {
			errors.addf(worksheet.path, "%s has an empty rubric", problemLabel)
		}
		for index, criterion := range problem.Rubric {
			if strings.TrimSpace(criterion) == "" {
				errors.addf(worksheet.path, "%s has empty rubric item %d", problemLabel, index+1)
			}
		}
	}
}

func validateWorksheetObjectiveIDs(pathName, owner string, ids []string, objectiveIDs map[string]string, errors *errorCollector) {
	if len(ids) == 0 {
		errors.addf(pathName, "%s has no objectiveIds", owner)
		return
	}
	validateUniqueStrings(pathName, owner, "objectiveIds", ids, errors)
	for _, objectiveID := range ids {
		if _, ok := objectiveIDs[objectiveID]; !ok {
			errors.addf(pathName, "%s has unknown objective id %q", owner, objectiveID)
		}
	}
}

func validateModule(module moduleFile, moduleIDs map[string]string, moduleOrders map[int]string, objectiveIDs, objectivePaths map[string]string, errors *errorCollector) {
	metadataPath := module.path
	metadata := module.metadata
	if !stableIDPattern.MatchString(metadata.ID) {
		errors.addf(metadataPath, "invalid module id %q", metadata.ID)
	}
	if previous, ok := moduleIDs[metadata.ID]; ok {
		errors.addf(metadataPath, "duplicate module id %q (already declared in %s)", metadata.ID, previous)
	} else {
		moduleIDs[metadata.ID] = metadataPath
	}
	if strings.TrimSpace(metadata.Title) == "" {
		errors.addf(metadataPath, "module %q has empty title", metadata.ID)
	}
	if metadata.Order == nil {
		errors.addf(metadataPath, "module %q is missing required order", metadata.ID)
	} else {
		if *metadata.Order < 0 {
			errors.addf(metadataPath, "module %q order must be non-negative", metadata.ID)
		}
		if previous, ok := moduleOrders[*metadata.Order]; ok {
			errors.addf(metadataPath, "duplicate module order %d (already declared in %s)", *metadata.Order, previous)
		} else {
			moduleOrders[*metadata.Order] = metadataPath
		}
	}

	validateModuleObjectives(metadata, metadataPath, objectiveIDs, objectivePaths, errors)
	validateModuleVideos(module, errors)
	validateModuleLessonReferences(module, errors)
}

func validateModuleObjectives(metadata moduleAuthoring, metadataPath string, objectiveIDs, objectivePaths map[string]string, errors *errorCollector) {
	for _, objective := range metadata.Objectives {
		if !stableIDPattern.MatchString(objective.ID) {
			errors.addf(metadataPath, "invalid objective id %q", objective.ID)
		}
		if strings.TrimSpace(objective.Title) == "" {
			errors.addf(metadataPath, "objective %q has empty title", objective.ID)
		}
		if strings.TrimSpace(objective.Description) == "" {
			errors.addf(metadataPath, "objective %q has empty description", objective.ID)
		}
		validateUniqueStrings(metadataPath, fmt.Sprintf("objective %q", objective.ID), "prerequisites", objective.Prerequisites, errors)
		if previous, ok := objectiveIDs[objective.ID]; ok {
			errors.addf(metadataPath, "duplicate objective id %q (already declared in %s)", objective.ID, previous)
		} else {
			objectiveIDs[objective.ID] = metadataPath
			objectivePaths[objective.ID] = metadataPath
		}
	}
}

func validateModuleVideos(module moduleFile, errors *errorCollector) {
	metadata := module.metadata
	metadataPath := module.path
	videoIDs := map[string]struct{}{}
	videoOrders := map[int]string{}
	objectiveIDs := make(map[string]struct{}, len(metadata.Objectives))
	for _, objective := range metadata.Objectives {
		objectiveIDs[objective.ID] = struct{}{}
	}
	lessonIDs := make(map[string]struct{}, len(module.lessons))
	for _, lesson := range module.lessons {
		if lesson.metadataOK {
			lessonIDs[lesson.metadata.ID] = struct{}{}
		}
	}
	for _, video := range metadata.Videos {
		if !stableIDPattern.MatchString(video.ID) {
			errors.addf(metadataPath, "invalid video id %q", video.ID)
		}
		if _, ok := videoIDs[video.ID]; ok {
			errors.addf(metadataPath, "duplicate video id %q", video.ID)
		} else {
			videoIDs[video.ID] = struct{}{}
		}
		if strings.TrimSpace(video.Title) == "" {
			errors.addf(metadataPath, "video %q has empty title", video.ID)
		}
		if !youtubeIDPattern.MatchString(video.YouTubeID) {
			errors.addf(metadataPath, "video %q has invalid youtubeId %q; expected an 11-character YouTube video ID", video.ID, video.YouTubeID)
		}
		if strings.TrimSpace(video.Channel) == "" {
			errors.addf(metadataPath, "video %q has empty channel", video.ID)
		}
		if video.DurationMinutes == nil {
			errors.addf(metadataPath, "video %q is missing required durationMinutes", video.ID)
		} else if *video.DurationMinutes <= 0 {
			errors.addf(metadataPath, "video %q durationMinutes must be greater than zero", video.ID)
		}
		if video.Order == nil {
			errors.addf(metadataPath, "video %q is missing required order", video.ID)
		} else if *video.Order < 0 {
			errors.addf(metadataPath, "video %q order must be non-negative", video.ID)
		} else if previous, ok := videoOrders[*video.Order]; ok {
			errors.addf(metadataPath, "duplicate video order %d for videos %q and %q", *video.Order, previous, video.ID)
		} else {
			videoOrders[*video.Order] = video.ID
		}
		if len(video.ObjectiveIDs) == 0 {
			errors.addf(metadataPath, "video %q has no objectiveIds", video.ID)
		}
		validateUniqueStrings(metadataPath, fmt.Sprintf("video %q", video.ID), "objectiveIds", video.ObjectiveIDs, errors)
		for _, objectiveID := range video.ObjectiveIDs {
			if _, ok := objectiveIDs[objectiveID]; !ok {
				errors.addf(metadataPath, "video %q has unknown objective id %q in this module", video.ID, objectiveID)
			}
		}
		validateUniqueStrings(metadataPath, fmt.Sprintf("video %q", video.ID), "lessonIds", video.LessonIDs, errors)
		for _, lessonID := range video.LessonIDs {
			if _, ok := lessonIDs[lessonID]; !ok {
				errors.addf(metadataPath, "video %q has unknown lesson id %q in this module", video.ID, lessonID)
			}
		}
	}
}

func validateModuleLessonReferences(module moduleFile, errors *errorCollector) {
	if !module.metadataOK {
		return
	}
	validateUniqueStrings(module.path, fmt.Sprintf("module %q", module.metadata.ID), "lessons", module.metadata.Lessons, errors)
	declared := map[string]struct{}{}
	for _, lessonID := range module.metadata.Lessons {
		declared[lessonID] = struct{}{}
	}

	discovered := map[string]string{}
	for _, lesson := range module.lessons {
		if !lesson.metadataOK {
			continue
		}
		if previous, ok := discovered[lesson.metadata.ID]; ok {
			errors.addf(lesson.path, "duplicate lesson id %q (already declared in %s)", lesson.metadata.ID, previous)
		} else {
			discovered[lesson.metadata.ID] = lesson.path
		}
	}
	for _, lessonID := range module.metadata.Lessons {
		if _, ok := discovered[lessonID]; !ok {
			errors.addf(module.path, "lesson id %q is declared but no matching MDX lesson was found", lessonID)
		}
	}
	for lessonID, lessonPath := range discovered {
		if _, ok := declared[lessonID]; !ok {
			errors.addf(lessonPath, "lesson id %q is not declared in %s", lessonID, module.path)
		}
	}
}

func validateLessonFiles(module moduleFile, objectiveIDs map[string]string, sources map[string]sourceAuthoring, errors *errorCollector) {
	videosByID := make(map[string]videoAuthoring, len(module.metadata.Videos))
	for _, video := range module.metadata.Videos {
		videosByID[video.ID] = video
	}
	for _, lesson := range module.lessons {
		if !lesson.metadataOK {
			continue
		}
		metadata := lesson.metadata
		if !stableIDPattern.MatchString(metadata.ID) {
			errors.addf(lesson.path, "invalid lesson id %q", metadata.ID)
		}
		if strings.TrimSpace(metadata.Title) == "" {
			errors.addf(lesson.path, "lesson %q has empty title", metadata.ID)
		}
		if strings.TrimSpace(lesson.content) == "" {
			errors.addf(lesson.path, "lesson %q has empty content body", metadata.ID)
		}
		validateUniqueStrings(lesson.path, fmt.Sprintf("lesson %q", metadata.ID), "objectiveIds", metadata.ObjectiveIDs, errors)
		validateUniqueStrings(lesson.path, fmt.Sprintf("lesson %q", metadata.ID), "sourceIds", metadata.SourceIDs, errors)
		for _, objectiveID := range metadata.ObjectiveIDs {
			if _, ok := objectiveIDs[objectiveID]; !ok {
				errors.addf(lesson.path, "unknown objective id %q", objectiveID)
			}
		}
		for _, sourceID := range metadata.SourceIDs {
			if _, ok := sources[sourceID]; !ok {
				errors.addf(lesson.path, "unknown source id %q", sourceID)
			}
		}
		validateLessonVideoEmbeds(lesson, videosByID, errors)
	}
}

func validateLessonVideoEmbeds(lesson lessonFile, videosByID map[string]videoAuthoring, errors *errorCollector) {
	content := mdxOutsideFencedCode(lesson.content)
	if element := arbitraryEmbedElementPattern.FindStringSubmatch(content); element != nil {
		errors.addf(lesson.path, "lesson %q cannot contain arbitrary <%s> markup; use a trusted MDX component such as YouTubeVideo", lesson.metadata.ID, strings.ToLower(element[1]))
	}
	for _, element := range youtubeVideoElementPattern.FindAllStringSubmatch(content, -1) {
		attribute := youtubeVideoIDAttributePattern.FindStringSubmatch(element[1])
		if attribute == nil {
			errors.addf(lesson.path, "lesson %q YouTubeVideo must use a static quoted id", lesson.metadata.ID)
			continue
		}
		videoID := attribute[1]
		video, ok := videosByID[videoID]
		if !ok {
			errors.addf(lesson.path, "lesson %q references unknown curated video id %q in this module", lesson.metadata.ID, videoID)
			continue
		}
		if !containsString(video.LessonIDs, lesson.metadata.ID) {
			errors.addf(lesson.path, "lesson %q embeds video %q but the video does not associate that lesson in lessonIds", lesson.metadata.ID, videoID)
		}
	}
}

func mdxOutsideFencedCode(content string) string {
	var builder strings.Builder
	fence := ""
	for _, line := range strings.SplitAfter(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if fence == "" && (strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")) {
			fence = trimmed[:3]
			continue
		}
		if fence != "" {
			if strings.HasPrefix(trimmed, fence) {
				fence = ""
			}
			continue
		}
		builder.WriteString(line)
	}
	return builder.String()
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func validateObjectiveReferences(modules []moduleFile, objectiveIDs map[string]string, errors *errorCollector) {
	for _, module := range modules {
		if !module.metadataOK {
			continue
		}
		for _, objective := range module.metadata.Objectives {
			for _, prerequisite := range objective.Prerequisites {
				if prerequisite == objective.ID {
					errors.addf(module.path, "objective %q cannot list itself as a prerequisite", objective.ID)
				}
				if _, ok := objectiveIDs[prerequisite]; !ok {
					errors.addf(module.path, "objective %q has unknown prerequisite %q", objective.ID, prerequisite)
				}
			}
		}
	}
}

func validatePrerequisiteCycles(modules []moduleFile, objectiveIDs, objectivePaths map[string]string, errors *errorCollector) {
	graph := make(map[string][]string, len(objectiveIDs))
	for _, module := range modules {
		if !module.metadataOK {
			continue
		}
		for _, objective := range module.metadata.Objectives {
			prerequisites := append([]string(nil), objective.Prerequisites...)
			sort.Strings(prerequisites)
			graph[objective.ID] = prerequisites
		}
	}

	state := map[string]uint8{}
	stack := []string{}
	stackIndex := map[string]int{}
	for _, id := range sortedKeys(objectiveIDs) {
		if state[id] == 0 {
			detectCycle(id, graph, objectivePaths, state, stack, stackIndex, errors)
		}
	}
}

func detectCycle(id string, graph map[string][]string, objectivePaths map[string]string, state map[string]uint8, stack []string, stackIndex map[string]int, errors *errorCollector) {
	state[id] = 1
	stackIndex[id] = len(stack)
	stack = append(stack, id)
	for _, prerequisite := range graph[id] {
		if state[prerequisite] == 0 {
			detectCycle(prerequisite, graph, objectivePaths, state, stack, stackIndex, errors)
			continue
		}
		if state[prerequisite] != 1 {
			continue
		}
		cycle := append([]string{}, stack[stackIndex[prerequisite]:]...)
		cycle = append(cycle, prerequisite)
		errors.addf(objectivePaths[id], "objective prerequisite cycle: %s", strings.Join(cycle, " -> "))
	}
	state[id] = 2
	delete(stackIndex, id)
}

func validateHTTPURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("URL %q must be a valid HTTP or HTTPS URL", raw)
	}
	return nil
}

func sortedKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func validateUniqueStrings(pathName, owner, field string, values []string, errors *errorCollector) {
	seen := make(map[string]struct{}, len(values))
	reported := make(map[string]struct{})
	for _, value := range values {
		if _, ok := seen[value]; !ok {
			seen[value] = struct{}{}
			continue
		}
		if _, ok := reported[value]; ok {
			continue
		}
		reported[value] = struct{}{}
		errors.addf(pathName, "%s field %s contains duplicate value %q", owner, field, value)
	}
}
