package curriculum

import (
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadRejectsUnexpectedEntriesInAuthoredContentDirectories(t *testing.T) {
	for name, entry := range map[string]*fstest.MapFile{
		"unsupported yml":       {Data: []byte("id: ignored\n")},
		"backup file":           {Data: []byte("id: ignored\n")},
		"nested YAML directory": {Mode: fs.ModeDir},
	} {
		t.Run(name, func(t *testing.T) {
			entryName := map[string]string{
				"unsupported yml":       "ignored.yml",
				"backup file":           "ignored.yaml.bak",
				"nested YAML directory": "chapter",
			}[name]
			for _, directory := range []string{"worksheets", "exercises", "reviews"} {
				t.Run(directory, func(t *testing.T) {
					entryPath := "courses/ai-ml/modules/one/" + directory + "/" + entryName
					fsys := curriculumFS(map[string]string{
						"courses/ai-ml/modules/one/module.yaml": moduleYAML("one", 0, emptyModuleLists),
					})
					fsys[entryPath] = entry
					if name == "nested YAML directory" {
						fsys[entryPath+"/foo.yaml"] = &fstest.MapFile{Data: []byte("id: nested\n")}
					}

					_, err := Load(fsys)
					if err == nil {
						t.Fatal("expected unexpected entry error")
					}
					if !strings.Contains(err.Error(), entryPath+": unexpected") {
						t.Fatalf("expected path-qualified unexpected entry error, got:\n%s", err)
					}
				})
			}
		})
	}
}

func TestLoadRejectsDuplicateAuthoredReferences(t *testing.T) {
	fsys := curriculumFS(map[string]string{
		"courses/ai-ml/modules/one/module.yaml":               moduleYAML("one", 0, "objectives:\n  - id: first\n    title: First\n    description: First objective.\n    prerequisites: []\n  - id: second\n    title: Second\n    description: Second objective.\n    prerequisites:\n      - first\n      - first\nvideos:\n  - id: video\n    title: Video\n    url: https://example.com/video\n    objectiveIds:\n      - first\n      - first\nlessons:\n  - lesson\n"),
		"courses/ai-ml/modules/one/lesson.mdx":                "---\nid: lesson\ntitle: Lesson\nobjectiveIds:\n  - first\n  - first\nsourceIds:\n  - source\n  - source\n---\n# Lesson\n",
		"courses/ai-ml/modules/one/worksheets/worksheet.yaml": "id: worksheet\ntitle: Worksheet\nlessonId: lesson\norder: 0\nobjectiveIds:\n  - first\n  - first\ninstructions: Answer.\nproblems:\n  - id: problem\n    prompt: Answer.\n    objectiveIds:\n      - first\n      - first\n    expectedAnswer: Yes.\n    requiresWork: true\n    responseLines: 1\n    rubric:\n      - Correct.\n",
		"courses/ai-ml/modules/one/exercises/exercise.yaml":   "id: exercise\ntitle: Exercise\nlessonId: lesson\norder: 0\nobjectiveIds:\n  - first\n  - first\nprompt: Implement it.\nstarterCode: pass\ntests:\n  - id: test\n    title: Test\n    visibility: visible\n    code: assert True\n",
		"courses/ai-ml/modules/one/reviews/review.yaml":       "id: review\norder: 0\nobjectiveIds:\n  - first\n  - first\nsourceLessonId: lesson\nprompt: Prompt?\nanswer: Answer.\n",
	})
	fsys["sources.yaml"] = &fstest.MapFile{Data: []byte("sources:\n  source:\n    title: Source\n    url: https://example.com/source\n")}

	_, err := Load(fsys)
	if err == nil {
		t.Fatal("expected duplicate reference errors")
	}
	for _, expected := range []string{
		`objective "second" field prerequisites contains duplicate value "first"`,
		`video "video" field objectiveIds contains duplicate value "first"`,
		`lesson "lesson" field objectiveIds contains duplicate value "first"`,
		`lesson "lesson" field sourceIds contains duplicate value "source"`,
		`worksheet "worksheet" field objectiveIds contains duplicate value "first"`,
		`worksheet "worksheet" problem "problem" field objectiveIds contains duplicate value "first"`,
		`exercise "exercise" field objectiveIds contains duplicate value "first"`,
		`review item "review" field objectiveIds contains duplicate value "first"`,
	} {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("expected %q, got:\n%s", expected, err)
		}
	}
}

func TestLoadRejectsMissingObjectiveDescriptionAndNegativeTopLevelOrders(t *testing.T) {
	fsys := fstest.MapFS{
		"sources.yaml":                          &fstest.MapFile{Data: []byte("sources: {}\n")},
		"courses/ai-ml/course.yaml":             &fstest.MapFile{Data: []byte(courseYAML("ai-ml", -1))},
		"courses/ai-ml/modules/one/module.yaml": &fstest.MapFile{Data: []byte(moduleYAML("one", -1, "objectives:\n  - id: objective\n    title: Objective\n    description: '  '\n    prerequisites: []\nvideos: []\nlessons: []\n"))},
	}
	expectLoadError(t, fsys,
		`course "ai-ml" order must be non-negative`,
		`module "one" order must be non-negative`,
		`objective "objective" has empty description`,
	)
}

func TestCatalogReportsUnusedSourcesDeterministically(t *testing.T) {
	fsys := curriculumFS(map[string]string{
		"courses/ai-ml/modules/one/module.yaml": moduleYAML("one", 0, "objectives: []\nvideos: []\nlessons:\n  - lesson\n"),
		"courses/ai-ml/modules/one/lesson.mdx":  "---\nid: lesson\ntitle: Lesson\nobjectiveIds: []\nsourceIds:\n  - used\n---\n# Lesson\n",
	})
	fsys["sources.yaml"] = &fstest.MapFile{Data: []byte("sources:\n  z-unused:\n    title: Z\n    url: https://example.com/z\n  used:\n    title: Used\n    url: https://example.com/used\n  a-unused:\n    title: A\n    url: https://example.com/a\n")}

	catalog, err := Load(fsys)
	if err != nil {
		t.Fatalf("load curriculum: %v", err)
	}
	if got := strings.Join(catalog.UnusedSourceIDs(), ","); got != "a-unused,z-unused" {
		t.Fatalf("unexpected unused source IDs: %s", got)
	}
}
