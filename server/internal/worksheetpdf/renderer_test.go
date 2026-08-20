package worksheetpdf

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeRunner struct {
	missing string
	runs    []string
}

func (runner *fakeRunner) LookPath(file string) (string, error) {
	if runner.missing == file {
		return "", os.ErrNotExist
	}
	return file, nil
}

func (runner *fakeRunner) Run(_ context.Context, directory, command string, arguments ...string) ([]byte, error) {
	runner.runs = append(runner.runs, command)
	switch filepath.Base(command) {
	case "pandoc", "pandoc.exe":
		var outputPath string
		for _, argument := range arguments {
			if strings.HasPrefix(argument, "--output=") {
				outputPath = strings.TrimPrefix(argument, "--output=")
			}
		}
		if outputPath == "" {
			return nil, errors.New("missing Pandoc output argument")
		}
		return nil, os.WriteFile(outputPath, []byte("\\documentclass{article}\\begin{document}ok\\end{document}"), 0o600)
	case "tectonic", "tectonic.exe":
		return nil, os.WriteFile(filepath.Join(directory, "worksheet.pdf"), []byte("%PDF-1.7\nfake"), 0o600)
	default:
		return nil, errors.New("unexpected command")
	}
}

func TestRendererUsesTemporaryWorkspaceAndReturnsPDF(t *testing.T) {
	tempRoot := t.TempDir()
	runner := &fakeRunner{}
	renderer := &Renderer{runner: runner, tempRoot: tempRoot}

	pdf, err := renderer.Render(context.Background(), testWorksheet(), Student)
	if err != nil {
		t.Fatalf("render student PDF: %v", err)
	}
	if !strings.HasPrefix(string(pdf), "%PDF-") {
		t.Fatalf("rendered bytes do not start with PDF header: %q", pdf)
	}
	if got := strings.Join(runner.runs, ","); got != "pandoc,tectonic" {
		t.Fatalf("commands = %q", got)
	}
	entries, err := os.ReadDir(tempRoot)
	if err != nil {
		t.Fatalf("read temporary root: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("renderer left temporary artifacts: %v", entries)
	}
}

func TestRendererReportsMissingTools(t *testing.T) {
	for _, tool := range []string{"pandoc", "tectonic"} {
		t.Run(tool, func(t *testing.T) {
			renderer := &Renderer{runner: &fakeRunner{missing: tool}}
			_, err := renderer.Render(context.Background(), testWorksheet(), Student)
			if !errors.Is(err, ErrToolUnavailable) {
				t.Fatalf("error = %v, want ErrToolUnavailable", err)
			}
			var unavailable *ToolUnavailableError
			if !errors.As(err, &unavailable) || unavailable.Tool != tool {
				t.Fatalf("error = %#v, want missing %s", err, tool)
			}
		})
	}
}
