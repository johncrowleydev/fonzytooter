package worksheetpdf

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
)

func TestRealWorksheetRendering(t *testing.T) {
	requirePDFTools(t)
	catalog, err := curriculum.Load(os.DirFS("../../../curriculum"))
	if err != nil {
		t.Fatalf("load curriculum: %v", err)
	}

	renderer := NewRenderer()
	rendered := 0
	for _, course := range catalog.Courses() {
		for _, module := range course.Modules {
			for _, worksheet := range module.Worksheets {
				pdf, renderErr := renderer.Render(context.Background(), worksheet, Student)
				if renderErr != nil {
					t.Fatalf("render student worksheet %s: %v", worksheet.ID, renderErr)
				}
				if !strings.HasPrefix(string(pdf), "%PDF-") {
					t.Fatalf("worksheet %s did not produce PDF bytes", worksheet.ID)
				}
				rendered++
			}
		}
	}
	if rendered == 0 {
		t.Fatal("expected at least one authored worksheet")
	}

	worksheet, ok := catalog.WorksheetByCourse("ai-ml", "scientific-python", "python-execution-model-practice")
	if !ok {
		t.Fatal("expected authored python-execution-model-practice worksheet")
	}
	solutions, err := renderer.Render(context.Background(), worksheet, Solutions)
	if err != nil {
		t.Fatalf("render real solutions PDF: %v", err)
	}
	if !strings.HasPrefix(string(solutions), "%PDF-") {
		t.Fatal("solutions did not produce PDF bytes")
	}
}

func TestPandocConvertsFencedCodeToReadableLatex(t *testing.T) {
	requirePDFTools(t)
	directory := t.TempDir()
	markdown, err := BuildMarkdown(testWorksheet(), Student)
	if err != nil {
		t.Fatalf("build Markdown: %v", err)
	}
	inputPath := filepath.Join(directory, "worksheet.md")
	outputPath := filepath.Join(directory, "worksheet.tex")
	if err := os.WriteFile(inputPath, []byte(markdown), 0o600); err != nil {
		t.Fatalf("write Markdown: %v", err)
	}
	command := exec.Command(
		"pandoc",
		"--from=markdown+tex_math_dollars+fenced_code_blocks+raw_tex",
		"--to=latex",
		"--no-highlight",
		"--output="+outputPath,
		inputPath,
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("run Pandoc: %v: %s", err, output)
	}
	latex, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read LaTeX: %v", err)
	}
	for _, expected := range []string{"\\begin{verbatim}", "print(f(3))", "f(3)"} {
		if !strings.Contains(string(latex), expected) {
			t.Fatalf("Pandoc output omitted %q:\n%s", expected, latex)
		}
	}
}

func requirePDFTools(t *testing.T) {
	t.Helper()
	for _, tool := range []string{"pandoc", "tectonic"} {
		if _, err := exec.LookPath(tool); err != nil {
			if os.Getenv("FONZYTOOTER_REQUIRE_PDF_TOOLS") == "1" {
				t.Fatalf("required PDF tool %s is unavailable: %v", tool, err)
			}
			t.Skipf("PDF tool %s is unavailable", tool)
		}
	}
}
