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

type Worksheet struct {
	CourseID     string
	ModuleID     string
	ID           string
	Title        string
	LessonID     string
	Order        int
	ObjectiveIDs []string
	Instructions string
	Problems     []WorksheetProblem
}

type WorksheetProblem struct {
	ID             string
	Prompt         string
	ObjectiveIDs   []string
	ExpectedAnswer string
	RequiresWork   bool
	ResponseLines  int
	Rubric         []string
}

type Module struct {
	CourseID   string
	ID         string
	Title      string
	Order      int
	Objectives []Objective
	Videos     []Video
	Lessons    []Lesson
	Worksheets []Worksheet
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
	courses            []Course
	coursesByID        map[string]int
	modules            []Module
	modulesByKey       map[moduleKey]Module
	lessonsByKey       map[lessonKey]Lesson
	worksheetsByKey    map[worksheetKey]Worksheet
	worksheetsByModule map[moduleKey][]Worksheet
	worksheetsByLesson map[lessonKey][]Worksheet
	objectives         map[string]Objective
	sources            map[string]Source
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

type worksheetKey struct {
	courseID    string
	moduleID    string
	worksheetID string
}

func NewEmptyCatalog() *Catalog {
	return &Catalog{
		courses:            []Course{},
		coursesByID:        make(map[string]int),
		modules:            []Module{},
		modulesByKey:       make(map[moduleKey]Module),
		lessonsByKey:       make(map[lessonKey]Lesson),
		worksheetsByKey:    make(map[worksheetKey]Worksheet),
		worksheetsByModule: make(map[moduleKey][]Worksheet),
		worksheetsByLesson: make(map[lessonKey][]Worksheet),
		objectives:         make(map[string]Objective),
		sources:            make(map[string]Source),
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
		for moduleIndex := range course.Modules {
			module := &course.Modules[moduleIndex]
			lessonOrder := make(map[string]int, len(module.Lessons))
			for index, lesson := range module.Lessons {
				lessonOrder[lesson.ID] = index
			}
			sort.Slice(module.Worksheets, func(i, j int) bool {
				leftLesson, rightLesson := lessonOrder[module.Worksheets[i].LessonID], lessonOrder[module.Worksheets[j].LessonID]
				if leftLesson != rightLesson {
					return leftLesson < rightLesson
				}
				if module.Worksheets[i].Order != module.Worksheets[j].Order {
					return module.Worksheets[i].Order < module.Worksheets[j].Order
				}
				return module.Worksheets[i].ID < module.Worksheets[j].ID
			})
			clonedModule := cloneModule(*module)
			catalog.modules = append(catalog.modules, clonedModule)
			catalog.modulesByKey[moduleKey{courseID: course.ID, moduleID: module.ID}] = clonedModule
			for _, lesson := range module.Lessons {
				catalog.lessonsByKey[lessonKey{courseID: course.ID, moduleID: module.ID, lessonID: lesson.ID}] = cloneLesson(lesson)
			}
			for _, worksheet := range clonedModule.Worksheets {
				clonedWorksheet := cloneWorksheet(worksheet)
				catalog.worksheetsByKey[worksheetKey{courseID: course.ID, moduleID: module.ID, worksheetID: worksheet.ID}] = clonedWorksheet
				moduleLookupKey := moduleKey{courseID: course.ID, moduleID: module.ID}
				catalog.worksheetsByModule[moduleLookupKey] = append(catalog.worksheetsByModule[moduleLookupKey], clonedWorksheet)
				lessonLookupKey := lessonKey{courseID: course.ID, moduleID: module.ID, lessonID: worksheet.LessonID}
				catalog.worksheetsByLesson[lessonLookupKey] = append(catalog.worksheetsByLesson[lessonLookupKey], clonedWorksheet)
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

func (c *Catalog) WorksheetByCourse(courseID, moduleID, worksheetID string) (Worksheet, bool) {
	if c == nil {
		return Worksheet{}, false
	}
	worksheet, ok := c.worksheetsByKey[worksheetKey{courseID: courseID, moduleID: moduleID, worksheetID: worksheetID}]
	if !ok {
		return Worksheet{}, false
	}
	return cloneWorksheet(worksheet), true
}

func (c *Catalog) WorksheetsByModule(courseID, moduleID string) []Worksheet {
	if c == nil {
		return []Worksheet{}
	}
	return cloneWorksheets(c.worksheetsByModule[moduleKey{courseID: courseID, moduleID: moduleID}])
}

func (c *Catalog) WorksheetsByLesson(courseID, moduleID, lessonID string) []Worksheet {
	if c == nil {
		return []Worksheet{}
	}
	return cloneWorksheets(c.worksheetsByLesson[lessonKey{courseID: courseID, moduleID: moduleID, lessonID: lessonID}])
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

func (c *Catalog) WorksheetCount() int {
	if c == nil {
		return 0
	}
	return len(c.worksheetsByKey)
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
	cloned.Worksheets = cloneWorksheets(module.Worksheets)
	return cloned
}

func cloneWorksheets(worksheets []Worksheet) []Worksheet {
	if worksheets == nil {
		return []Worksheet{}
	}
	cloned := make([]Worksheet, len(worksheets))
	for index, worksheet := range worksheets {
		cloned[index] = cloneWorksheet(worksheet)
	}
	return cloned
}

func cloneWorksheet(worksheet Worksheet) Worksheet {
	cloned := worksheet
	cloned.ObjectiveIDs = cloneStrings(worksheet.ObjectiveIDs)
	cloned.Problems = make([]WorksheetProblem, len(worksheet.Problems))
	for index, problem := range worksheet.Problems {
		cloned.Problems[index] = problem
		cloned.Problems[index].ObjectiveIDs = cloneStrings(problem.ObjectiveIDs)
		cloned.Problems[index].Rubric = cloneStrings(problem.Rubric)
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
