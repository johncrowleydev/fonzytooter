package curriculum

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadAllowsModulesWithoutReviewItems(t *testing.T) {
	catalog, err := Load(reviewItemCurriculumFS(nil))
	if err != nil {
		t.Fatalf("load curriculum without review items: %v", err)
	}
	if catalog.ReviewItemCount() != 0 || len(catalog.ReviewItemsByModule("course", "module")) != 0 || len(catalog.ReviewItemsByObjective("objective")) != 0 {
		t.Fatalf("expected no review items, got %#v", catalog.ReviewItemsByModule("course", "module"))
	}
}

func TestLoadAcceptsReviewItemWithoutHint(t *testing.T) {
	withoutHint := strings.Replace(validReviewItemYAML(), "hint: A hint.\n", "", 1)
	catalog, err := Load(reviewItemCurriculumFS(map[string]string{"item.yaml": withoutHint}))
	if err != nil {
		t.Fatalf("load review item without hint: %v", err)
	}
	item, ok := catalog.ReviewItemByCourse("course", "module", "review-item")
	if !ok || item.Hint != "" {
		t.Fatalf("unexpected review item without hint: %#v, %v", item, ok)
	}
}

func TestLoadBuildsOrderedImmutableReviewItemCatalog(t *testing.T) {
	later := strings.Replace(validReviewItemYAML(), "id: review-item\n", "id: later\n", 1)
	later = strings.Replace(later, "order: 0\n", "order: 1\n", 1)
	earlier := strings.Replace(validReviewItemYAML(), "id: review-item\n", "id: earlier\n", 1)
	catalog, err := Load(reviewItemCurriculumFS(map[string]string{
		"00-later.yaml":   later,
		"99-earlier.yaml": earlier,
	}))
	if err != nil {
		t.Fatalf("load review items: %v", err)
	}

	items := catalog.ReviewItemsByModule("course", "module")
	if len(items) != 2 || items[0].ID != "earlier" || items[1].ID != "later" {
		t.Fatalf("review item order = %#v", items)
	}
	byObjective := catalog.ReviewItemsByObjective("objective")
	if len(byObjective) != 2 || byObjective[0].ID != "earlier" || byObjective[1].ID != "later" {
		t.Fatalf("objective review item order = %#v", byObjective)
	}
	item, ok := catalog.ReviewItemByCourse("course", "module", "earlier")
	if !ok || item.CourseID != "course" || item.ModuleID != "module" || item.SourceLessonID != "lesson" || item.Hint != "A hint." {
		t.Fatalf("unexpected review item lookup: %#v, %v", item, ok)
	}
	if catalog.ReviewItemCount() != 2 {
		t.Fatalf("review item count = %d, want 2", catalog.ReviewItemCount())
	}
	if _, ok := catalog.ReviewItemByCourse("other", "module", "earlier"); ok {
		t.Fatal("review item lookup ignored course ownership")
	}
	if _, ok := catalog.ReviewItemByCourse("course", "other", "earlier"); ok {
		t.Fatal("review item lookup ignored module ownership")
	}

	items[0].ObjectiveIDs[0] = "mutated"
	byObjective[0].Prompt = "mutated"
	item.ObjectiveIDs[0] = "mutated"
	again, _ := catalog.ReviewItemByCourse("course", "module", "earlier")
	if again.ObjectiveIDs[0] == "mutated" || again.Prompt == "mutated" {
		t.Fatal("catalog exposed mutable review item state")
	}
}

func TestReviewItemOrderingUsesStableIDAsTieBreaker(t *testing.T) {
	zeta := strings.Replace(validReviewItemYAML(), "id: review-item\n", "id: zeta\n", 1)
	alpha := strings.Replace(validReviewItemYAML(), "id: review-item\n", "id: alpha\n", 1)
	catalog, err := Load(reviewItemCurriculumFS(map[string]string{
		"first.yaml":  zeta,
		"second.yaml": alpha,
	}))
	if err != nil {
		t.Fatalf("load tied review items: %v", err)
	}
	items := catalog.ReviewItemsByModule("course", "module")
	if len(items) != 2 || items[0].ID != "alpha" || items[1].ID != "zeta" {
		t.Fatalf("stable review item order = %#v", items)
	}
}

func TestLoadRejectsUnknownReviewItemFields(t *testing.T) {
	yaml := validReviewItemYAML() + "difficulty: hard\n"
	expectLoadError(t, reviewItemCurriculumFS(map[string]string{"item.yaml": yaml}), "not found in type")
}

func TestLoadValidatesReviewItemStructure(t *testing.T) {
	valid := validReviewItemYAML()
	tests := map[string]struct {
		files    map[string]string
		contains string
	}{
		"invalid id":          {map[string]string{"item.yaml": strings.Replace(valid, "id: review-item\n", "id: Invalid Item\n", 1)}, "invalid review item id"},
		"duplicate id":        {map[string]string{"first.yaml": valid, "second.yaml": valid}, "duplicate review item id"},
		"missing order":       {map[string]string{"item.yaml": strings.Replace(valid, "order: 0\n", "", 1)}, "missing required order"},
		"negative order":      {map[string]string{"item.yaml": strings.Replace(valid, "order: 0\n", "order: -1\n", 1)}, "order must be non-negative"},
		"empty objectives":    {map[string]string{"item.yaml": strings.Replace(valid, "objectiveIds:\n  - objective\n", "objectiveIds: []\n", 1)}, "has no objectiveIds"},
		"unknown objective":   {map[string]string{"item.yaml": strings.Replace(valid, "  - objective\n", "  - missing\n", 1)}, "unknown objective id"},
		"empty source lesson": {map[string]string{"item.yaml": strings.Replace(valid, "sourceLessonId: lesson\n", "sourceLessonId: \"\"\n", 1)}, "empty sourceLessonId"},
		"wrong module lesson": {map[string]string{"item.yaml": strings.Replace(valid, "sourceLessonId: lesson\n", "sourceLessonId: missing\n", 1)}, "unknown lesson id"},
		"empty prompt":        {map[string]string{"item.yaml": strings.Replace(valid, "prompt: A question?\n", "prompt: \"\"\n", 1)}, "empty prompt"},
		"empty answer":        {map[string]string{"item.yaml": strings.Replace(valid, "answer: An answer.\n", "answer: \"\"\n", 1)}, "empty answer"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			expectLoadError(t, reviewItemCurriculumFS(test.files), test.contains)
		})
	}
}

func reviewItemCurriculumFS(reviewItems map[string]string) fstest.MapFS {
	fsys := fstest.MapFS{
		"sources.yaml":                              &fstest.MapFile{Data: []byte("sources: {}\n")},
		"courses/course/course.yaml":                &fstest.MapFile{Data: []byte("id: course\ntitle: Course\ndescription: Description.\norder: 0\n")},
		"courses/course/modules/module/module.yaml": &fstest.MapFile{Data: []byte("id: module\ntitle: Module\norder: 0\nobjectives:\n  - id: objective\n    title: Objective\n    description: Description.\n    prerequisites: []\nvideos: []\nlessons:\n  - lesson\n")},
		"courses/course/modules/module/lesson.mdx":  &fstest.MapFile{Data: []byte("---\nid: lesson\ntitle: Lesson\nobjectiveIds:\n  - objective\nsourceIds: []\n---\n# Lesson\n")},
	}
	for name, contents := range reviewItems {
		fsys["courses/course/modules/module/reviews/"+name] = &fstest.MapFile{Data: []byte(contents)}
	}
	return fsys
}

func validReviewItemYAML() string {
	return "id: review-item\n" +
		"order: 0\n" +
		"objectiveIds:\n" +
		"  - objective\n" +
		"sourceLessonId: lesson\n" +
		"prompt: A question?\n" +
		"answer: An answer.\n" +
		"hint: A hint.\n"
}
