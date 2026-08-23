package curriculum

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadRealWorksheets(t *testing.T) {
	curriculumPath := filepath.Join("..", "..", "..", "curriculum")
	catalog, err := Load(os.DirFS(curriculumPath))
	if err != nil {
		t.Fatalf("load real curriculum: %v", err)
	}
	if catalog.WorksheetCount() != 6 {
		t.Fatalf("worksheet count = %d, want 6", catalog.WorksheetCount())
	}

	worksheets := catalog.WorksheetsByModule("ai-ml", "scientific-python")
	wantIDs := []string{
		"python-execution-model-practice",
		"function-mappings-and-properties",
		"inverses-and-composition",
		"functions-in-machine-learning",
		"array-shapes-and-indexing",
		"broadcasting-and-shape-reasoning",
	}
	if len(worksheets) != len(wantIDs) {
		t.Fatalf("module worksheets = %#v", worksheets)
	}
	for index, wantID := range wantIDs {
		if worksheets[index].ID != wantID {
			t.Fatalf("worksheet %d = %q, want %q", index, worksheets[index].ID, wantID)
		}
	}
	lessonWorksheets := catalog.WorksheetsByLesson("ai-ml", "scientific-python", "02-functions-code-and-mathematics")
	if len(lessonWorksheets) != 3 || lessonWorksheets[0].Order != 1 || lessonWorksheets[2].Order != 3 {
		t.Fatalf("unexpected lesson worksheet order: %#v", lessonWorksheets)
	}
	worksheet, ok := catalog.WorksheetByCourse("ai-ml", "scientific-python", "inverses-and-composition")
	if !ok || worksheet.CourseID != "ai-ml" || worksheet.ModuleID != "scientific-python" || len(worksheet.Problems) == 0 {
		t.Fatalf("unexpected worksheet lookup: %#v, %v", worksheet, ok)
	}
	if !strings.Contains(worksheet.Problems[0].ExpectedAnswer, `f^{-1}`) {
		t.Fatalf("expected authored LaTeX to be preserved, got %q", worksheet.Problems[0].ExpectedAnswer)
	}

	worksheets[0].ObjectiveIDs[0] = "mutated"
	worksheets[0].Problems[0].Prompt = "mutated"
	worksheets[0].Problems[0].Rubric[0] = "mutated"
	again, _ := catalog.WorksheetByCourse("ai-ml", "scientific-python", wantIDs[0])
	if again.ObjectiveIDs[0] == "mutated" || again.Problems[0].Prompt == "mutated" || again.Problems[0].Rubric[0] == "mutated" {
		t.Fatal("catalog exposed mutable worksheet state")
	}
	if _, ok := catalog.WorksheetByCourse("other", "scientific-python", wantIDs[0]); ok {
		t.Fatal("worksheet lookup ignored course ownership")
	}
	if _, ok := catalog.WorksheetByCourse("ai-ml", "other", wantIDs[0]); ok {
		t.Fatal("worksheet lookup ignored module ownership")
	}
}

func TestLoadAllowsModulesWithoutWorksheets(t *testing.T) {
	catalog, err := Load(worksheetCurriculumFS(nil))
	if err != nil {
		t.Fatalf("load curriculum without worksheets: %v", err)
	}
	if catalog.WorksheetCount() != 0 || len(catalog.WorksheetsByModule("course", "module")) != 0 || len(catalog.WorksheetsByLesson("course", "module", "lesson")) != 0 {
		t.Fatalf("expected no worksheets, got %#v", catalog.WorksheetsByModule("course", "module"))
	}
}

func TestWorksheetOrderingDoesNotDependOnFilenames(t *testing.T) {
	later := strings.Replace(validWorksheetYAML(), "id: worksheet\n", "id: later\n", 1)
	later = strings.Replace(later, "order: 0\n", "order: 1\n", 1)
	earlier := strings.Replace(validWorksheetYAML(), "id: worksheet\n", "id: earlier\n", 1)
	catalog, err := Load(worksheetCurriculumFS(map[string]string{
		"00-later.yaml":   later,
		"99-earlier.yaml": earlier,
	}))
	if err != nil {
		t.Fatalf("load ordered worksheets: %v", err)
	}
	for name, worksheets := range map[string][]Worksheet{
		"module": catalog.WorksheetsByModule("course", "module"),
		"lesson": catalog.WorksheetsByLesson("course", "module", "lesson"),
	} {
		if len(worksheets) != 2 || worksheets[0].ID != "earlier" || worksheets[1].ID != "later" {
			t.Fatalf("%s worksheet order = %#v", name, worksheets)
		}
	}
	module, _ := catalog.ModuleByCourse("course", "module")
	if module.Worksheets[0].ID != "earlier" {
		t.Fatalf("module worksheet order = %#v", module.Worksheets)
	}
}

func TestLoadRejectsUnknownWorksheetFields(t *testing.T) {
	yaml := strings.Replace(validWorksheetYAML(), "title: Worksheet\n", "title: Worksheet\nunknown: true\n", 1)
	expectLoadError(t, worksheetCurriculumFS(map[string]string{"first.yaml": yaml}), "not found in type")
}

func TestLoadValidatesWorksheetStructure(t *testing.T) {
	valid := validWorksheetYAML()
	tests := map[string]struct {
		files    map[string]string
		contains string
	}{
		"invalid worksheet id":   {map[string]string{"first.yaml": strings.Replace(valid, "id: worksheet\n", "id: Invalid Worksheet\n", 1)}, "invalid worksheet id"},
		"empty title":            {map[string]string{"first.yaml": strings.Replace(valid, "title: Worksheet\n", "title: \"\"\n", 1)}, "empty title"},
		"empty lesson id":        {map[string]string{"first.yaml": strings.Replace(valid, "lessonId: lesson\n", "lessonId: \"\"\n", 1)}, "empty lessonId"},
		"unknown lesson id":      {map[string]string{"first.yaml": strings.Replace(valid, "lessonId: lesson\n", "lessonId: missing\n", 1)}, "unknown lesson id"},
		"missing order":          {map[string]string{"first.yaml": strings.Replace(valid, "order: 0\n", "", 1)}, "missing required order"},
		"negative order":         {map[string]string{"first.yaml": strings.Replace(valid, "order: 0\n", "order: -1\n", 1)}, "order must be non-negative"},
		"duplicate id":           {map[string]string{"first.yaml": valid, "second.yaml": valid}, "duplicate worksheet id"},
		"duplicate lesson order": {map[string]string{"first.yaml": valid, "second.yaml": strings.Replace(valid, "id: worksheet\n", "id: worksheet-two\n", 1)}, "duplicate worksheet order"},
		"empty objectives":       {map[string]string{"first.yaml": strings.Replace(valid, "objectiveIds:\n  - objective\n", "objectiveIds: []\n", 1)}, "has no objectiveIds"},
		"unknown objective":      {map[string]string{"first.yaml": strings.Replace(valid, "  - objective\n", "  - missing\n", 1)}, "unknown objective id"},
		"empty instructions":     {map[string]string{"first.yaml": strings.Replace(valid, "instructions: Instructions.\n", "instructions: \"\"\n", 1)}, "empty instructions"},
		"empty problems":         {map[string]string{"first.yaml": strings.Split(valid, "problems:")[0] + "problems: []\n"}, "has no problems"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			expectLoadError(t, worksheetCurriculumFS(test.files), test.contains)
		})
	}
}

func TestLoadValidatesWorksheetProblems(t *testing.T) {
	valid := validWorksheetYAML()
	tests := map[string]struct {
		yaml     string
		contains string
	}{
		"invalid problem id":     {strings.Replace(valid, "  - id: problem\n", "  - id: Invalid Problem\n", 1), "invalid problem id"},
		"duplicate problem id":   {valid + validProblemYAML(), "duplicate problem id"},
		"empty prompt":           {strings.Replace(valid, "    prompt: Prompt.\n", "    prompt: \"\"\n", 1), "empty prompt"},
		"empty objectives":       {strings.Replace(valid, "    objectiveIds:\n      - objective\n", "    objectiveIds: []\n", 1), "has no objectiveIds"},
		"unknown objective":      {strings.Replace(valid, "      - objective\n", "      - missing\n", 1), "unknown objective id"},
		"empty expected answer":  {strings.Replace(valid, "    expectedAnswer: Answer.\n", "    expectedAnswer: \"\"\n", 1), "empty expectedAnswer"},
		"missing requires work":  {strings.Replace(valid, "    requiresWork: false\n", "", 1), "missing required requiresWork"},
		"missing response lines": {strings.Replace(valid, "    responseLines: 1\n", "", 1), "missing required responseLines"},
		"zero response lines":    {strings.Replace(valid, "    responseLines: 1\n", "    responseLines: 0\n", 1), "greater than zero"},
		"empty rubric":           {strings.Replace(valid, "    rubric:\n      - Criterion.\n", "    rubric: []\n", 1), "empty rubric"},
		"empty rubric item":      {strings.Replace(valid, "      - Criterion.\n", "      - \"\"\n", 1), "empty rubric item"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			expectLoadError(t, worksheetCurriculumFS(map[string]string{"first.yaml": test.yaml}), test.contains)
		})
	}
}

func worksheetCurriculumFS(worksheets map[string]string) fstest.MapFS {
	fys := fstest.MapFS{
		"sources.yaml":                              &fstest.MapFile{Data: []byte("sources: {}\n")},
		"courses/course/course.yaml":                &fstest.MapFile{Data: []byte("id: course\ntitle: Course\ndescription: Description.\norder: 0\n")},
		"courses/course/modules/module/module.yaml": &fstest.MapFile{Data: []byte("id: module\ntitle: Module\norder: 0\nobjectives:\n  - id: objective\n    title: Objective\n    description: Description.\n    prerequisites: []\nvideos: []\nlessons:\n  - lesson\n")},
		"courses/course/modules/module/lesson.mdx":  &fstest.MapFile{Data: []byte("---\nid: lesson\ntitle: Lesson\nobjectiveIds:\n  - objective\nsourceIds: []\n---\n# Lesson\n")},
	}
	for name, contents := range worksheets {
		fys["courses/course/modules/module/worksheets/"+name] = &fstest.MapFile{Data: []byte(contents)}
	}
	return fys
}

func validWorksheetYAML() string {
	return "id: worksheet\n" +
		"title: Worksheet\n" +
		"lessonId: lesson\n" +
		"order: 0\n" +
		"objectiveIds:\n" +
		"  - objective\n" +
		"instructions: Instructions.\n" +
		"problems:\n" + validProblemYAML()
}

func validProblemYAML() string {
	return "  - id: problem\n" +
		"    prompt: Prompt.\n" +
		"    objectiveIds:\n" +
		"      - objective\n" +
		"    expectedAnswer: Answer.\n" +
		"    requiresWork: false\n" +
		"    responseLines: 1\n" +
		"    rubric:\n" +
		"      - Criterion.\n"
}
