package curriculum

import "sort"

// StableIDPattern is the single stable-ID convention used by authored
// curriculum records.
const StableIDPattern = `^[a-z0-9]+(?:[.-][a-z0-9]+)*$`

type Course struct {
	ID          string
	Title       string
	Description string
	Order       int
	Modules     []Module
}

type Objective struct {
	CourseID      string
	ModuleID      string
	ID            string
	Title         string
	Description   string
	Prerequisites []string
}

type Video struct {
	CourseID     string
	ModuleID     string
	ID           string
	Title        string
	URL          string
	ObjectiveIDs []string
}

type Lesson struct {
	CourseID     string
	ModuleID     string
	ID           string
	Title        string
	ObjectiveIDs []string
	SourceIDs    []string
	Content      string
}

type Module struct {
	CourseID   string
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
	courses      []Course
	coursesByID  map[string]int
	modules      []Module
	modulesByKey map[moduleKey]Module
	modulesByID  map[string]Module
	lessonsByKey map[lessonKey]Lesson
	objectives   map[string]Objective
	sources      map[string]Source
}

type moduleKey struct {
	courseID string
	moduleID string
}

type lessonKey struct {
	courseID string
	moduleID string
	lessonID string
}

func NewEmptyCatalog() *Catalog {
	return &Catalog{
		courses:      []Course{},
		coursesByID:  make(map[string]int),
		modules:      []Module{},
		modulesByKey: make(map[moduleKey]Module),
		modulesByID:  make(map[string]Module),
		lessonsByKey: make(map[lessonKey]Lesson),
		objectives:   make(map[string]Objective),
		sources:      make(map[string]Source),
	}
}

func newCatalog(courses []Course, sources map[string]Source) *Catalog {
	catalog := NewEmptyCatalog()
	catalog.courses = cloneCourses(courses)
	sort.Slice(catalog.courses, func(i, j int) bool {
		if catalog.courses[i].Order != catalog.courses[j].Order {
			return catalog.courses[i].Order < catalog.courses[j].Order
		}
		return catalog.courses[i].ID < catalog.courses[j].ID
	})

	for courseIndex := range catalog.courses {
		course := &catalog.courses[courseIndex]
		sort.Slice(course.Modules, func(i, j int) bool {
			if course.Modules[i].Order != course.Modules[j].Order {
				return course.Modules[i].Order < course.Modules[j].Order
			}
			return course.Modules[i].ID < course.Modules[j].ID
		})
		catalog.coursesByID[course.ID] = courseIndex
		for _, module := range course.Modules {
			clonedModule := cloneModule(module)
			catalog.modules = append(catalog.modules, clonedModule)
			catalog.modulesByKey[moduleKey{courseID: course.ID, moduleID: module.ID}] = clonedModule
			catalog.modulesByID[module.ID] = clonedModule
			for _, lesson := range module.Lessons {
				catalog.lessonsByKey[lessonKey{courseID: course.ID, moduleID: module.ID, lessonID: lesson.ID}] = cloneLesson(lesson)
			}
			for _, objective := range module.Objectives {
				catalog.objectives[objective.ID] = cloneObjective(objective)
			}
		}
	}
	for id, source := range sources {
		catalog.sources[id] = source
	}

	return catalog
}

func (c *Catalog) Courses() []Course {
	if c == nil {
		return []Course{}
	}
	return cloneCourses(c.courses)
}

func (c *Catalog) CourseByID(courseID string) (Course, bool) {
	if c == nil {
		return Course{}, false
	}
	index, ok := c.coursesByID[courseID]
	if !ok {
		return Course{}, false
	}
	return cloneCourse(c.courses[index]), true
}

func (c *Catalog) ModulesByCourse(courseID string) []Module {
	course, ok := c.CourseByID(courseID)
	if !ok {
		return []Module{}
	}
	return course.Modules
}

func (c *Catalog) ModuleByCourse(courseID, moduleID string) (Module, bool) {
	if c == nil {
		return Module{}, false
	}
	module, ok := c.modulesByKey[moduleKey{courseID: courseID, moduleID: moduleID}]
	if !ok {
		return Module{}, false
	}
	return cloneModule(module), true
}

func (c *Catalog) LessonByCourse(courseID, moduleID, lessonID string) (Lesson, bool) {
	if c == nil {
		return Lesson{}, false
	}
	lesson, ok := c.lessonsByKey[lessonKey{courseID: courseID, moduleID: moduleID, lessonID: lessonID}]
	if !ok {
		return Lesson{}, false
	}
	return cloneLesson(lesson), true
}

// Modules is a transitional compatibility method for the current single-course
// HTTP API. Course-aware callers should use ModulesByCourse.
func (c *Catalog) Modules() []Module {
	if c == nil {
		return []Module{}
	}
	return cloneModules(c.modules)
}

// ModuleByID is a transitional compatibility method for the current
// single-course HTTP API. Module IDs remain globally unique in this phase.
func (c *Catalog) ModuleByID(moduleID string) (Module, bool) {
	if c == nil {
		return Module{}, false
	}
	module, ok := c.modulesByID[moduleID]
	if !ok {
		return Module{}, false
	}
	return cloneModule(module), true
}

// LessonByID is a transitional compatibility method for the current
// single-course HTTP API. Module IDs remain globally unique in this phase.
func (c *Catalog) LessonByID(moduleID, lessonID string) (Lesson, bool) {
	module, ok := c.ModuleByID(moduleID)
	if !ok {
		return Lesson{}, false
	}
	return c.LessonByCourse(module.CourseID, moduleID, lessonID)
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

func (c *Catalog) CourseCount() int {
	if c == nil {
		return 0
	}
	return len(c.courses)
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

func cloneCourses(courses []Course) []Course {
	cloned := make([]Course, len(courses))
	for index, course := range courses {
		cloned[index] = cloneCourse(course)
	}
	return cloned
}

func cloneCourse(course Course) Course {
	cloned := course
	cloned.Modules = cloneModules(course.Modules)
	return cloned
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
