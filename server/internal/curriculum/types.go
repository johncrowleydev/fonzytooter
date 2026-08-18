package curriculum

import "sort"

// StableIDPattern is the single stable-ID convention used by authored
// curriculum records.
const StableIDPattern = `^[a-z0-9]+(?:[.-][a-z0-9]+)*$`

type Objective struct {
	ID            string
	Title         string
	Description   string
	Prerequisites []string
}

type Video struct {
	ID           string
	Title        string
	URL          string
	ObjectiveIDs []string
}

type Lesson struct {
	ID           string
	Title        string
	ObjectiveIDs []string
	SourceIDs    []string
	Content      string
}

type Module struct {
	ID         string
	Title      string
	Order      int
	Objectives []Objective
	Videos     []Video
	Lessons    []Lesson
}

type Source struct {
	ID    string
	Title string
	URL   string
}

// Catalog is a validated, read-only snapshot of the Git-authored curriculum.
// Its exported lookup methods return copies, so callers cannot mutate the
// catalog's internal indexes or slices.
type Catalog struct {
	modules      []Module
	modulesByID  map[string]int
	lessonsByKey map[lessonKey]Lesson
	objectives   map[string]Objective
	sources      map[string]Source
}

type lessonKey struct {
	moduleID string
	lessonID string
}

func NewEmptyCatalog() *Catalog {
	return &Catalog{
		modules:      []Module{},
		modulesByID:  make(map[string]int),
		lessonsByKey: make(map[lessonKey]Lesson),
		objectives:   make(map[string]Objective),
		sources:      make(map[string]Source),
	}
}

func newCatalog(modules []Module, sources map[string]Source) *Catalog {
	catalog := NewEmptyCatalog()
	catalog.modules = cloneModules(modules)
	sort.Slice(catalog.modules, func(i, j int) bool {
		if catalog.modules[i].Order != catalog.modules[j].Order {
			return catalog.modules[i].Order < catalog.modules[j].Order
		}
		return catalog.modules[i].ID < catalog.modules[j].ID
	})

	for index, module := range catalog.modules {
		catalog.modulesByID[module.ID] = index
		for _, lesson := range module.Lessons {
			catalog.lessonsByKey[lessonKey{moduleID: module.ID, lessonID: lesson.ID}] = cloneLesson(lesson)
		}
		for _, objective := range module.Objectives {
			catalog.objectives[objective.ID] = cloneObjective(objective)
		}
	}
	for id, source := range sources {
		catalog.sources[id] = source
	}

	return catalog
}

func (c *Catalog) Modules() []Module {
	if c == nil {
		return []Module{}
	}
	return cloneModules(c.modules)
}

func (c *Catalog) ModuleByID(id string) (Module, bool) {
	if c == nil {
		return Module{}, false
	}
	index, ok := c.modulesByID[id]
	if !ok {
		return Module{}, false
	}
	return cloneModule(c.modules[index]), true
}

func (c *Catalog) LessonByID(moduleID, lessonID string) (Lesson, bool) {
	if c == nil {
		return Lesson{}, false
	}
	lesson, ok := c.lessonsByKey[lessonKey{moduleID: moduleID, lessonID: lessonID}]
	if !ok {
		return Lesson{}, false
	}
	return cloneLesson(lesson), true
}

func (c *Catalog) ObjectiveByID(id string) (Objective, bool) {
	if c == nil {
		return Objective{}, false
	}
	objective, ok := c.objectives[id]
	if !ok {
		return Objective{}, false
	}
	return cloneObjective(objective), true
}

func (c *Catalog) SourceByID(id string) (Source, bool) {
	if c == nil {
		return Source{}, false
	}
	source, ok := c.sources[id]
	return source, ok
}

func (c *Catalog) ModuleCount() int {
	if c == nil {
		return 0
	}
	return len(c.modules)
}

func (c *Catalog) LessonCount() int {
	if c == nil {
		return 0
	}
	return len(c.lessonsByKey)
}

func (c *Catalog) ObjectiveCount() int {
	if c == nil {
		return 0
	}
	return len(c.objectives)
}

func (c *Catalog) SourceCount() int {
	if c == nil {
		return 0
	}
	return len(c.sources)
}

func cloneModules(modules []Module) []Module {
	cloned := make([]Module, len(modules))
	for index, module := range modules {
		cloned[index] = cloneModule(module)
	}
	return cloned
}

func cloneModule(module Module) Module {
	cloned := module
	cloned.Objectives = make([]Objective, len(module.Objectives))
	for index, objective := range module.Objectives {
		cloned.Objectives[index] = cloneObjective(objective)
	}
	cloned.Videos = make([]Video, len(module.Videos))
	for index, video := range module.Videos {
		cloned.Videos[index] = cloneVideo(video)
	}
	cloned.Lessons = make([]Lesson, len(module.Lessons))
	for index, lesson := range module.Lessons {
		cloned.Lessons[index] = cloneLesson(lesson)
	}
	return cloned
}

func cloneObjective(objective Objective) Objective {
	cloned := objective
	cloned.Prerequisites = cloneStrings(objective.Prerequisites)
	return cloned
}

func cloneVideo(video Video) Video {
	cloned := video
	cloned.ObjectiveIDs = cloneStrings(video.ObjectiveIDs)
	return cloned
}

func cloneLesson(lesson Lesson) Lesson {
	cloned := lesson
	cloned.ObjectiveIDs = cloneStrings(lesson.ObjectiveIDs)
	cloned.SourceIDs = cloneStrings(lesson.SourceIDs)
	return cloned
}

func cloneStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return append([]string(nil), values...)
}
