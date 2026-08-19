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

func validateModules(modules []moduleFile, sources map[string]sourceAuthoring, errors *errorCollector) {
	moduleIDs := map[string]string{}
	moduleOrders := map[int]string{}
	objectiveIDs := map[string]string{}
	objectivePaths := map[string]string{}

	for _, module := range modules {
		if !module.metadataOK {
			continue
		}
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

	for _, module := range modules {
		validateLessonFiles(module, objectiveIDs, sources, errors)
	}
	validateObjectiveReferences(modules, objectiveIDs, errors)
	validatePrerequisiteCycles(modules, objectiveIDs, objectivePaths, errors)
	validateVideoObjectiveReferences(modules, objectiveIDs, errors)
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
