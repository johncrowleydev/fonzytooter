package curriculum

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadRequiresCurriculumRootStructure(t *testing.T) {
	directory := &fstest.MapFile{Mode: fs.ModeDir}
	tests := map[string]struct {
		fsys     fstest.MapFS
		contains string
	}{
		"empty filesystem": {
			fsys:     fstest.MapFS{},
			contains: "sources.yaml: required file is missing",
		},
		"missing sources": {
			fsys:     fstest.MapFS{"courses": directory},
			contains: "sources.yaml: required file is missing",
		},
		"missing courses": {
			fsys: fstest.MapFS{
				"sources.yaml": &fstest.MapFile{Data: []byte("sources: {}\n")},
			},
			contains: "courses: required directory is missing",
		},
		"sources is a directory": {
			fsys: fstest.MapFS{
				"sources.yaml": directory,
				"courses":      directory,
			},
			contains: "sources.yaml: required file is a directory",
		},
		"courses is a file": {
			fsys: fstest.MapFS{
				"sources.yaml": &fstest.MapFile{Data: []byte("sources: {}\n")},
				"courses":      &fstest.MapFile{Data: []byte("not a directory")},
			},
			contains: "courses: required directory is not a directory",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			expectLoadError(t, test.fsys, test.contains)
		})
	}
}

func TestLoadAcceptsValidEmptyCurriculumRoot(t *testing.T) {
	catalog, err := Load(fstest.MapFS{
		"sources.yaml": &fstest.MapFile{Data: []byte("sources: {}\n")},
		"courses":      &fstest.MapFile{Mode: fs.ModeDir},
	})
	if err != nil {
		t.Fatalf("load empty curriculum: %v", err)
	}
	if catalog.CourseCount() != 0 || catalog.ModuleCount() != 0 || catalog.SourceCount() != 0 {
		t.Fatalf("expected empty catalog, got %d courses, %d modules, and %d sources", catalog.CourseCount(), catalog.ModuleCount(), catalog.SourceCount())
	}
}

func TestLoadUsesDeclaredOrderAndFrontmatterIDs(t *testing.T) {
	fsys := fstest.MapFS{
		"sources.yaml":                                     &fstest.MapFile{Data: []byte("sources:\n  go-docs:\n    title: Go documentation\n    url: https://go.dev/doc/\n")},
		"courses/ai-ml/course.yaml":                        &fstest.MapFile{Data: []byte(courseYAML("ai-ml", 0))},
		"courses/ai-ml/modules/10-second/module.yaml":      &fstest.MapFile{Data: []byte("id: second\ntitle: Second\norder: 1\nobjectives:\n  - id: second.objective\n    title: Second objective\n    prerequisites: []\nvideos: []\nlessons:\n  - second.lesson\n")},
		"courses/ai-ml/modules/10-second/storage-name.mdx": &fstest.MapFile{Data: []byte("---\r\nid: second.lesson\r\ntitle: Second lesson\r\nobjectiveIds: []\r\nsourceIds:\r\n  - go-docs\r\n---\r\n# Second lesson\r\n")},
		"courses/ai-ml/modules/00-first/module.yaml":       &fstest.MapFile{Data: []byte("id: first\ntitle: First\norder: 0\nobjectives: []\nvideos: []\nlessons:\n  - first.lesson\n")},
		"courses/ai-ml/modules/00-first/storage-name.mdx":  &fstest.MapFile{Data: []byte("---\nid: first.lesson\ntitle: First lesson\nobjectiveIds:\n  - second.objective\nsourceIds: []\n---\n# First lesson\n")},
	}

	catalog, err := Load(fsys)
	if err != nil {
		t.Fatalf("load curriculum: %v", err)
	}
	modules := catalog.ModulesByCourse("ai-ml")
	if len(modules) != 2 || modules[0].ID != "first" || modules[1].ID != "second" {
		t.Fatalf("unexpected module order: %#v", modules)
	}
	if modules[0].Lessons[0].ID != "first.lesson" {
		t.Fatalf("expected frontmatter lesson ID, got %#v", modules[0].Lessons)
	}
	if got := modules[1].Lessons[0].Content; got != "# Second lesson\r\n" {
		t.Fatalf("expected untouched CRLF body, got %q", got)
	}
	if _, ok := catalog.ObjectiveByID("second.objective"); !ok {
		t.Fatal("expected global objective lookup")
	}
	if source, ok := catalog.SourceByID("go-docs"); !ok || source.URL != "https://go.dev/doc/" {
		t.Fatalf("expected source lookup, got %#v, %v", source, ok)
	}

	modules[0].Lessons[0].Content = "mutated"
	modules[0].Objectives = append(modules[0].Objectives, Objective{ID: "mutated"})
	again, ok := catalog.LessonByCourse("ai-ml", "first", "first.lesson")
	if !ok || again.Content == "mutated" {
		t.Fatal("catalog exposed mutable lesson state")
	}
	first, ok := catalog.ModuleByCourse("ai-ml", "first")
	if !ok || len(first.Objectives) != 0 {
		t.Fatalf("catalog exposed mutable module state: %#v", first)
	}
}

func TestLoadAggregatesDeterministicAuthoringErrors(t *testing.T) {
	fsys := fstest.MapFS{
		"sources.yaml":                                 &fstest.MapFile{Data: []byte("sources:\n  bad source:\n    title: \"\"\n    url: ftp://example.com\n")},
		"courses/ai-ml/course.yaml":                    &fstest.MapFile{Data: []byte(courseYAML("ai-ml", 0))},
		"courses/ai-ml/modules/01-invalid/module.yaml": &fstest.MapFile{Data: []byte("id: Invalid\ntitle: \"\"\norder: 0\nobjectives: []\nvideos:\n  - id: bad video\n    title: \"\"\n    url: ftp://example.com\nlessons:\n  - missing.lesson\n")},
		"courses/ai-ml/modules/01-invalid/orphan.mdx":  &fstest.MapFile{Data: []byte("---\nid: orphan.lesson\ntitle: Orphan\nobjectiveIds:\n  - missing.objective\nsourceIds:\n  - missing.source\n---\n   \n")},
	}

	_, err := Load(fsys)
	if err == nil {
		t.Fatal("expected invalid curriculum")
	}
	message := err.Error()
	for _, expected := range []string{
		"courses/ai-ml/modules/01-invalid/module.yaml:",
		"invalid source id \"bad source\"",
		"unknown objective id \"missing.objective\"",
		"unknown source id \"missing.source\"",
		"empty content body",
		"not declared",
	} {
		if !strings.Contains(message, expected) {
			t.Errorf("expected validation error to contain %q, got:\n%s", expected, message)
		}
	}
	if strings.Index(message, "courses/ai-ml/modules/01-invalid/module.yaml:") > strings.Index(message, "sources.yaml:") {
		t.Fatalf("expected deterministic path ordering, got:\n%s", message)
	}
}

func TestLoadRejectsUnknownLessonFrontmatterAndPrerequisiteCycles(t *testing.T) {
	fsys := fstest.MapFS{
		"sources.yaml":                             &fstest.MapFile{Data: []byte("sources: {}\n")},
		"courses/ai-ml/course.yaml":                &fstest.MapFile{Data: []byte(courseYAML("ai-ml", 0))},
		"courses/ai-ml/modules/module/module.yaml": &fstest.MapFile{Data: []byte("id: module\ntitle: Module\norder: 0\nobjectives:\n  - id: a\n    title: A\n    prerequisites:\n      - b\n  - id: b\n    title: B\n    prerequisites:\n      - c\n  - id: c\n    title: C\n    prerequisites:\n      - a\nvideos: []\nlessons:\n  - lesson\n")},
		"courses/ai-ml/modules/module/lesson.mdx":  &fstest.MapFile{Data: []byte("---\nid: lesson\ntitle: Lesson\nobjetiveIds: []\nsourceIds: []\n---\n# Lesson\n")},
	}

	_, err := Load(fsys)
	if err == nil {
		t.Fatal("expected invalid curriculum")
	}
	message := err.Error()
	if !strings.Contains(message, "not found in type") || !strings.Contains(message, "prerequisite cycle") {
		t.Fatalf("expected strict frontmatter and cycle errors, got:\n%s", message)
	}
}

func TestLoadRejectsDuplicateModuleIDs(t *testing.T) {
	expectLoadError(t, curriculumFS(map[string]string{
		"courses/ai-ml/modules/first/module.yaml":  moduleYAML("same", 0, emptyModuleLists),
		"courses/ai-ml/modules/second/module.yaml": moduleYAML("same", 1, emptyModuleLists),
	}), `duplicate module id "same"`)
}

func TestLoadRejectsDuplicateModuleIDsAcrossCourses(t *testing.T) {
	expectLoadError(t, fstest.MapFS{
		"sources.yaml":                           &fstest.MapFile{Data: []byte("sources: {}\n")},
		"courses/first/course.yaml":              &fstest.MapFile{Data: []byte(courseYAML("first", 0))},
		"courses/first/modules/one/module.yaml":  &fstest.MapFile{Data: []byte(moduleYAML("same", 0, emptyModuleLists))},
		"courses/second/course.yaml":             &fstest.MapFile{Data: []byte(courseYAML("second", 1))},
		"courses/second/modules/two/module.yaml": &fstest.MapFile{Data: []byte(moduleYAML("same", 0, emptyModuleLists))},
	}, `duplicate module id "same"`)
}

func TestLoadRejectsDuplicateModuleOrders(t *testing.T) {
	expectLoadError(t, curriculumFS(map[string]string{
		"courses/ai-ml/modules/first/module.yaml":  moduleYAML("first", 0, emptyModuleLists),
		"courses/ai-ml/modules/second/module.yaml": moduleYAML("second", 0, emptyModuleLists),
	}), "duplicate module order 0")
}

func TestLoadRejectsUnknownModuleFields(t *testing.T) {
	expectLoadError(t, curriculumFS(map[string]string{
		"courses/ai-ml/modules/one/module.yaml": moduleYAML("one", 0, "unknown: true\n"+emptyModuleLists),
	}), "not found in type")
}

func TestLoadRejectsDuplicateObjectiveIDsAcrossModules(t *testing.T) {
	objective := "objectives:\n  - id: shared\n    title: Shared\n    prerequisites: []\nvideos: []\nlessons: []\n"
	expectLoadError(t, curriculumFS(map[string]string{
		"courses/ai-ml/modules/first/module.yaml":  moduleYAML("first", 0, objective),
		"courses/ai-ml/modules/second/module.yaml": moduleYAML("second", 1, objective),
	}), `duplicate objective id "shared"`)
}

func TestLoadRejectsSelfPrerequisite(t *testing.T) {
	expectLoadError(t, curriculumFS(map[string]string{
		"courses/ai-ml/modules/one/module.yaml": moduleYAML("one", 0, "objectives:\n  - id: foo\n    title: Foo\n    prerequisites:\n      - foo\nvideos: []\nlessons: []\n"),
	}), `objective "foo" cannot list itself as a prerequisite`)
}

func TestLoadRejectsUnknownPrerequisite(t *testing.T) {
	expectLoadError(t, curriculumFS(map[string]string{
		"courses/ai-ml/modules/one/module.yaml": moduleYAML("one", 0, "objectives:\n  - id: foo\n    title: Foo\n    prerequisites:\n      - bar\nvideos: []\nlessons: []\n"),
	}), `objective "foo" has unknown prerequisite "bar"`)
}

func TestLoadRejectsUnknownVideoObjective(t *testing.T) {
	expectLoadError(t, curriculumFS(map[string]string{
		"courses/ai-ml/modules/one/module.yaml": moduleYAML("one", 0, "objectives: []\nvideos:\n  - id: intro\n    title: Intro\n    url: https://example.com/intro\n    objectiveIds:\n      - missing\nlessons: []\n"),
	}), `video "intro" has unknown objective id "missing"`)
}

func TestLoadRejectsDuplicateVideoIDsWithinModule(t *testing.T) {
	expectLoadError(t, curriculumFS(map[string]string{
		"courses/ai-ml/modules/one/module.yaml": moduleYAML("one", 0, "objectives: []\nvideos:\n  - id: intro\n    title: Intro\n    url: https://example.com/intro\n    objectiveIds: []\n  - id: intro\n    title: Another intro\n    url: https://example.com/another-intro\n    objectiveIds: []\nlessons: []\n"),
	}), `duplicate video id "intro"`)
}

func TestLoadRejectsDuplicateFrontmatterLessonIDs(t *testing.T) {
	expectLoadError(t, curriculumFS(map[string]string{
		"courses/ai-ml/modules/one/module.yaml": moduleYAML("one", 0, "objectives: []\nvideos: []\nlessons:\n  - lesson\n"),
		"courses/ai-ml/modules/one/first.mdx":   lessonMDX("lesson"),
		"courses/ai-ml/modules/one/second.mdx":  lessonMDX("lesson"),
	}), `duplicate lesson id "lesson"`)
}

func TestLoadRejectsDuplicateLessonIDReferences(t *testing.T) {
	expectLoadError(t, curriculumFS(map[string]string{
		"courses/ai-ml/modules/one/module.yaml": moduleYAML("one", 0, "objectives: []\nvideos: []\nlessons:\n  - lesson\n  - lesson\n"),
		"courses/ai-ml/modules/one/lesson.mdx":  lessonMDX("lesson"),
	}), `duplicate lesson id reference "lesson"`)
}

func TestSplitFrontmatterRequiresDelimiters(t *testing.T) {
	for name, input := range map[string]string{
		"missing opening": "# Lesson\n",
		"missing closing": "---\nid: lesson\n",
	} {
		t.Run(name, func(t *testing.T) {
			_, _, err := splitFrontmatter([]byte(input))
			if err == nil {
				t.Fatal("expected frontmatter error")
			}
		})
	}
}

const emptyModuleLists = "objectives: []\nvideos: []\nlessons: []\n"

func curriculumFS(files map[string]string) fstest.MapFS {
	fsys := fstest.MapFS{
		"sources.yaml":              &fstest.MapFile{Data: []byte("sources: {}\n")},
		"courses/ai-ml/course.yaml": &fstest.MapFile{Data: []byte(courseYAML("ai-ml", 0))},
	}
	for path, contents := range files {
		fsys[path] = &fstest.MapFile{Data: []byte(contents)}
	}
	return fsys
}

func TestLoadBuildsImmutableMultiCourseCatalog(t *testing.T) {
	fsys := fstest.MapFS{
		"sources.yaml":                                  &fstest.MapFile{Data: []byte("sources:\n  shared-source:\n    title: Shared source\n    url: https://example.com/shared\n")},
		"courses/second/course.yaml":                    &fstest.MapFile{Data: []byte(courseYAML("second-course", 0))},
		"courses/second/modules/beta/module.yaml":       &fstest.MapFile{Data: []byte(moduleYAML("beta", 0, "objectives:\n  - id: beta.objective\n    title: Beta objective\n    prerequisites: []\nvideos: []\nlessons:\n  - shared-lesson\n"))},
		"courses/second/modules/beta/lesson.mdx":        &fstest.MapFile{Data: []byte("---\nid: shared-lesson\ntitle: Beta lesson\nobjectiveIds:\n  - beta.objective\nsourceIds:\n  - shared-source\n---\n# Beta\n")},
		"courses/first/course.yaml":                     &fstest.MapFile{Data: []byte(courseYAML("first-course", 1))},
		"courses/first/modules/alpha-later/module.yaml": &fstest.MapFile{Data: []byte(moduleYAML("alpha-later", 1, emptyModuleLists))},
		"courses/first/modules/alpha/module.yaml":       &fstest.MapFile{Data: []byte(moduleYAML("alpha", 0, "objectives:\n  - id: alpha.objective\n    title: Alpha objective\n    prerequisites:\n      - beta.objective\nvideos: []\nlessons:\n  - shared-lesson\n"))},
		"courses/first/modules/alpha/lesson.mdx":        &fstest.MapFile{Data: []byte("---\nid: shared-lesson\ntitle: Alpha lesson\nobjectiveIds:\n  - alpha.objective\nsourceIds:\n  - shared-source\n---\n# Alpha\n")},
	}

	catalog, err := Load(fsys)
	if err != nil {
		t.Fatalf("load multi-course curriculum: %v", err)
	}

	courses := catalog.Courses()
	if len(courses) != 2 || courses[0].ID != "second-course" || courses[1].ID != "first-course" {
		t.Fatalf("unexpected course order: %#v", courses)
	}
	firstModules := catalog.ModulesByCourse("first-course")
	if len(firstModules) != 2 || firstModules[0].ID != "alpha" || firstModules[1].ID != "alpha-later" {
		t.Fatalf("unexpected first-course modules: %#v", firstModules)
	}
	secondModules := catalog.ModulesByCourse("second-course")
	if len(secondModules) != 1 || secondModules[0].ID != "beta" {
		t.Fatalf("unexpected second-course modules: %#v", secondModules)
	}
	if firstModules[0].Order != 0 || secondModules[0].Order != 0 {
		t.Fatalf("expected course-scoped module order zero, got %#v and %#v", firstModules[0], secondModules[0])
	}

	if _, ok := catalog.ModuleByCourse("first-course", "beta"); ok {
		t.Fatal("course-aware module lookup returned another course's module")
	}
	if _, ok := catalog.LessonByCourse("first-course", "beta", "shared-lesson"); ok {
		t.Fatal("course-aware lesson lookup returned a lesson under the wrong course")
	}
	alphaLesson, ok := catalog.LessonByCourse("first-course", "alpha", "shared-lesson")
	if !ok || alphaLesson.CourseID != "first-course" || alphaLesson.ModuleID != "alpha" || alphaLesson.Content != "# Alpha\n" {
		t.Fatalf("unexpected course-aware lesson: %#v, %v", alphaLesson, ok)
	}
	if _, ok := catalog.LessonByCourse("first-course", "alpha-later", "shared-lesson"); ok {
		t.Fatal("course-aware lesson lookup returned a lesson under the wrong module")
	}
	if source, ok := catalog.SourceByID("shared-source"); !ok || source.Title != "Shared source" {
		t.Fatalf("expected globally shared source, got %#v, %v", source, ok)
	}
	if objective, ok := catalog.ObjectiveByID("alpha.objective"); !ok || objective.CourseID != "first-course" || objective.ModuleID != "alpha" {
		t.Fatalf("expected course-qualified global objective, got %#v, %v", objective, ok)
	}

	courses[0].Title = "mutated"
	courses[0].Modules[0].Lessons[0].Content = "mutated"
	firstModules[0].Objectives[0].Prerequisites[0] = "mutated"
	alphaLesson.ObjectiveIDs[0] = "mutated"
	again, _ := catalog.CourseByID("second-course")
	againLesson, _ := catalog.LessonByCourse("first-course", "alpha", "shared-lesson")
	againObjective, _ := catalog.ObjectiveByID("alpha.objective")
	if again.Title == "mutated" || again.Modules[0].Lessons[0].Content == "mutated" || againLesson.ObjectiveIDs[0] == "mutated" || againObjective.Prerequisites[0] == "mutated" {
		t.Fatal("catalog exposed mutable course, module, lesson, or objective state")
	}
}

func TestLoadValidatesPrerequisiteCyclesAcrossCourses(t *testing.T) {
	fsys := fstest.MapFS{
		"sources.yaml":                              &fstest.MapFile{Data: []byte("sources: {}\n")},
		"courses/first/course.yaml":                 &fstest.MapFile{Data: []byte(courseYAML("first", 0))},
		"courses/first/modules/first/module.yaml":   &fstest.MapFile{Data: []byte(moduleYAML("first-module", 0, "objectives:\n  - id: first.objective\n    title: First\n    prerequisites:\n      - second.objective\nvideos: []\nlessons: []\n"))},
		"courses/second/course.yaml":                &fstest.MapFile{Data: []byte(courseYAML("second", 1))},
		"courses/second/modules/second/module.yaml": &fstest.MapFile{Data: []byte(moduleYAML("second-module", 0, "objectives:\n  - id: second.objective\n    title: Second\n    prerequisites:\n      - first.objective\nvideos: []\nlessons: []\n"))},
	}

	expectLoadError(t, fsys, "objective prerequisite cycle", "courses/")
}

func TestLoadRejectsInvalidCourseMetadataDeterministically(t *testing.T) {
	fsys := fstest.MapFS{
		"sources.yaml":                &fstest.MapFile{Data: []byte("sources: {}\n")},
		"courses/invalid/course.yaml": &fstest.MapFile{Data: []byte("id: Invalid Course\ntitle: \"\"\ndescription: \"\"\n")},
		"courses/invalid/modules":     &fstest.MapFile{Mode: fs.ModeDir},
	}

	_, err := Load(fsys)
	if err == nil {
		t.Fatal("expected invalid course metadata")
	}
	want := "curriculum validation failed:\n" +
		"- courses/invalid/course.yaml: course \"Invalid Course\" has empty description\n" +
		"- courses/invalid/course.yaml: course \"Invalid Course\" has empty title\n" +
		"- courses/invalid/course.yaml: course \"Invalid Course\" is missing required order\n" +
		"- courses/invalid/course.yaml: invalid course id \"Invalid Course\""
	if err.Error() != want {
		t.Fatalf("unexpected deterministic errors:\n%s\nwant:\n%s", err, want)
	}
}

func TestLoadRejectsUnknownCourseFields(t *testing.T) {
	expectLoadError(t, fstest.MapFS{
		"sources.yaml":              &fstest.MapFile{Data: []byte("sources: {}\n")},
		"courses/ai-ml/course.yaml": &fstest.MapFile{Data: []byte(courseYAML("ai-ml", 0) + "semester: fall\n")},
		"courses/ai-ml/modules":     &fstest.MapFile{Mode: fs.ModeDir},
	}, "courses/ai-ml/course.yaml:", "not found in type")
}

func TestLoadRejectsMissingCourseMetadataAndModules(t *testing.T) {
	t.Run("missing course.yaml", func(t *testing.T) {
		expectLoadError(t, fstest.MapFS{
			"sources.yaml":          &fstest.MapFile{Data: []byte("sources: {}\n")},
			"courses/ai-ml/modules": &fstest.MapFile{Mode: fs.ModeDir},
		}, "courses/ai-ml/course.yaml: required file is missing")
	})

	t.Run("missing modules", func(t *testing.T) {
		expectLoadError(t, fstest.MapFS{
			"sources.yaml":              &fstest.MapFile{Data: []byte("sources: {}\n")},
			"courses/ai-ml/course.yaml": &fstest.MapFile{Data: []byte(courseYAML("ai-ml", 0))},
		}, "courses/ai-ml/modules: required directory is missing")
	})

	t.Run("modules is not a directory", func(t *testing.T) {
		expectLoadError(t, fstest.MapFS{
			"sources.yaml":              &fstest.MapFile{Data: []byte("sources: {}\n")},
			"courses/ai-ml/course.yaml": &fstest.MapFile{Data: []byte(courseYAML("ai-ml", 0))},
			"courses/ai-ml/modules":     &fstest.MapFile{Data: []byte("not a directory")},
		}, "courses/ai-ml/modules: required directory is not a directory")
	})
}

func TestLoadRejectsDuplicateCourseIDsAndOrders(t *testing.T) {
	t.Run("duplicate IDs", func(t *testing.T) {
		expectLoadError(t, fstest.MapFS{
			"sources.yaml":               &fstest.MapFile{Data: []byte("sources: {}\n")},
			"courses/first/course.yaml":  &fstest.MapFile{Data: []byte(courseYAML("same", 0))},
			"courses/first/modules":      &fstest.MapFile{Mode: fs.ModeDir},
			"courses/second/course.yaml": &fstest.MapFile{Data: []byte(courseYAML("same", 1))},
			"courses/second/modules":     &fstest.MapFile{Mode: fs.ModeDir},
		}, `duplicate course id "same"`)
	})

	t.Run("duplicate orders", func(t *testing.T) {
		expectLoadError(t, fstest.MapFS{
			"sources.yaml":               &fstest.MapFile{Data: []byte("sources: {}\n")},
			"courses/first/course.yaml":  &fstest.MapFile{Data: []byte(courseYAML("first", 0))},
			"courses/first/modules":      &fstest.MapFile{Mode: fs.ModeDir},
			"courses/second/course.yaml": &fstest.MapFile{Data: []byte(courseYAML("second", 0))},
			"courses/second/modules":     &fstest.MapFile{Mode: fs.ModeDir},
		}, "duplicate course order 0")
	})
}

func courseYAML(id string, order int) string {
	return fmt.Sprintf("id: %s\ntitle: %s\ndescription: Learn %s.\norder: %d\n", id, id, id, order)
}

func moduleYAML(id string, order int, contents string) string {
	return fmt.Sprintf("id: %s\ntitle: %s\norder: %d\n%s", id, id, order, contents)
}

func lessonMDX(id string) string {
	return fmt.Sprintf("---\nid: %s\ntitle: Lesson\nobjectiveIds: []\nsourceIds: []\n---\n# Lesson\n", id)
}

func expectLoadError(t *testing.T, fsys fs.FS, contains ...string) {
	t.Helper()
	_, err := Load(fsys)
	if err == nil {
		t.Fatal("expected curriculum load error")
	}
	message := err.Error()
	for _, expected := range contains {
		if !strings.Contains(message, expected) {
			t.Errorf("expected load error to contain %q, got:\n%s", expected, message)
		}
	}
}
