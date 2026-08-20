package worksheetpdf

import (
	"strings"
	"testing"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
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
