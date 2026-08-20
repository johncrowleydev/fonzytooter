package curriculumstate

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
)

type queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

type Finding struct {
	Table  string
	Record string
	Reason string
}

type Report struct {
	Findings []Finding
}

func (r Report) Clean() bool { return len(r.Findings) == 0 }

func Audit(ctx context.Context, db *sql.DB, catalog *curriculum.Catalog) (Report, error) {
	checks := []func(context.Context, queryer, *curriculum.Catalog) ([]Finding, error){
		auditActivities,
		auditExerciseAttempts,
		auditExerciseTestResults,
		auditExerciseWorkspaces,
		auditLessonProgress,
		auditReviewCards,
		auditReviewLogs,
	}
	report := Report{}
	for _, check := range checks {
		findings, err := check(ctx, db, catalog)
		if err != nil {
			return Report{}, err
		}
		report.Findings = append(report.Findings, findings...)
	}
	sort.Slice(report.Findings, func(i, j int) bool {
		if report.Findings[i].Table != report.Findings[j].Table {
			return report.Findings[i].Table < report.Findings[j].Table
		}
		if report.Findings[i].Record != report.Findings[j].Record {
			return report.Findings[i].Record < report.Findings[j].Record
		}
		return report.Findings[i].Reason < report.Findings[j].Reason
	})
	return report, nil
}

func WriteReport(writer io.Writer, report Report) error {
	if report.Clean() {
		_, err := fmt.Fprintln(writer, "curriculum state valid: no orphaned curriculum references")
		return err
	}
	if _, err := fmt.Fprintf(writer, "curriculum state invalid: %d orphaned rows\n", len(report.Findings)); err != nil {
		return err
	}
	for index, finding := range report.Findings {
		if index == 0 || report.Findings[index-1].Table != finding.Table {
			count := 0
			for _, candidate := range report.Findings[index:] {
				if candidate.Table != finding.Table {
					break
				}
				count++
			}
			if _, err := fmt.Fprintf(writer, "\n%s (%d)\n", finding.Table, count); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(writer, "- %s: %s\n", finding.Record, finding.Reason); err != nil {
			return err
		}
	}
	return nil
}

func auditLessonProgress(ctx context.Context, db queryer, catalog *curriculum.Catalog) ([]Finding, error) {
	rows, err := db.QueryContext(ctx, `SELECT course_id, module_id, lesson_id FROM lesson_progress ORDER BY course_id, module_id, lesson_id`)
	if err != nil {
		return nil, fmt.Errorf("audit lesson_progress: %w", err)
	}
	defer rows.Close()
	var findings []Finding
	for rows.Next() {
		var courseID, moduleID, lessonID string
		if err := rows.Scan(&courseID, &moduleID, &lessonID); err != nil {
			return nil, fmt.Errorf("scan lesson_progress: %w", err)
		}
		if _, ok := catalog.LessonByCourse(courseID, moduleID, lessonID); !ok {
			findings = append(findings, Finding{"lesson_progress", qualified(courseID, moduleID, lessonID), "lesson does not exist in the current curriculum"})
		}
	}
	return findings, rows.Err()
}

func auditExerciseWorkspaces(ctx context.Context, db queryer, catalog *curriculum.Catalog) ([]Finding, error) {
	rows, err := db.QueryContext(ctx, `SELECT course_id, module_id, exercise_id FROM exercise_workspaces ORDER BY course_id, module_id, exercise_id`)
	if err != nil {
		return nil, fmt.Errorf("audit exercise_workspaces: %w", err)
	}
	defer rows.Close()
	var findings []Finding
	for rows.Next() {
		var courseID, moduleID, exerciseID string
		if err := rows.Scan(&courseID, &moduleID, &exerciseID); err != nil {
			return nil, fmt.Errorf("scan exercise_workspaces: %w", err)
		}
		if _, ok := catalog.ExerciseByCourse(courseID, moduleID, exerciseID); !ok {
			findings = append(findings, Finding{"exercise_workspaces", qualified(courseID, moduleID, exerciseID), "exercise does not exist in the current curriculum"})
		}
	}
	return findings, rows.Err()
}

func auditExerciseAttempts(ctx context.Context, db queryer, catalog *curriculum.Catalog) ([]Finding, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, course_id, module_id, exercise_id FROM exercise_attempts ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("audit exercise_attempts: %w", err)
	}
	defer rows.Close()
	var findings []Finding
	for rows.Next() {
		var id int64
		var courseID, moduleID, exerciseID string
		if err := rows.Scan(&id, &courseID, &moduleID, &exerciseID); err != nil {
			return nil, fmt.Errorf("scan exercise_attempts: %w", err)
		}
		if _, ok := catalog.ExerciseByCourse(courseID, moduleID, exerciseID); !ok {
			findings = append(findings, Finding{"exercise_attempts", fmt.Sprintf("id=%d %s", id, qualified(courseID, moduleID, exerciseID)), "exercise does not exist in the current curriculum"})
		}
	}
	return findings, rows.Err()
}

func auditExerciseTestResults(ctx context.Context, db queryer, catalog *curriculum.Catalog) ([]Finding, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT r.attempt_id, a.course_id, a.module_id, a.exercise_id, r.test_id
		FROM exercise_test_results r
		LEFT JOIN exercise_attempts a ON a.id = r.attempt_id
		ORDER BY r.attempt_id, r.test_id`)
	if err != nil {
		return nil, fmt.Errorf("audit exercise_test_results: %w", err)
	}
	defer rows.Close()
	var findings []Finding
	for rows.Next() {
		var attemptID int64
		var courseID, moduleID, exerciseID sql.NullString
		var testID string
		if err := rows.Scan(&attemptID, &courseID, &moduleID, &exerciseID, &testID); err != nil {
			return nil, fmt.Errorf("scan exercise_test_results: %w", err)
		}
		if !courseID.Valid || !moduleID.Valid || !exerciseID.Valid {
			findings = append(findings, Finding{"exercise_test_results", fmt.Sprintf("attempt_id=%d test_id=%s", attemptID, testID), "owning exercise attempt does not exist"})
			continue
		}
		exercise, exerciseOK := catalog.ExerciseByCourse(courseID.String, moduleID.String, exerciseID.String)
		testOK := false
		if exerciseOK {
			for _, test := range exercise.Tests {
				if test.ID == testID {
					testOK = true
					break
				}
			}
		}
		if !testOK {
			reason := "exercise test does not exist in the current curriculum"
			if !exerciseOK {
				reason = "owning exercise does not exist in the current curriculum"
			}
			findings = append(findings, Finding{"exercise_test_results", fmt.Sprintf("attempt_id=%d %s", attemptID, qualified(courseID.String, moduleID.String, exerciseID.String, testID)), reason})
		}
	}
	return findings, rows.Err()
}

func auditReviewCards(ctx context.Context, db queryer, catalog *curriculum.Catalog) ([]Finding, error) {
	rows, err := db.QueryContext(ctx, `SELECT course_id, module_id, review_item_id FROM review_cards ORDER BY course_id, module_id, review_item_id`)
	if err != nil {
		return nil, fmt.Errorf("audit review_cards: %w", err)
	}
	defer rows.Close()
	var findings []Finding
	for rows.Next() {
		var courseID, moduleID, reviewItemID string
		if err := rows.Scan(&courseID, &moduleID, &reviewItemID); err != nil {
			return nil, fmt.Errorf("scan review_cards: %w", err)
		}
		if _, ok := catalog.ReviewItemByCourse(courseID, moduleID, reviewItemID); !ok {
			findings = append(findings, Finding{"review_cards", qualified(courseID, moduleID, reviewItemID), "review item does not exist in the current curriculum"})
		}
	}
	return findings, rows.Err()
}

func auditReviewLogs(ctx context.Context, db queryer, catalog *curriculum.Catalog) ([]Finding, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, course_id, module_id, review_item_id FROM review_logs ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("audit review_logs: %w", err)
	}
	defer rows.Close()
	var findings []Finding
	for rows.Next() {
		var id int64
		var courseID, moduleID, reviewItemID string
		if err := rows.Scan(&id, &courseID, &moduleID, &reviewItemID); err != nil {
			return nil, fmt.Errorf("scan review_logs: %w", err)
		}
		if _, ok := catalog.ReviewItemByCourse(courseID, moduleID, reviewItemID); !ok {
			findings = append(findings, Finding{"review_logs", fmt.Sprintf("id=%d %s", id, qualified(courseID, moduleID, reviewItemID)), "review item does not exist in the current curriculum"})
		}
	}
	return findings, rows.Err()
}

func auditActivities(ctx context.Context, db queryer, catalog *curriculum.Catalog) ([]Finding, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, kind, course_id, module_id, lesson_id, exercise_id, review_item_id
		FROM activities ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("audit activities: %w", err)
	}
	defer rows.Close()
	var findings []Finding
	for rows.Next() {
		var id int64
		var kind, courseID string
		var moduleID, lessonID, exerciseID, reviewItemID sql.NullString
		if err := rows.Scan(&id, &kind, &courseID, &moduleID, &lessonID, &exerciseID, &reviewItemID); err != nil {
			return nil, fmt.Errorf("scan activities: %w", err)
		}
		var reasons []string
		if _, ok := catalog.CourseByID(courseID); !ok {
			reasons = append(reasons, "course does not exist")
		}
		if moduleID.Valid {
			if _, ok := catalog.ModuleByCourse(courseID, moduleID.String); !ok {
				reasons = append(reasons, "module does not exist")
			}
		}
		if lessonID.Valid {
			if !moduleID.Valid {
				reasons = append(reasons, "lesson reference has no module_id")
			} else if _, ok := catalog.LessonByCourse(courseID, moduleID.String, lessonID.String); !ok {
				reasons = append(reasons, "lesson does not exist")
			}
		}
		if exerciseID.Valid {
			if !moduleID.Valid {
				reasons = append(reasons, "exercise reference has no module_id")
			} else if _, ok := catalog.ExerciseByCourse(courseID, moduleID.String, exerciseID.String); !ok {
				reasons = append(reasons, "exercise does not exist")
			}
		}
		if reviewItemID.Valid {
			if !moduleID.Valid {
				reasons = append(reasons, "review-item reference has no module_id")
			} else if _, ok := catalog.ReviewItemByCourse(courseID, moduleID.String, reviewItemID.String); !ok {
				reasons = append(reasons, "review item does not exist")
			}
		}
		switch kind {
		case "lesson_completed":
			if !lessonID.Valid {
				reasons = append(reasons, "lesson_completed activity has no lesson_id")
			}
			if exerciseID.Valid || reviewItemID.Valid {
				reasons = append(reasons, "lesson_completed activity has unrelated identity columns")
			}
		case "exercise_checked":
			if !exerciseID.Valid {
				reasons = append(reasons, "exercise_checked activity has no exercise_id")
			}
			if lessonID.Valid || reviewItemID.Valid {
				reasons = append(reasons, "exercise_checked activity has unrelated identity columns")
			}
		case "review_completed":
			if !reviewItemID.Valid {
				reasons = append(reasons, "review_completed activity has no review_item_id")
			}
			if lessonID.Valid || exerciseID.Valid {
				reasons = append(reasons, "review_completed activity has unrelated identity columns")
			}
		}
		if len(reasons) != 0 {
			record := fmt.Sprintf("id=%d kind=%s course=%s module=%s lesson=%s exercise=%s review_item=%s", id, kind, courseID, nullable(moduleID), nullable(lessonID), nullable(exerciseID), nullable(reviewItemID))
			findings = append(findings, Finding{"activities", record, strings.Join(reasons, "; ")})
		}
	}
	return findings, rows.Err()
}

func qualified(parts ...string) string { return strings.Join(parts, "/") }

func nullable(value sql.NullString) string {
	if !value.Valid {
		return "-"
	}
	return value.String
}
