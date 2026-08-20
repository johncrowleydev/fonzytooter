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
	args    [][]string
}

func (runner *fakeRunner) LookPath(file string) (string, error) {
	if runner.missing == file {
		return "", os.ErrNotExist
	}
	return file, nil
}

func (runner *fakeRunner) Run(_ context.Context, directory, command string, arguments ...string) ([]byte, error) {
	runner.runs = append(runner.runs, command)
	runner.args = append(runner.args, append([]string(nil), arguments...))
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

func TestRendererReturnsWorkbookPDFWithTableOfContents(t *testing.T) {
	runner := &fakeRunner{}
	renderer := &Renderer{runner: runner, tempRoot: t.TempDir()}
	course, module := testWorkbook()

	pdf, err := renderer.RenderWorkbook(context.Background(), course, module, Solutions)
	if err != nil {
		t.Fatalf("render solutions workbook: %v", err)
	}
	if !strings.HasPrefix(string(pdf), "%PDF-") {
		t.Fatalf("rendered bytes do not start with PDF header: %q", pdf)
	}
	if len(runner.args) == 0 || !containsArgument(runner.args[0], "--table-of-contents") {
		t.Fatalf("Pandoc arguments omitted table of contents: %#v", runner.args)
	}
}

func containsArgument(arguments []string, want string) bool {
	for _, argument := range arguments {
		if argument == want {
			return true
		}
	}
	return false
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

func TestRendererTectonicArguments(t *testing.T) {
	for _, test := range []struct {
		name       string
		onlyCached bool
		want       []string
	}{
		{
			name: "default allows resource downloads",
			want: []string{"--outdir", "output", "worksheet.tex"},
		},
		{
			name:       "cache only prohibits resource downloads",
			onlyCached: true,
			want:       []string{"--outdir", "output", "--only-cached", "worksheet.tex"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			renderer := &Renderer{tectonicOnlyCached: test.onlyCached}
			got := renderer.tectonicArguments("output", "worksheet.tex")
			if strings.Join(got, "\x00") != strings.Join(test.want, "\x00") {
				t.Fatalf("Tectonic arguments = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestNewRendererReadsTectonicOnlyCachedEnvironment(t *testing.T) {
	for _, test := range []struct {
		name  string
		value string
		want  bool
	}{
		{name: "unset", value: "", want: false},
		{name: "enabled", value: "1", want: true},
		{name: "other values remain disabled", value: "true", want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(tectonicOnlyCachedEnvironment, test.value)
			t.Setenv(tectonicCacheDirectoryEnvironment, "test-cache")
			renderer := NewRenderer()
			if got := renderer.tectonicOnlyCached; got != test.want {
				t.Fatalf("cache-only mode = %t, want %t", got, test.want)
			}
			if renderer.tectonicCacheDir != "test-cache" {
				t.Fatalf("Tectonic cache directory = %q, want test-cache", renderer.tectonicCacheDir)
			}
		})
	}
}

func TestRendererRejectsUnprimedCacheBeforeRunningCommands(t *testing.T) {
	runner := &fakeRunner{}
	renderer := &Renderer{
		runner:             runner,
		tectonicOnlyCached: true,
		tectonicCacheDir:   t.TempDir(),
	}

	_, err := renderer.Render(context.Background(), testWorksheet(), Student)
	if err == nil || !strings.Contains(err.Error(), "requires a primed resource cache") {
		t.Fatalf("error = %v, want unprimed cache guidance", err)
	}
	if len(runner.runs) != 0 {
		t.Fatalf("commands ran with an unprimed cache: %v", runner.runs)
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
