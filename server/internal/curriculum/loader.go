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

type sourceAuthoring struct {
	Title string `yaml:"title"`
	URL   string `yaml:"url"`
}

type sourceRegistry struct {
	Sources map[string]sourceAuthoring `yaml:"sources"`
}

type moduleFile struct {
	path       string
	metadata   moduleAuthoring
	lessons    []lessonFile
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
	sources := loadSources(fsys, errors)
	modules := loadModules(fsys, errors)
	validateSources(sources, errors)
	validateModules(modules, sources, errors)

	if err := errors.Err(); err != nil {
		return nil, err
	}

	loadedModules := make([]Module, 0, len(modules))
	loadedSources := make(map[string]Source, len(sources))
	for id, source := range sources {
		loadedSources[id] = Source{ID: id, Title: source.Title, URL: source.URL}
	}

	for _, module := range modules {
		loaded := Module{
			ID:         module.metadata.ID,
			Title:      module.metadata.Title,
			Order:      *module.metadata.Order,
			Objectives: make([]Objective, 0, len(module.metadata.Objectives)),
			Videos:     make([]Video, 0, len(module.metadata.Videos)),
			Lessons:    make([]Lesson, 0, len(module.metadata.Lessons)),
		}
		for _, objective := range module.metadata.Objectives {
			loaded.Objectives = append(loaded.Objectives, Objective{
				ID:            objective.ID,
				Title:         objective.Title,
				Description:   objective.Description,
				Prerequisites: cloneStrings(objective.Prerequisites),
			})
		}
		for _, video := range module.metadata.Videos {
			loaded.Videos = append(loaded.Videos, Video{
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
			loaded.Lessons = append(loaded.Lessons, Lesson{
				ID:           lesson.metadata.ID,
				Title:        lesson.metadata.Title,
				ObjectiveIDs: cloneStrings(lesson.metadata.ObjectiveIDs),
				SourceIDs:    cloneStrings(lesson.metadata.SourceIDs),
				Content:      lesson.content,
			})
		}
		loadedModules = append(loadedModules, loaded)
	}

	return newCatalog(loadedModules, loadedSources), nil
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

func loadModules(fsys fs.FS, collector *errorCollector) []moduleFile {
	entries, err := fs.ReadDir(fsys, "modules")
	if errors.Is(err, fs.ErrNotExist) {
		return []moduleFile{}
	}
	if err != nil {
		collector.add("modules", err)
		return []moduleFile{}
	}

	modules := make([]moduleFile, 0, len(entries))
	for _, entry := range entries {
		modulePath := path.Join("modules", entry.Name())
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
		modules = append(modules, module)
	}
	return modules
}

func loadLessons(fsys fs.FS, modulePath string, collector *errorCollector) []lessonFile {
	lessons := []lessonFile{}
	walkErr := fs.WalkDir(fsys, modulePath, func(filePath string, entry fs.DirEntry, err error) error {
		if err != nil {
			collector.add(filePath, err)
			return nil
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
