package worksheetpdf

import (
	"errors"
	"strings"
	"testing"

	"github.com/johncrowleydev/helix-academy/server/internal/curriculum"
)

func TestBuildMarkdownSeparatesStudentAndSolutionContent(t *testing.T) {
	worksheet := testWorksheet()

	student, err := BuildMarkdown(worksheet, Student)
	if err != nil {
		t.Fatalf("build student Markdown: %v", err)
	}
	for _, forbidden := range []string{"The answer is 7.", "Give one point", "Rubric"} {
		if strings.Contains(student, forbidden) {
			t.Fatalf("student Markdown contains %q:\n%s", forbidden, student)
		}
	}
	for _, expected := range []string{"\\textbf{Name:}", "```python", "print(f(3))", "$f(3)$", "\\vspace*{4\\baselineskip}"} {
		if !strings.Contains(student, expected) {
			t.Fatalf("student Markdown does not contain %q:\n%s", expected, student)
		}
	}

	solutions, err := BuildMarkdown(worksheet, Solutions)
	if err != nil {
		t.Fatalf("build solutions Markdown: %v", err)
	}
	if !strings.Contains(solutions, "The answer is 7.") {
		t.Fatalf("solution Markdown omitted expected answer:\n%s", solutions)
	}
	for _, forbidden := range []string{"Give one point", "\\textbf{Name:}"} {
		if strings.Contains(solutions, forbidden) {
			t.Fatalf("solution Markdown contains %q:\n%s", forbidden, solutions)
		}
	}
}

func TestVariantFilenamesAreDeterministic(t *testing.T) {
	for _, test := range []struct {
		variant Variant
		want    string
	}{
		{variant: Student, want: "function-practice.pdf"},
		{variant: Solutions, want: "function-practice-solutions.pdf"},
	} {
		got, err := Filename("function-practice", test.variant)
		if err != nil {
			t.Fatalf("filename for %s: %v", test.variant, err)
		}
		if got != test.want {
			t.Fatalf("filename for %s = %q, want %q", test.variant, got, test.want)
		}
	}
	for _, test := range []struct {
		variant Variant
		want    string
	}{
		{variant: Student, want: "scientific-python-workbook.pdf"},
		{variant: Solutions, want: "scientific-python-workbook-solutions.pdf"},
	} {
		got, err := WorkbookFilename("scientific-python", test.variant)
		if err != nil {
			t.Fatalf("workbook filename for %s: %v", test.variant, err)
		}
		if got != test.want {
			t.Fatalf("workbook filename for %s = %q, want %q", test.variant, got, test.want)
		}
	}
}

func TestBuildWorkbookMarkdownOrdersWorksheetsAndSeparatesSolutions(t *testing.T) {
	course, module := testWorkbook()
	student, err := BuildWorkbookMarkdown(course, module, Student)
	if err != nil {
		t.Fatalf("build student workbook: %v", err)
	}
	for _, expected := range []string{
		`title: "Scientific Python Worksheet Workbook"`,
		`subtitle: "AI & Machine Learning"`,
		"*Lesson: First lesson*",
		"\\textbf{Name:}",
		"\\clearpage",
		"\\vspace*{4\\baselineskip}",
	} {
		if !strings.Contains(student, expected) {
			t.Fatalf("student workbook omitted %q:\n%s", expected, student)
		}
	}
	if first, second, third := strings.Index(student, "# First worksheet"), strings.Index(student, "# Second worksheet"), strings.Index(student, "# Later lesson worksheet"); first < 0 || second <= first || third <= second {
		t.Fatalf("unexpected workbook order:\n%s", student)
	}
	for _, forbidden := range []string{"First answer", "Second answer", "Later answer", "Private rubric"} {
		if strings.Contains(student, forbidden) {
			t.Fatalf("student workbook contains %q:\n%s", forbidden, student)
		}
	}

	solutions, err := BuildWorkbookMarkdown(course, module, Solutions)
	if err != nil {
		t.Fatalf("build solutions workbook: %v", err)
	}
	for _, expected := range []string{"First answer", "Second answer", "Later answer", "**Solution**"} {
		if !strings.Contains(solutions, expected) {
			t.Fatalf("solutions workbook omitted %q:\n%s", expected, solutions)
		}
	}
	for _, forbidden := range []string{"Private rubric", "\\textbf{Name:}"} {
		if strings.Contains(solutions, forbidden) {
			t.Fatalf("solutions workbook contains %q:\n%s", forbidden, solutions)
		}
	}
}

func TestBuildWorkbookMarkdownRejectsEmptyModules(t *testing.T) {
	_, err := BuildWorkbookMarkdown(curriculum.Course{}, curriculum.Module{}, Student)
	if !errors.Is(err, ErrNoWorksheets) {
		t.Fatalf("error = %v, want ErrNoWorksheets", err)
	}
}

func testWorksheet() curriculum.Worksheet {
	return curriculum.Worksheet{
		CourseID:     "ai-ml",
		ModuleID:     "scientific-python",
		ID:           "function-practice",
		Title:        "Function Practice",
		LessonID:     "functions",
		Instructions: "Show your reasoning.",
		Problems: []curriculum.WorksheetProblem{
			{
				ID:             "evaluate",
				Prompt:         "Evaluate $f(3)$.\n\n```python\nprint(f(3))\n```",
				ExpectedAnswer: "The answer is 7.",
				RequiresWork:   true,
				ResponseLines:  4,
				Rubric:         []string{"Give one point for substitution."},
			},
		},
	}
}

func testWorkbook() (curriculum.Course, curriculum.Module) {
	worksheet := func(id, title, lessonID string, order int, answer string) curriculum.Worksheet {
		result := testWorksheet()
		result.ID = id
		result.Title = title
		result.LessonID = lessonID
		result.Order = order
		result.Problems[0].ExpectedAnswer = answer
		result.Problems[0].Rubric = []string{"Private rubric for " + id}
		return result
	}
	module := curriculum.Module{
		CourseID: "ai-ml",
		ID:       "scientific-python",
		Title:    "Scientific Python",
		Lessons: []curriculum.Lesson{
			{ID: "first", Title: "First lesson"},
			{ID: "later", Title: "Later lesson"},
		},
		Worksheets: []curriculum.Worksheet{
			worksheet("later", "Later lesson worksheet", "later", 0, "Later answer"),
			worksheet("second", "Second worksheet", "first", 2, "Second answer"),
			worksheet("first", "First worksheet", "first", 1, "First answer"),
		},
	}
	return curriculum.Course{ID: "ai-ml", Title: "AI & Machine Learning", Modules: []curriculum.Module{module}}, module
}
