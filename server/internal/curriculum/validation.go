package curriculum

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var stableIDPattern = regexp.MustCompile(StableIDPattern)

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
	moduleIDs := map[string]string{}
	objectiveIDs := map[string]string{}
	objectivePaths := map[string]string{}
	modules := []moduleFile{}

	for _, course := range courses {
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
			} else if previous, ok := courseOrders[*metadata.Order]; ok {
				errors.addf(course.path, "duplicate course order %d (already declared in %s)", *metadata.Order, previous)
			} else {
				courseOrders[*metadata.Order] = course.path
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
	}
	validateObjectiveReferences(modules, objectiveIDs, errors)
	validatePrerequisiteCycles(modules, objectiveIDs, objectivePaths, errors)
	validateVideoObjectiveReferences(modules, objectiveIDs, errors)
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
	} else if previous, ok := moduleOrders[*metadata.Order]; ok {
		errors.addf(metadataPath, "duplicate module order %d (already declared in %s)", *metadata.Order, previous)
	} else {
		moduleOrders[*metadata.Order] = metadataPath
	}

	validateModuleObjectives(metadata, metadataPath, objectiveIDs, objectivePaths, errors)
	validateModuleVideos(metadata, metadataPath, errors)
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
		if previous, ok := objectiveIDs[objective.ID]; ok {
			errors.addf(metadataPath, "duplicate objective id %q (already declared in %s)", objective.ID, previous)
		} else {
			objectiveIDs[objective.ID] = metadataPath
			objectivePaths[objective.ID] = metadataPath
		}
	}
}

func validateModuleVideos(metadata moduleAuthoring, metadataPath string, errors *errorCollector) {
	videoIDs := map[string]struct{}{}
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
		if err := validateHTTPURL(video.URL); err != nil {
			errors.addf(metadataPath, "video %q: %v", video.ID, err)
		}
	}
}

func validateModuleLessonReferences(module moduleFile, errors *errorCollector) {
	if !module.metadataOK {
		return
	}
	declared := map[string]struct{}{}
	for _, lessonID := range module.metadata.Lessons {
		if _, ok := declared[lessonID]; ok {
			errors.addf(module.path, "duplicate lesson id reference %q", lessonID)
			continue
		}
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
	}
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

func validateVideoObjectiveReferences(modules []moduleFile, objectiveIDs map[string]string, errors *errorCollector) {
	for _, module := range modules {
		if !module.metadataOK {
			continue
		}
		for _, video := range module.metadata.Videos {
			for _, objectiveID := range video.ObjectiveIDs {
				if _, ok := objectiveIDs[objectiveID]; !ok {
					errors.addf(module.path, "video %q has unknown objective id %q", video.ID, objectiveID)
				}
			}
		}
	}
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
