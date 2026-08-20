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

type Exercise struct {
	CourseID     string
	ModuleID     string
	ID           string
	Title        string
	LessonID     string
	Order        int
	ObjectiveIDs []string
	Prompt       string
	StarterCode  string
	Tests        []ExerciseTest
}

type ExerciseTest struct {
	ID         string
	Title      string
	Visibility string
	Code       string
}

const (
	ExerciseTestVisible = "visible"
	ExerciseTestHidden  = "hidden"
)

type ReviewItem struct {
	CourseID       string
	ModuleID       string
	ID             string
	Order          int
	ObjectiveIDs   []string
	SourceLessonID string
	Prompt         string
	Answer         string
	Hint           string
}

type Module struct {
	CourseID    string
	ID          string
	Title       string
	Order       int
	Objectives  []Objective
	Videos      []Video
	Lessons     []Lesson
	Worksheets  []Worksheet
	Exercises   []Exercise
	ReviewItems []ReviewItem
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
	courses                []Course
	coursesByID            map[string]int
	modules                []Module
	modulesByKey           map[moduleKey]Module
	lessonsByKey           map[lessonKey]Lesson
	worksheetsByKey        map[worksheetKey]Worksheet
	worksheetsByModule     map[moduleKey][]Worksheet
	worksheetsByLesson     map[lessonKey][]Worksheet
	exercisesByKey         map[exerciseKey]Exercise
	exercisesByModule      map[moduleKey][]Exercise
	exercisesByLesson      map[lessonKey][]Exercise
	reviewItemsByKey       map[reviewItemKey]ReviewItem
	reviewItemsByModule    map[moduleKey][]ReviewItem
	reviewItemsByObjective map[string][]ReviewItem
	objectives             map[string]Objective
	sources                map[string]Source
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

type exerciseKey struct {
	courseID   string
	moduleID   string
	exerciseID string
}

type reviewItemKey struct {
	courseID     string
	moduleID     string
	reviewItemID string
}

func NewEmptyCatalog() *Catalog {
	return &Catalog{
		courses:                []Course{},
		coursesByID:            make(map[string]int),
		modules:                []Module{},
		modulesByKey:           make(map[moduleKey]Module),
		lessonsByKey:           make(map[lessonKey]Lesson),
		worksheetsByKey:        make(map[worksheetKey]Worksheet),
		worksheetsByModule:     make(map[moduleKey][]Worksheet),
		worksheetsByLesson:     make(map[lessonKey][]Worksheet),
		exercisesByKey:         make(map[exerciseKey]Exercise),
		exercisesByModule:      make(map[moduleKey][]Exercise),
		exercisesByLesson:      make(map[lessonKey][]Exercise),
		reviewItemsByKey:       make(map[reviewItemKey]ReviewItem),
		reviewItemsByModule:    make(map[moduleKey][]ReviewItem),
		reviewItemsByObjective: make(map[string][]ReviewItem),
		objectives:             make(map[string]Objective),
		sources:                make(map[string]Source),
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
			sort.Slice(module.Exercises, func(i, j int) bool {
				leftLesson, rightLesson := lessonOrder[module.Exercises[i].LessonID], lessonOrder[module.Exercises[j].LessonID]
				if leftLesson != rightLesson {
					return leftLesson < rightLesson
				}
				if module.Exercises[i].Order != module.Exercises[j].Order {
					return module.Exercises[i].Order < module.Exercises[j].Order
				}
				return module.Exercises[i].ID < module.Exercises[j].ID
			})
			sort.Slice(module.ReviewItems, func(i, j int) bool {
				if module.ReviewItems[i].Order != module.ReviewItems[j].Order {
					return module.ReviewItems[i].Order < module.ReviewItems[j].Order
				}
				return module.ReviewItems[i].ID < module.ReviewItems[j].ID
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
			for _, exercise := range clonedModule.Exercises {
				clonedExercise := cloneExercise(exercise)
				catalog.exercisesByKey[exerciseKey{courseID: course.ID, moduleID: module.ID, exerciseID: exercise.ID}] = clonedExercise
				moduleLookupKey := moduleKey{courseID: course.ID, moduleID: module.ID}
				catalog.exercisesByModule[moduleLookupKey] = append(catalog.exercisesByModule[moduleLookupKey], clonedExercise)
				lessonLookupKey := lessonKey{courseID: course.ID, moduleID: module.ID, lessonID: exercise.LessonID}
				catalog.exercisesByLesson[lessonLookupKey] = append(catalog.exercisesByLesson[lessonLookupKey], clonedExercise)
			}
			for _, reviewItem := range clonedModule.ReviewItems {
				clonedReviewItem := cloneReviewItem(reviewItem)
				catalog.reviewItemsByKey[reviewItemKey{courseID: course.ID, moduleID: module.ID, reviewItemID: reviewItem.ID}] = clonedReviewItem
				moduleLookupKey := moduleKey{courseID: course.ID, moduleID: module.ID}
				catalog.reviewItemsByModule[moduleLookupKey] = append(catalog.reviewItemsByModule[moduleLookupKey], clonedReviewItem)
				for _, objectiveID := range reviewItem.ObjectiveIDs {
					catalog.reviewItemsByObjective[objectiveID] = append(catalog.reviewItemsByObjective[objectiveID], clonedReviewItem)
				}
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

func (c *Catalog) ExerciseByCourse(courseID, moduleID, exerciseID string) (Exercise, bool) {
	if c == nil {
		return Exercise{}, false
	}
	exercise, ok := c.exercisesByKey[exerciseKey{courseID: courseID, moduleID: moduleID, exerciseID: exerciseID}]
	if !ok {
		return Exercise{}, false
	}
	return cloneExercise(exercise), true
}

func (c *Catalog) ExercisesByModule(courseID, moduleID string) []Exercise {
	if c == nil {
		return []Exercise{}
	}
	return cloneExercises(c.exercisesByModule[moduleKey{courseID: courseID, moduleID: moduleID}])
}

func (c *Catalog) ExercisesByLesson(courseID, moduleID, lessonID string) []Exercise {
	if c == nil {
		return []Exercise{}
	}
	return cloneExercises(c.exercisesByLesson[lessonKey{courseID: courseID, moduleID: moduleID, lessonID: lessonID}])
}

func (c *Catalog) ReviewItemByCourse(courseID, moduleID, reviewItemID string) (ReviewItem, bool) {
	if c == nil {
		return ReviewItem{}, false
	}
	reviewItem, ok := c.reviewItemsByKey[reviewItemKey{courseID: courseID, moduleID: moduleID, reviewItemID: reviewItemID}]
	if !ok {
		return ReviewItem{}, false
	}
	return cloneReviewItem(reviewItem), true
}

func (c *Catalog) ReviewItemsByModule(courseID, moduleID string) []ReviewItem {
	if c == nil {
		return []ReviewItem{}
	}
	return cloneReviewItems(c.reviewItemsByModule[moduleKey{courseID: courseID, moduleID: moduleID}])
}

func (c *Catalog) ReviewItemsByObjective(objectiveID string) []ReviewItem {
	if c == nil {
		return []ReviewItem{}
	}
	return cloneReviewItems(c.reviewItemsByObjective[objectiveID])
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

func (c *Catalog) ExerciseCount() int {
	if c == nil {
		return 0
	}
	return len(c.exercisesByKey)
}

func (c *Catalog) ReviewItemCount() int {
	if c == nil {
		return 0
	}
	return len(c.reviewItemsByKey)
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

// UnusedSourceIDs returns registered source IDs that no lesson references.
// Unused sources are diagnostics rather than invalid curriculum because authors
// may register a source before its first lesson is committed.
func (c *Catalog) UnusedSourceIDs() []string {
	if c == nil {
		return []string{}
	}
	used := make(map[string]struct{})
	for _, course := range c.courses {
		for _, module := range course.Modules {
			for _, lesson := range module.Lessons {
				for _, sourceID := range lesson.SourceIDs {
					used[sourceID] = struct{}{}
				}
			}
		}
	}
	unused := make([]string, 0)
	for sourceID := range c.sources {
		if _, ok := used[sourceID]; !ok {
			unused = append(unused, sourceID)
		}
	}
	sort.Strings(unused)
	return unused
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
	cloned.Exercises = cloneExercises(module.Exercises)
	cloned.ReviewItems = cloneReviewItems(module.ReviewItems)
	return cloned
}

func cloneReviewItems(reviewItems []ReviewItem) []ReviewItem {
	if reviewItems == nil {
		return []ReviewItem{}
	}
	cloned := make([]ReviewItem, len(reviewItems))
	for index, reviewItem := range reviewItems {
		cloned[index] = cloneReviewItem(reviewItem)
	}
	return cloned
}

func cloneReviewItem(reviewItem ReviewItem) ReviewItem {
	cloned := reviewItem
	cloned.ObjectiveIDs = cloneStrings(reviewItem.ObjectiveIDs)
	return cloned
}

func cloneExercises(exercises []Exercise) []Exercise {
	if exercises == nil {
		return []Exercise{}
	}
	cloned := make([]Exercise, len(exercises))
	for index, exercise := range exercises {
		cloned[index] = cloneExercise(exercise)
	}
	return cloned
}

func cloneExercise(exercise Exercise) Exercise {
	cloned := exercise
	cloned.ObjectiveIDs = cloneStrings(exercise.ObjectiveIDs)
	cloned.Tests = append([]ExerciseTest(nil), exercise.Tests...)
	if cloned.Tests == nil {
		cloned.Tests = []ExerciseTest{}
	}
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
