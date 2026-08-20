package curriculum

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadAllowsModulesWithoutExercises(t *testing.T) {
	catalog, err := Load(exerciseCurriculumFS(nil))
	if err != nil {
		t.Fatalf("load curriculum without exercises: %v", err)
	}
	if catalog.ExerciseCount() != 0 || len(catalog.ExercisesByModule("course", "module")) != 0 || len(catalog.ExercisesByLesson("course", "module", "lesson-one")) != 0 {
		t.Fatalf("expected no exercises, got %#v", catalog.ExercisesByModule("course", "module"))
	}
}

func TestLoadBuildsOrderedImmutableExerciseCatalog(t *testing.T) {
	lessonOneLater := strings.Replace(validExerciseYAML(), "id: exercise\n", "id: later\n", 1)
	lessonOneLater = strings.Replace(lessonOneLater, "order: 0\n", "order: 2\n", 1)
	lessonOneEarlier := strings.Replace(validExerciseYAML(), "id: exercise\n", "id: earlier\n", 1)
	lessonTwo := strings.Replace(validExerciseYAML(), "id: exercise\n", "id: second-lesson\n", 1)
	lessonTwo = strings.Replace(lessonTwo, "lessonId: lesson-one\n", "lessonId: lesson-two\n", 1)

	catalog, err := Load(exerciseCurriculumFS(map[string]string{
		"00-later-filename.yaml": lessonTwo,
		"50-later-order.yaml":    lessonOneLater,
		"99-earlier-order.yaml":  lessonOneEarlier,
	}))
	if err != nil {
		t.Fatalf("load exercise curriculum: %v", err)
	}

	wantIDs := []string{"earlier", "later", "second-lesson"}
	exercises := catalog.ExercisesByModule("course", "module")
	if len(exercises) != len(wantIDs) {
		t.Fatalf("exercise count = %d, want %d: %#v", len(exercises), len(wantIDs), exercises)
	}
	for index, wantID := range wantIDs {
		if exercises[index].ID != wantID {
			t.Fatalf("exercise %d = %q, want %q", index, exercises[index].ID, wantID)
		}
	}
	lessonExercises := catalog.ExercisesByLesson("course", "module", "lesson-one")
	if len(lessonExercises) != 2 || lessonExercises[0].ID != "earlier" || lessonExercises[1].ID != "later" {
		t.Fatalf("unexpected lesson exercises: %#v", lessonExercises)
	}
	exercise, ok := catalog.ExerciseByCourse("course", "module", "earlier")
	if !ok || exercise.CourseID != "course" || exercise.ModuleID != "module" || len(exercise.Tests) != 2 || exercise.Tests[1].Visibility != ExerciseTestHidden {
		t.Fatalf("unexpected exercise lookup: %#v, %v", exercise, ok)
	}
	if _, ok := catalog.ExerciseByCourse("other", "module", "earlier"); ok {
		t.Fatal("exercise lookup ignored course ownership")
	}
	if _, ok := catalog.ExerciseByCourse("course", "other", "earlier"); ok {
		t.Fatal("exercise lookup ignored module ownership")
	}

	exercises[0].ObjectiveIDs[0] = "mutated"
	exercises[0].Tests[0].Code = "mutated"
	again, _ := catalog.ExerciseByCourse("course", "module", "earlier")
	if again.ObjectiveIDs[0] == "mutated" || again.Tests[0].Code == "mutated" {
		t.Fatal("catalog exposed mutable exercise state")
	}
}

func TestLoadRejectsUnknownExerciseFields(t *testing.T) {
	yaml := strings.Replace(validExerciseYAML(), "title: Exercise\n", "title: Exercise\nunknown: true\n", 1)
	expectLoadError(t, exerciseCurriculumFS(map[string]string{"exercise.yaml": yaml}), "not found in type")
}

func TestLoadValidatesExerciseStructure(t *testing.T) {
	valid := validExerciseYAML()
	tests := map[string]struct {
		files    map[string]string
		contains string
	}{
		"invalid id":             {map[string]string{"one.yaml": strings.Replace(valid, "id: exercise\n", "id: Invalid Exercise\n", 1)}, "invalid exercise id"},
		"empty title":            {map[string]string{"one.yaml": strings.Replace(valid, "title: Exercise\n", "title: \"\"\n", 1)}, "empty title"},
		"empty lesson":           {map[string]string{"one.yaml": strings.Replace(valid, "lessonId: lesson-one\n", "lessonId: \"\"\n", 1)}, "empty lessonId"},
		"unknown lesson":         {map[string]string{"one.yaml": strings.Replace(valid, "lessonId: lesson-one\n", "lessonId: missing\n", 1)}, "unknown lesson id"},
		"missing order":          {map[string]string{"one.yaml": strings.Replace(valid, "order: 0\n", "", 1)}, "missing required order"},
		"negative order":         {map[string]string{"one.yaml": strings.Replace(valid, "order: 0\n", "order: -1\n", 1)}, "order must be non-negative"},
		"duplicate id":           {map[string]string{"one.yaml": valid, "two.yaml": valid}, "duplicate exercise id"},
		"duplicate lesson order": {map[string]string{"one.yaml": valid, "two.yaml": strings.Replace(valid, "id: exercise\n", "id: exercise-two\n", 1)}, "duplicate exercise order"},
		"empty objectives":       {map[string]string{"one.yaml": strings.Replace(valid, "objectiveIds:\n  - objective\n", "objectiveIds: []\n", 1)}, "has no objectiveIds"},
		"unknown objective":      {map[string]string{"one.yaml": strings.Replace(valid, "  - objective\n", "  - missing\n", 1)}, "unknown objective id"},
		"empty prompt":           {map[string]string{"one.yaml": strings.Replace(valid, "prompt: Implement it.\n", "prompt: \"\"\n", 1)}, "empty prompt"},
		"empty starter code":     {map[string]string{"one.yaml": strings.Replace(valid, "starterCode: pass\n", "starterCode: \"\"\n", 1)}, "empty starterCode"},
		"empty tests":            {map[string]string{"one.yaml": strings.Split(valid, "tests:")[0] + "tests: []\n"}, "has no tests"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			expectLoadError(t, exerciseCurriculumFS(test.files), test.contains)
		})
	}
}

func TestLoadValidatesExerciseTests(t *testing.T) {
	valid := validExerciseYAML()
	tests := map[string]struct {
		yaml     string
		contains string
	}{
		"invalid id":         {strings.Replace(valid, "  - id: visible-case\n", "  - id: Invalid Case\n", 1), "invalid test id"},
		"duplicate id":       {strings.Replace(valid, "  - id: hidden-case\n", "  - id: visible-case\n", 1), "duplicate test id"},
		"empty title":        {strings.Replace(valid, "    title: Visible case\n", "    title: \"\"\n", 1), "has empty title"},
		"invalid visibility": {strings.Replace(valid, "    visibility: visible\n", "    visibility: public\n", 1), "invalid visibility"},
		"missing visibility": {strings.Replace(valid, "    visibility: visible\n", "", 1), "invalid visibility"},
		"empty code":         {strings.Replace(valid, "    code: assert solution()\n", "    code: \"\"\n", 1), "has empty code"},
		"unknown field":      {strings.Replace(valid, "    title: Visible case\n", "    title: Visible case\n    timeout: 1\n", 1), "not found in type"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			expectLoadError(t, exerciseCurriculumFS(map[string]string{"exercise.yaml": test.yaml}), test.contains)
		})
	}
}

func exerciseCurriculumFS(exercises map[string]string) fstest.MapFS {
	fsys := fstest.MapFS{
		"sources.yaml":                              &fstest.MapFile{Data: []byte("sources: {}\n")},
		"courses/course/course.yaml":                &fstest.MapFile{Data: []byte("id: course\ntitle: Course\ndescription: Description.\norder: 0\n")},
		"courses/course/modules/module/module.yaml": &fstest.MapFile{Data: []byte("id: module\ntitle: Module\norder: 0\nobjectives:\n  - id: objective\n    title: Objective\n    description: Description.\n    prerequisites: []\nvideos: []\nlessons:\n  - lesson-one\n  - lesson-two\n")},
		"courses/course/modules/module/one.mdx":     &fstest.MapFile{Data: []byte("---\nid: lesson-one\ntitle: Lesson one\nobjectiveIds:\n  - objective\nsourceIds: []\n---\n# One\n")},
		"courses/course/modules/module/two.mdx":     &fstest.MapFile{Data: []byte("---\nid: lesson-two\ntitle: Lesson two\nobjectiveIds:\n  - objective\nsourceIds: []\n---\n# Two\n")},
	}
	for name, contents := range exercises {
		fsys["courses/course/modules/module/exercises/"+name] = &fstest.MapFile{Data: []byte(contents)}
	}
	return fsys
}

func validExerciseYAML() string {
	return "id: exercise\n" +
		"title: Exercise\n" +
		"lessonId: lesson-one\n" +
		"order: 0\n" +
		"objectiveIds:\n" +
		"  - objective\n" +
		"prompt: Implement it.\n" +
		"starterCode: pass\n" +
		"tests:\n" +
		"  - id: visible-case\n" +
		"    title: Visible case\n" +
		"    visibility: visible\n" +
		"    code: assert solution()\n" +
		"  - id: hidden-case\n" +
		"    title: Hidden case\n" +
		"    visibility: hidden\n" +
		"    code: assert secret_solution()\n"
}
