package curriculum

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type courseAuthoring struct {
	ID          string `yaml:"id"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Order       *int   `yaml:"order"`
}

type moduleAuthoring struct {
	ID         string               `yaml:"id"`
	Title      string               `yaml:"title"`
	Order      *int                 `yaml:"order"`
	Objectives []objectiveAuthoring `yaml:"objectives"`
	Videos     []videoAuthoring     `yaml:"videos"`
	Lessons    []string             `yaml:"lessons"`
}

type objectiveAuthoring struct {
	ID            string   `yaml:"id"`
	Title         string   `yaml:"title"`
	Description   string   `yaml:"description"`
	Prerequisites []string `yaml:"prerequisites"`
}

type videoAuthoring struct {
	ID           string   `yaml:"id"`
	Title        string   `yaml:"title"`
	URL          string   `yaml:"url"`
	ObjectiveIDs []string `yaml:"objectiveIds"`
}

type lessonFrontmatter struct {
	ID           string   `yaml:"id"`
	Title        string   `yaml:"title"`
	ObjectiveIDs []string `yaml:"objectiveIds"`
	SourceIDs    []string `yaml:"sourceIds"`
}

type worksheetAuthoring struct {
	ID           string                      `yaml:"id"`
	Title        string                      `yaml:"title"`
	LessonID     string                      `yaml:"lessonId"`
	Order        *int                        `yaml:"order"`
	ObjectiveIDs []string                    `yaml:"objectiveIds"`
	Instructions string                      `yaml:"instructions"`
	Problems     []worksheetProblemAuthoring `yaml:"problems"`
}

type worksheetProblemAuthoring struct {
	ID             string   `yaml:"id"`
	Prompt         string   `yaml:"prompt"`
	ObjectiveIDs   []string `yaml:"objectiveIds"`
	ExpectedAnswer string   `yaml:"expectedAnswer"`
	RequiresWork   *bool    `yaml:"requiresWork"`
	ResponseLines  *int     `yaml:"responseLines"`
	Rubric         []string `yaml:"rubric"`
}

type exerciseAuthoring struct {
	ID           string                  `yaml:"id"`
	Title        string                  `yaml:"title"`
	LessonID     string                  `yaml:"lessonId"`
	Order        *int                    `yaml:"order"`
	ObjectiveIDs []string                `yaml:"objectiveIds"`
	Prompt       string                  `yaml:"prompt"`
	StarterCode  string                  `yaml:"starterCode"`
	Tests        []exerciseTestAuthoring `yaml:"tests"`
}

type exerciseTestAuthoring struct {
	ID         string `yaml:"id"`
	Title      string `yaml:"title"`
	Visibility string `yaml:"visibility"`
	Code       string `yaml:"code"`
}

type reviewItemAuthoring struct {
	ID             string   `yaml:"id"`
	Order          *int     `yaml:"order"`
	ObjectiveIDs   []string `yaml:"objectiveIds"`
	SourceLessonID string   `yaml:"sourceLessonId"`
	Prompt         string   `yaml:"prompt"`
	Answer         string   `yaml:"answer"`
	Hint           string   `yaml:"hint"`
}

type sourceAuthoring struct {
	Title string `yaml:"title"`
	URL   string `yaml:"url"`
}

type sourceRegistry struct {
	Sources map[string]sourceAuthoring `yaml:"sources"`
}

type courseFile struct {
	path       string
	metadata   courseAuthoring
	modules    []moduleFile
	metadataOK bool
}

type moduleFile struct {
	path        string
	metadata    moduleAuthoring
	lessons     []lessonFile
	worksheets  []worksheetFile
	exercises   []exerciseFile
	reviewItems []reviewItemFile
	metadataOK  bool
}

type reviewItemFile struct {
	path       string
	metadata   reviewItemAuthoring
	metadataOK bool
}

type worksheetFile struct {
	path       string
	metadata   worksheetAuthoring
	metadataOK bool
}

type exerciseFile struct {
	path       string
	metadata   exerciseAuthoring
	metadataOK bool
}

type lessonFile struct {
	path       string
	metadata   lessonFrontmatter
	content    string
	metadataOK bool
}

// Load reads and validates one curriculum snapshot from fsys. It never keeps
// the filesystem or authored YAML nodes; the returned catalog is in-memory.
func Load(fsys fs.FS) (*Catalog, error) {
	errors := &errorCollector{}
	validateCurriculumRoot(fsys, errors)
	if err := errors.Err(); err != nil {
		return nil, err
	}

	sources := loadSources(fsys, errors)
	courses := loadCourses(fsys, errors)
	validateSources(sources, errors)
	validateCourses(courses, sources, errors)

	if err := errors.Err(); err != nil {
		return nil, err
	}

	loadedSources := make(map[string]Source, len(sources))
	for id, source := range sources {
		loadedSources[id] = Source{ID: id, Title: source.Title, URL: source.URL}
	}

	loadedCourses := make([]Course, 0, len(courses))
	for _, course := range courses {
		loadedCourse := Course{
			ID:          course.metadata.ID,
			Title:       course.metadata.Title,
			Description: course.metadata.Description,
			Order:       *course.metadata.Order,
			Modules:     make([]Module, 0, len(course.modules)),
		}
		for _, module := range course.modules {
			loadedModule := Module{
				CourseID:    course.metadata.ID,
				ID:          module.metadata.ID,
				Title:       module.metadata.Title,
				Order:       *module.metadata.Order,
				Objectives:  make([]Objective, 0, len(module.metadata.Objectives)),
				Videos:      make([]Video, 0, len(module.metadata.Videos)),
				Lessons:     make([]Lesson, 0, len(module.metadata.Lessons)),
				Worksheets:  make([]Worksheet, 0, len(module.worksheets)),
				Exercises:   make([]Exercise, 0, len(module.exercises)),
				ReviewItems: make([]ReviewItem, 0, len(module.reviewItems)),
			}
			for _, objective := range module.metadata.Objectives {
				loadedModule.Objectives = append(loadedModule.Objectives, Objective{
					CourseID:      course.metadata.ID,
					ModuleID:      module.metadata.ID,
					ID:            objective.ID,
					Title:         objective.Title,
					Description:   objective.Description,
					Prerequisites: cloneStrings(objective.Prerequisites),
				})
			}
			for _, video := range module.metadata.Videos {
				loadedModule.Videos = append(loadedModule.Videos, Video{
					CourseID:     course.metadata.ID,
					ModuleID:     module.metadata.ID,
					ID:           video.ID,
					Title:        video.Title,
					URL:          video.URL,
					ObjectiveIDs: cloneStrings(video.ObjectiveIDs),
				})
			}

			lessonsByID := make(map[string]lessonFile, len(module.lessons))
			for _, lesson := range module.lessons {
				lessonsByID[lesson.metadata.ID] = lesson
			}
			for _, lessonID := range module.metadata.Lessons {
				lesson, ok := lessonsByID[lessonID]
				if !ok {
					continue
				}
				loadedModule.Lessons = append(loadedModule.Lessons, Lesson{
					CourseID:     course.metadata.ID,
					ModuleID:     module.metadata.ID,
					ID:           lesson.metadata.ID,
					Title:        lesson.metadata.Title,
					ObjectiveIDs: cloneStrings(lesson.metadata.ObjectiveIDs),
					SourceIDs:    cloneStrings(lesson.metadata.SourceIDs),
					Content:      lesson.content,
				})
			}
			for _, worksheet := range module.worksheets {
				loadedWorksheet := Worksheet{
					CourseID:     course.metadata.ID,
					ModuleID:     module.metadata.ID,
					ID:           worksheet.metadata.ID,
					Title:        worksheet.metadata.Title,
					LessonID:     worksheet.metadata.LessonID,
					Order:        *worksheet.metadata.Order,
					ObjectiveIDs: cloneStrings(worksheet.metadata.ObjectiveIDs),
					Instructions: worksheet.metadata.Instructions,
					Problems:     make([]WorksheetProblem, 0, len(worksheet.metadata.Problems)),
				}
				for _, problem := range worksheet.metadata.Problems {
					loadedWorksheet.Problems = append(loadedWorksheet.Problems, WorksheetProblem{
						ID:             problem.ID,
						Prompt:         problem.Prompt,
						ObjectiveIDs:   cloneStrings(problem.ObjectiveIDs),
						ExpectedAnswer: problem.ExpectedAnswer,
						RequiresWork:   *problem.RequiresWork,
						ResponseLines:  *problem.ResponseLines,
						Rubric:         cloneStrings(problem.Rubric),
					})
				}
				loadedModule.Worksheets = append(loadedModule.Worksheets, loadedWorksheet)
			}
			for _, exercise := range module.exercises {
				loadedExercise := Exercise{
					CourseID: course.metadata.ID, ModuleID: module.metadata.ID,
					ID: exercise.metadata.ID, Title: exercise.metadata.Title,
					LessonID: exercise.metadata.LessonID, Order: *exercise.metadata.Order,
					ObjectiveIDs: cloneStrings(exercise.metadata.ObjectiveIDs),
					Prompt:       exercise.metadata.Prompt, StarterCode: exercise.metadata.StarterCode,
					Tests: make([]ExerciseTest, 0, len(exercise.metadata.Tests)),
				}
				for _, test := range exercise.metadata.Tests {
					loadedExercise.Tests = append(loadedExercise.Tests, ExerciseTest{
						ID: test.ID, Title: test.Title, Visibility: test.Visibility, Code: test.Code,
					})
				}
				loadedModule.Exercises = append(loadedModule.Exercises, loadedExercise)
			}
			for _, reviewItem := range module.reviewItems {
				loadedModule.ReviewItems = append(loadedModule.ReviewItems, ReviewItem{
					CourseID:       course.metadata.ID,
					ModuleID:       module.metadata.ID,
					ID:             reviewItem.metadata.ID,
					Order:          *reviewItem.metadata.Order,
					ObjectiveIDs:   cloneStrings(reviewItem.metadata.ObjectiveIDs),
					SourceLessonID: reviewItem.metadata.SourceLessonID,
					Prompt:         reviewItem.metadata.Prompt,
					Answer:         reviewItem.metadata.Answer,
					Hint:           reviewItem.metadata.Hint,
				})
			}
			loadedCourse.Modules = append(loadedCourse.Modules, loadedModule)
		}
		loadedCourses = append(loadedCourses, loadedCourse)
	}

	return newCatalog(loadedCourses, loadedSources), nil
}

func validateCurriculumRoot(fsys fs.FS, collector *errorCollector) {
	validateRequiredFile(fsys, "sources.yaml", collector)
	validateRequiredDirectory(fsys, "courses", collector)
}

func validateRequiredFile(fsys fs.FS, filePath string, collector *errorCollector) bool {
	info, err := fs.Stat(fsys, filePath)
	if errors.Is(err, fs.ErrNotExist) {
		collector.add(filePath, errors.New("required file is missing"))
		return false
	}
	if err != nil {
		collector.add(filePath, fmt.Errorf("cannot stat required file: %w", err))
		return false
	}
	if info.IsDir() {
		collector.add(filePath, errors.New("required file is a directory"))
		return false
	}
	if !info.Mode().IsRegular() {
		collector.add(filePath, errors.New("required file is not a regular file"))
		return false
	}
	return true
}

func validateRequiredDirectory(fsys fs.FS, directoryPath string, collector *errorCollector) bool {
	info, err := fs.Stat(fsys, directoryPath)
	if errors.Is(err, fs.ErrNotExist) {
		collector.add(directoryPath, errors.New("required directory is missing"))
		return false
	}
	if err != nil {
		collector.add(directoryPath, fmt.Errorf("cannot stat required directory: %w", err))
		return false
	}
	if !info.IsDir() {
		collector.add(directoryPath, errors.New("required directory is not a directory"))
		return false
	}
	return true
}

func loadSources(fsys fs.FS, collector *errorCollector) map[string]sourceAuthoring {
	sources := map[string]sourceAuthoring{}
	data, err := fs.ReadFile(fsys, "sources.yaml")
	if errors.Is(err, fs.ErrNotExist) {
		return sources
	}
	if err != nil {
		collector.add("sources.yaml", err)
		return sources
	}

	var registry sourceRegistry
	if err := decodeYAML(data, &registry); err != nil {
		collector.add("sources.yaml", err)
		return sources
	}
	if registry.Sources == nil {
		collector.add("sources.yaml", errors.New("sources must be a map"))
		return sources
	}
	return registry.Sources
}

func loadCourses(fsys fs.FS, collector *errorCollector) []courseFile {
	entries, err := fs.ReadDir(fsys, "courses")
	if err != nil {
		collector.add("courses", err)
		return []courseFile{}
	}

	courses := make([]courseFile, 0, len(entries))
	for _, entry := range entries {
		coursePath := path.Join("courses", entry.Name())
		if !entry.IsDir() {
			collector.add(coursePath, errors.New("expected a course directory"))
			continue
		}

		course := courseFile{path: path.Join(coursePath, "course.yaml")}
		if validateRequiredFile(fsys, course.path, collector) {
			data, readErr := fs.ReadFile(fsys, course.path)
			if readErr != nil {
				collector.add(course.path, readErr)
			} else if decodeErr := decodeYAML(data, &course.metadata); decodeErr != nil {
				collector.add(course.path, decodeErr)
			} else {
				course.metadataOK = true
			}
		}

		modulesPath := path.Join(coursePath, "modules")
		if validateRequiredDirectory(fsys, modulesPath, collector) {
			course.modules = loadModules(fsys, modulesPath, collector)
		}
		courses = append(courses, course)
	}
	return courses
}

func loadModules(fsys fs.FS, modulesPath string, collector *errorCollector) []moduleFile {
	entries, err := fs.ReadDir(fsys, modulesPath)
	if err != nil {
		collector.add(modulesPath, err)
		return []moduleFile{}
	}

	modules := make([]moduleFile, 0, len(entries))
	for _, entry := range entries {
		modulePath := path.Join(modulesPath, entry.Name())
		if !entry.IsDir() {
			collector.add(modulePath, errors.New("expected a module directory"))
			continue
		}

		module := moduleFile{path: path.Join(modulePath, "module.yaml")}
		data, readErr := fs.ReadFile(fsys, module.path)
		if readErr != nil {
			if errors.Is(readErr, fs.ErrNotExist) {
				collector.add(module.path, errors.New("missing module.yaml"))
			} else {
				collector.add(module.path, readErr)
			}
		} else if decodeErr := decodeYAML(data, &module.metadata); decodeErr != nil {
			collector.add(module.path, decodeErr)
		} else {
			module.metadataOK = true
		}

		module.lessons = loadLessons(fsys, modulePath, collector)
		module.worksheets = loadWorksheets(fsys, modulePath, collector)
		module.exercises = loadExercises(fsys, modulePath, collector)
		module.reviewItems = loadReviewItems(fsys, modulePath, collector)
		modules = append(modules, module)
	}
	return modules
}

func loadReviewItems(fsys fs.FS, modulePath string, collector *errorCollector) []reviewItemFile {
	reviewsPath := path.Join(modulePath, "reviews")
	entries, err := fs.ReadDir(fsys, reviewsPath)
	if errors.Is(err, fs.ErrNotExist) {
		return []reviewItemFile{}
	}
	if err != nil {
		collector.add(reviewsPath, err)
		return []reviewItemFile{}
	}

	reviewItems := make([]reviewItemFile, 0, len(entries))
	for _, entry := range entries {
		reviewItemPath := path.Join(reviewsPath, entry.Name())
		if !validateAuthoredYAMLEntry(entry, reviewItemPath, collector) {
			continue
		}
		reviewItem := reviewItemFile{path: reviewItemPath}
		data, readErr := fs.ReadFile(fsys, reviewItem.path)
		if readErr != nil {
			collector.add(reviewItem.path, readErr)
		} else if decodeErr := decodeYAML(data, &reviewItem.metadata); decodeErr != nil {
			collector.add(reviewItem.path, decodeErr)
		} else {
			reviewItem.metadataOK = true
		}
		reviewItems = append(reviewItems, reviewItem)
	}
	sort.Slice(reviewItems, func(i, j int) bool { return reviewItems[i].path < reviewItems[j].path })
	return reviewItems
}

func loadExercises(fsys fs.FS, modulePath string, collector *errorCollector) []exerciseFile {
	exercisesPath := path.Join(modulePath, "exercises")
	entries, err := fs.ReadDir(fsys, exercisesPath)
	if errors.Is(err, fs.ErrNotExist) {
		return []exerciseFile{}
	}
	if err != nil {
		collector.add(exercisesPath, err)
		return []exerciseFile{}
	}

	exercises := make([]exerciseFile, 0, len(entries))
	for _, entry := range entries {
		exercisePath := path.Join(exercisesPath, entry.Name())
		if !validateAuthoredYAMLEntry(entry, exercisePath, collector) {
			continue
		}
		exercise := exerciseFile{path: exercisePath}
		data, readErr := fs.ReadFile(fsys, exercise.path)
		if readErr != nil {
			collector.add(exercise.path, readErr)
		} else if decodeErr := decodeYAML(data, &exercise.metadata); decodeErr != nil {
			collector.add(exercise.path, decodeErr)
		} else {
			exercise.metadataOK = true
		}
		exercises = append(exercises, exercise)
	}
	sort.Slice(exercises, func(i, j int) bool { return exercises[i].path < exercises[j].path })
	return exercises
}

func loadWorksheets(fsys fs.FS, modulePath string, collector *errorCollector) []worksheetFile {
	worksheetsPath := path.Join(modulePath, "worksheets")
	entries, err := fs.ReadDir(fsys, worksheetsPath)
	if errors.Is(err, fs.ErrNotExist) {
		return []worksheetFile{}
	}
	if err != nil {
		collector.add(worksheetsPath, err)
		return []worksheetFile{}
	}

	worksheets := make([]worksheetFile, 0, len(entries))
	for _, entry := range entries {
		worksheetPath := path.Join(worksheetsPath, entry.Name())
		if !validateAuthoredYAMLEntry(entry, worksheetPath, collector) {
			continue
		}
		worksheet := worksheetFile{path: worksheetPath}
		data, readErr := fs.ReadFile(fsys, worksheet.path)
		if readErr != nil {
			collector.add(worksheet.path, readErr)
		} else if decodeErr := decodeYAML(data, &worksheet.metadata); decodeErr != nil {
			collector.add(worksheet.path, decodeErr)
		} else {
			worksheet.metadataOK = true
		}
		worksheets = append(worksheets, worksheet)
	}
	sort.Slice(worksheets, func(i, j int) bool { return worksheets[i].path < worksheets[j].path })
	return worksheets
}

func validateAuthoredYAMLEntry(entry fs.DirEntry, entryPath string, collector *errorCollector) bool {
	if entry.IsDir() {
		collector.add(entryPath, errors.New("unexpected directory; authored YAML files must be direct children"))
		return false
	}
	if path.Ext(entry.Name()) != ".yaml" {
		collector.add(entryPath, errors.New("unexpected entry; expected a .yaml file"))
		return false
	}
	if entry.Type().IsRegular() {
		return true
	}
	info, err := entry.Info()
	if err != nil {
		collector.add(entryPath, fmt.Errorf("cannot inspect authored YAML file: %w", err))
		return false
	}
	if !info.Mode().IsRegular() {
		collector.add(entryPath, errors.New("authored YAML entry is not a regular file"))
		return false
	}
	return true
}

func loadLessons(fsys fs.FS, modulePath string, collector *errorCollector) []lessonFile {
	lessons := []lessonFile{}
	walkErr := fs.WalkDir(fsys, modulePath, func(filePath string, entry fs.DirEntry, err error) error {
		if err != nil {
			collector.add(filePath, err)
			return nil
		}
		if filePath != modulePath && entry.IsDir() {
			return fs.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".mdx") {
			return nil
		}

		lesson := lessonFile{path: filePath}
		data, readErr := fs.ReadFile(fsys, filePath)
		if readErr != nil {
			collector.add(filePath, readErr)
			lessons = append(lessons, lesson)
			return nil
		}
		frontmatter, content, splitErr := splitFrontmatter(data)
		if splitErr != nil {
			collector.add(filePath, splitErr)
			lessons = append(lessons, lesson)
			return nil
		}
		if decodeErr := decodeYAML(frontmatter, &lesson.metadata); decodeErr != nil {
			collector.add(filePath, decodeErr)
		} else {
			lesson.metadataOK = true
			lesson.content = content
		}
		lessons = append(lessons, lesson)
		return nil
	})
	if walkErr != nil {
		collector.add(modulePath, walkErr)
	}
	sort.Slice(lessons, func(i, j int) bool { return lessons[i].path < lessons[j].path })
	return lessons
}

func decodeYAML(data []byte, target any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("multiple YAML documents are not supported")
		}
		return err
	}
	return nil
}

func splitFrontmatter(data []byte) ([]byte, string, error) {
	firstLineEnd, firstNext := lineEnd(data, 0)
	if firstLineEnd < 0 || !bytes.Equal(bytes.TrimSuffix(data[:firstLineEnd], []byte{'\r'}), []byte("---")) {
		return nil, "", errors.New("lesson must begin with YAML frontmatter delimiter ---")
	}

	frontmatterStart := firstNext
	for lineStart := frontmatterStart; lineStart < len(data); {
		lineEndIndex, nextLine := lineEnd(data, lineStart)
		if lineEndIndex < 0 {
			break
		}
		line := bytes.TrimSuffix(data[lineStart:lineEndIndex], []byte{'\r'})
		if bytes.Equal(line, []byte("---")) {
			return data[frontmatterStart:lineStart], string(data[nextLine:]), nil
		}
		lineStart = nextLine
	}
	return nil, "", errors.New("lesson frontmatter is missing closing delimiter ---")
}

func lineEnd(data []byte, start int) (end, next int) {
	if start >= len(data) {
		return -1, -1
	}
	for index := start; index < len(data); index++ {
		if data[index] == '\n' {
			return index, index + 1
		}
	}
	return len(data), len(data)
}

func formatPath(pathName, message string) error {
	return fmt.Errorf("%s: %s", pathName, message)
}
