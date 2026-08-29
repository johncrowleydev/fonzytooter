package worksheetpdf

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
)

type Variant string

const (
	Student   Variant = "student"
	Solutions Variant = "solutions"
)

var (
	ErrInvalidVariant = errors.New("invalid worksheet document variant")
	ErrNoWorksheets   = errors.New("module has no worksheets")
)

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

func WorkbookFilename(moduleID string, variant Variant) (string, error) {
	switch variant {
	case Student:
		return moduleID + "-workbook.pdf", nil
	case Solutions:
		return moduleID + "-workbook-solutions.pdf", nil
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
	document.WriteString(" {-}\n\n")
	writeIdentityFields(&document, variant)
	writeInstructions(&document, worksheet.Instructions)
	writeProblems(&document, worksheet, variant)
	return document.String(), nil
}

// BuildWorkbookMarkdown aggregates the module's existing worksheets into one
// Pandoc input document. It sorts a defensive copy so workbook order never
// depends on filesystem discovery or caller slice order.
func BuildWorkbookMarkdown(course curriculum.Course, module curriculum.Module, variant Variant) (string, error) {
	if variant != Student && variant != Solutions {
		return "", fmt.Errorf("%w: %q", ErrInvalidVariant, variant)
	}
	worksheets := orderedWorksheets(module)
	if len(worksheets) == 0 {
		return "", ErrNoWorksheets
	}

	title := module.Title + " Worksheet Workbook"
	if variant == Solutions {
		title += " Solutions"
	}
	var document strings.Builder
	document.WriteString("---\n")
	document.WriteString("title: ")
	document.WriteString(strconv.Quote(title))
	document.WriteString("\nsubtitle: ")
	document.WriteString(strconv.Quote(course.Title))
	document.WriteString("\n---\n\n")

	lessons := make(map[string]string, len(module.Lessons))
	for _, lesson := range module.Lessons {
		lessons[lesson.ID] = lesson.Title
	}
	for index, worksheet := range worksheets {
		if index > 0 {
			document.WriteString("\\clearpage\n\n")
		}
		document.WriteString("# ")
		document.WriteString(worksheet.Title)
		document.WriteString("\n\n")
		if lessonTitle := lessons[worksheet.LessonID]; lessonTitle != "" {
			document.WriteString("*Lesson: ")
			document.WriteString(lessonTitle)
			document.WriteString("*\n\n")
		}
		writeIdentityFields(&document, variant)
		writeInstructions(&document, worksheet.Instructions)
		writeProblems(&document, worksheet, variant)
	}
	return document.String(), nil
}

func writeIdentityFields(document *strings.Builder, variant Variant) {
	if variant == Student {
		document.WriteString("\\identityline\n\n")
	}
}

// writeInstructions keeps the instructions heading unnumbered so it stays
// out of the workbook table of contents, which lists worksheets only.
func writeInstructions(document *strings.Builder, instructions string) {
	if instructions == "" {
		return
	}
	document.WriteString("## Instructions {-}\n\n")
	document.WriteString(instructions)
	document.WriteString("\n\n")
}

func writeProblems(document *strings.Builder, worksheet curriculum.Worksheet, variant Variant) {
	for index, problem := range worksheet.Problems {
		writingLines := problem.ResponseLines
		if writingLines < 1 {
			writingLines = 1
		}
		document.WriteString(fmt.Sprintf("\\Needspace{%d\\baselineskip}\n\n", min(writingLines*3/2+5, 20)))
		document.WriteString(fmt.Sprintf("## %d. {-}\n\n", index+1))
		document.WriteString(problem.Prompt)
		document.WriteString("\n\n")

		if variant == Solutions {
			document.WriteString("**Solution**\n\n")
			document.WriteString(problem.ExpectedAnswer)
			document.WriteString("\n\n")
			continue
		}

		document.WriteString(fmt.Sprintf("\\answerlines{%d}\n\n", writingLines))
	}
}

func orderedWorksheets(module curriculum.Module) []curriculum.Worksheet {
	worksheets := append([]curriculum.Worksheet(nil), module.Worksheets...)
	lessonOrder := make(map[string]int, len(module.Lessons))
	for index, lesson := range module.Lessons {
		lessonOrder[lesson.ID] = index
	}
	sort.SliceStable(worksheets, func(i, j int) bool {
		leftLesson, leftOK := lessonOrder[worksheets[i].LessonID]
		rightLesson, rightOK := lessonOrder[worksheets[j].LessonID]
		if leftOK != rightOK {
			return leftOK
		}
		if leftLesson != rightLesson {
			return leftLesson < rightLesson
		}
		if worksheets[i].Order != worksheets[j].Order {
			return worksheets[i].Order < worksheets[j].Order
		}
		return worksheets[i].ID < worksheets[j].ID
	})
	return worksheets
}
