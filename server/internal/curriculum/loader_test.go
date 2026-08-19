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
			fsys:     fstest.MapFS{"modules": directory},
			contains: "sources.yaml: required file is missing",
		},
		"missing modules": {
			fsys: fstest.MapFS{
				"sources.yaml": &fstest.MapFile{Data: []byte("sources: {}\n")},
			},
			contains: "modules: required directory is missing",
		},
		"sources is a directory": {
			fsys: fstest.MapFS{
				"sources.yaml": directory,
				"modules":      &fstest.MapFile{Mode: fs.ModeDir},
			},
			contains: "sources.yaml: required file is a directory",
		},
		"modules is a file": {
			fsys: fstest.MapFS{
				"sources.yaml": &fstest.MapFile{Data: []byte("sources: {}\n")},
				"modules":      &fstest.MapFile{Data: []byte("not a directory")},
			},
			contains: "modules: required directory is not a directory",
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
		"modules":      &fstest.MapFile{Mode: fs.ModeDir},
	})
	if err != nil {
		t.Fatalf("load empty curriculum: %v", err)
	}
	if catalog.ModuleCount() != 0 || catalog.SourceCount() != 0 {
		t.Fatalf("expected empty catalog, got %d modules and %d sources", catalog.ModuleCount(), catalog.SourceCount())
	}
}

func TestLoadUsesDeclaredOrderAndFrontmatterIDs(t *testing.T) {
	fsys := fstest.MapFS{
		"sources.yaml":                       &fstest.MapFile{Data: []byte("sources:\n  go-docs:\n    title: Go documentation\n    url: https://go.dev/doc/\n")},
		"modules/10-second/module.yaml":      &fstest.MapFile{Data: []byte("id: second\ntitle: Second\norder: 1\nobjectives:\n  - id: second.objective\n    title: Second objective\n    prerequisites: []\nvideos: []\nlessons:\n  - second.lesson\n")},
		"modules/10-second/storage-name.mdx": &fstest.MapFile{Data: []byte("---\r\nid: second.lesson\r\ntitle: Second lesson\r\nobjectiveIds: []\r\nsourceIds:\r\n  - go-docs\r\n---\r\n# Second lesson\r\n")},
		"modules/00-first/module.yaml":       &fstest.MapFile{Data: []byte("id: first\ntitle: First\norder: 0\nobjectives: []\nvideos: []\nlessons:\n  - first.lesson\n")},
		"modules/00-first/storage-name.mdx":  &fstest.MapFile{Data: []byte("---\nid: first.lesson\ntitle: First lesson\nobjectiveIds:\n  - second.objective\nsourceIds: []\n---\n# First lesson\n")},
	}

	catalog, err := Load(fsys)
	if err != nil {
		t.Fatalf("load curriculum: %v", err)
	}
	modules := catalog.Modules()
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
	again, ok := catalog.LessonByID("first", "first.lesson")
	if !ok || again.Content == "mutated" {
		t.Fatal("catalog exposed mutable lesson state")
	}
	first, ok := catalog.ModuleByID("first")
	if !ok || len(first.Objectives) != 0 {
		t.Fatalf("catalog exposed mutable module state: %#v", first)
	}
}

func TestLoadAggregatesDeterministicAuthoringErrors(t *testing.T) {
	fsys := fstest.MapFS{
		"sources.yaml":                   &fstest.MapFile{Data: []byte("sources:\n  bad source:\n    title: \"\"\n    url: ftp://example.com\n")},
		"modules/01-invalid/module.yaml": &fstest.MapFile{Data: []byte("id: Invalid\ntitle: \"\"\norder: 0\nobjectives: []\nvideos:\n  - id: bad video\n    title: \"\"\n    url: ftp://example.com\nlessons:\n  - missing.lesson\n")},
		"modules/01-invalid/orphan.mdx":  &fstest.MapFile{Data: []byte("---\nid: orphan.lesson\ntitle: Orphan\nobjectiveIds:\n  - missing.objective\nsourceIds:\n  - missing.source\n---\n   \n")},
	}

	_, err := Load(fsys)
	if err == nil {
		t.Fatal("expected invalid curriculum")
	}
	message := err.Error()
	for _, expected := range []string{
		"modules/01-invalid/module.yaml:",
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
	if strings.Index(message, "modules/01-invalid/module.yaml:") > strings.Index(message, "sources.yaml:") {
		t.Fatalf("expected deterministic path ordering, got:\n%s", message)
	}
}

func TestLoadRejectsUnknownLessonFrontmatterAndPrerequisiteCycles(t *testing.T) {
	fsys := fstest.MapFS{
		"sources.yaml":               &fstest.MapFile{Data: []byte("sources: {}\n")},
		"modules/module/module.yaml": &fstest.MapFile{Data: []byte("id: module\ntitle: Module\norder: 0\nobjectives:\n  - id: a\n    title: A\n    prerequisites:\n      - b\n  - id: b\n    title: B\n    prerequisites:\n      - c\n  - id: c\n    title: C\n    prerequisites:\n      - a\nvideos: []\nlessons:\n  - lesson\n")},
		"modules/module/lesson.mdx":  &fstest.MapFile{Data: []byte("---\nid: lesson\ntitle: Lesson\nobjetiveIds: []\nsourceIds: []\n---\n# Lesson\n")},
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
		"modules/first/module.yaml":  moduleYAML("same", 0, emptyModuleLists),
		"modules/second/module.yaml": moduleYAML("same", 1, emptyModuleLists),
	}), `duplicate module id "same"`)
}

func TestLoadRejectsDuplicateModuleOrders(t *testing.T) {
	expectLoadError(t, curriculumFS(map[string]string{
		"modules/first/module.yaml":  moduleYAML("first", 0, emptyModuleLists),
		"modules/second/module.yaml": moduleYAML("second", 0, emptyModuleLists),
	}), "duplicate module order 0")
}

func TestLoadRejectsUnknownModuleFields(t *testing.T) {
	expectLoadError(t, curriculumFS(map[string]string{
		"modules/one/module.yaml": moduleYAML("one", 0, "unknown: true\n"+emptyModuleLists),
	}), "not found in type")
}

func TestLoadRejectsDuplicateObjectiveIDsAcrossModules(t *testing.T) {
	objective := "objectives:\n  - id: shared\n    title: Shared\n    prerequisites: []\nvideos: []\nlessons: []\n"
	expectLoadError(t, curriculumFS(map[string]string{
		"modules/first/module.yaml":  moduleYAML("first", 0, objective),
		"modules/second/module.yaml": moduleYAML("second", 1, objective),
	}), `duplicate objective id "shared"`)
}

func TestLoadRejectsSelfPrerequisite(t *testing.T) {
	expectLoadError(t, curriculumFS(map[string]string{
		"modules/one/module.yaml": moduleYAML("one", 0, "objectives:\n  - id: foo\n    title: Foo\n    prerequisites:\n      - foo\nvideos: []\nlessons: []\n"),
	}), `objective "foo" cannot list itself as a prerequisite`)
}

func TestLoadRejectsUnknownPrerequisite(t *testing.T) {
	expectLoadError(t, curriculumFS(map[string]string{
		"modules/one/module.yaml": moduleYAML("one", 0, "objectives:\n  - id: foo\n    title: Foo\n    prerequisites:\n      - bar\nvideos: []\nlessons: []\n"),
	}), `objective "foo" has unknown prerequisite "bar"`)
}

func TestLoadRejectsUnknownVideoObjective(t *testing.T) {
	expectLoadError(t, curriculumFS(map[string]string{
		"modules/one/module.yaml": moduleYAML("one", 0, "objectives: []\nvideos:\n  - id: intro\n    title: Intro\n    url: https://example.com/intro\n    objectiveIds:\n      - missing\nlessons: []\n"),
	}), `video "intro" has unknown objective id "missing"`)
}

func TestLoadRejectsDuplicateVideoIDsWithinModule(t *testing.T) {
	expectLoadError(t, curriculumFS(map[string]string{
		"modules/one/module.yaml": moduleYAML("one", 0, "objectives: []\nvideos:\n  - id: intro\n    title: Intro\n    url: https://example.com/intro\n    objectiveIds: []\n  - id: intro\n    title: Another intro\n    url: https://example.com/another-intro\n    objectiveIds: []\nlessons: []\n"),
	}), `duplicate video id "intro"`)
}

func TestLoadRejectsDuplicateFrontmatterLessonIDs(t *testing.T) {
	expectLoadError(t, curriculumFS(map[string]string{
		"modules/one/module.yaml": moduleYAML("one", 0, "objectives: []\nvideos: []\nlessons:\n  - lesson\n"),
		"modules/one/first.mdx":   lessonMDX("lesson"),
		"modules/one/second.mdx":  lessonMDX("lesson"),
	}), `duplicate lesson id "lesson"`)
}

func TestLoadRejectsDuplicateLessonIDReferences(t *testing.T) {
	expectLoadError(t, curriculumFS(map[string]string{
		"modules/one/module.yaml": moduleYAML("one", 0, "objectives: []\nvideos: []\nlessons:\n  - lesson\n  - lesson\n"),
		"modules/one/lesson.mdx":  lessonMDX("lesson"),
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
		"sources.yaml": &fstest.MapFile{Data: []byte("sources: {}\n")},
		"modules":      &fstest.MapFile{Mode: fs.ModeDir},
	}
	for path, contents := range files {
		fsys[path] = &fstest.MapFile{Data: []byte(contents)}
	}
	return fsys
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
