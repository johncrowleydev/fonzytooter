package worksheetpdf

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
)

//go:embed template.tex
var latexTemplate string

var ErrToolUnavailable = errors.New("worksheet PDF tool unavailable")

type ToolUnavailableError struct {
	Tool string
	Err  error
}

func (e *ToolUnavailableError) Error() string {
	return fmt.Sprintf("required worksheet PDF tool %q is unavailable: %v", e.Tool, e.Err)
}

func (e *ToolUnavailableError) Unwrap() error {
	return ErrToolUnavailable
}

type commandRunner interface {
	LookPath(file string) (string, error)
	Run(ctx context.Context, directory, command string, arguments ...string) ([]byte, error)
}

type execRunner struct{}

func (execRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (execRunner) Run(ctx context.Context, directory, command string, arguments ...string) ([]byte, error) {
	process := exec.CommandContext(ctx, command, arguments...)
	process.Dir = directory
	return process.CombinedOutput()
}

type Renderer struct {
	runner   commandRunner
	tempRoot string
}

func NewRenderer() *Renderer {
	return &Renderer{runner: execRunner{}}
}

func (renderer *Renderer) Render(ctx context.Context, worksheet curriculum.Worksheet, variant Variant) ([]byte, error) {
	markdown, err := BuildMarkdown(worksheet, variant)
	if err != nil {
		return nil, err
	}

	pandoc, err := renderer.runner.LookPath("pandoc")
	if err != nil {
		return nil, &ToolUnavailableError{Tool: "pandoc", Err: err}
	}
	tectonic, err := renderer.runner.LookPath("tectonic")
	if err != nil {
		return nil, &ToolUnavailableError{Tool: "tectonic", Err: err}
	}

	directory, err := os.MkdirTemp(renderer.tempRoot, "fonzytooter-worksheet-*")
	if err != nil {
		return nil, fmt.Errorf("create worksheet PDF workspace: %w", err)
	}
	defer os.RemoveAll(directory)

	markdownPath := filepath.Join(directory, "worksheet.md")
	templatePath := filepath.Join(directory, "template.tex")
	texPath := filepath.Join(directory, "worksheet.tex")
	pdfPath := filepath.Join(directory, "worksheet.pdf")
	if err := os.WriteFile(markdownPath, []byte(markdown), 0o600); err != nil {
		return nil, fmt.Errorf("write worksheet Markdown: %w", err)
	}
	if err := os.WriteFile(templatePath, []byte(latexTemplate), 0o600); err != nil {
		return nil, fmt.Errorf("write worksheet LaTeX template: %w", err)
	}

	output, err := renderer.runner.Run(
		ctx,
		directory,
		pandoc,
		"--from=markdown+tex_math_dollars+fenced_code_blocks+raw_tex",
		"--to=latex",
		"--standalone",
		"--listings",
		"--template="+templatePath,
		"--output="+texPath,
		markdownPath,
	)
	if err != nil {
		return nil, commandError(ctx, "pandoc", err, output)
	}

	output, err = renderer.runner.Run(ctx, directory, tectonic, "--outdir", directory, texPath)
	if err != nil {
		return nil, commandError(ctx, "tectonic", err, output)
	}

	pdf, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("read rendered worksheet PDF: %w", err)
	}
	if !strings.HasPrefix(string(pdf), "%PDF-") {
		return nil, errors.New("worksheet renderer produced an invalid PDF")
	}
	return pdf, nil
}

func commandError(ctx context.Context, tool string, err error, output []byte) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return fmt.Errorf("render worksheet with %s: %w", tool, err)
	}
	return fmt.Errorf("render worksheet with %s: %w: %s", tool, err, detail)
}
