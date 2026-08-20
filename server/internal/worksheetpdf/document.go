package worksheetpdf

import (
	"errors"
	"fmt"
	"strings"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
)

type Variant string

const (
	Student   Variant = "student"
	Solutions Variant = "solutions"
)

var ErrInvalidVariant = errors.New("invalid worksheet document variant")

func ParseVariant(documentID string) (Variant, error) {
	variant := Variant(documentID)
	switch variant {
	case Student, Solutions:
		return variant, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidVariant, documentID)
	}
}

func Filename(worksheetID string, variant Variant) (string, error) {
	switch variant {
	case Student:
		return worksheetID + ".pdf", nil
	case Solutions:
		return worksheetID + "-solutions.pdf", nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidVariant, variant)
	}
}

// BuildMarkdown converts validated worksheet data into Pandoc input. Authored
// Markdown and dollar-delimited LaTeX are preserved for Pandoc to interpret.
func BuildMarkdown(worksheet curriculum.Worksheet, variant Variant) (string, error) {
	if variant != Student && variant != Solutions {
		return "", fmt.Errorf("%w: %q", ErrInvalidVariant, variant)
	}

	var document strings.Builder
	document.WriteString("# ")
	document.WriteString(worksheet.Title)
	document.WriteString("\n\n")
	if variant == Student {
		document.WriteString("\\textbf{Name:} \\rule{2.5in}{0.4pt} \\hfill \\textbf{Date:} \\rule{1.75in}{0.4pt}\n\n")
	}
	if worksheet.Instructions != "" {
		document.WriteString("## Instructions\n\n")
		document.WriteString(worksheet.Instructions)
		document.WriteString("\n\n")
	}

	for index, problem := range worksheet.Problems {
		writingLines := problem.ResponseLines
		if writingLines < 1 {
			writingLines = 1
		}
		document.WriteString(fmt.Sprintf("\\Needspace{%d\\baselineskip}\n\n", min(writingLines+5, 16)))
		document.WriteString(fmt.Sprintf("## %d.\n\n", index+1))
		document.WriteString(problem.Prompt)
		document.WriteString("\n\n")

		if variant == Solutions {
			document.WriteString("**Solution**\n\n")
			document.WriteString(problem.ExpectedAnswer)
			document.WriteString("\n\n")
			continue
		}

		document.WriteString(fmt.Sprintf("\\vspace*{%d\\baselineskip}\n\n", writingLines))
	}

	return document.String(), nil
}
