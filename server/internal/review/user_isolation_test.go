package review

import (
	"context"
	"testing"

	"github.com/johncrowleydev/helix-academy/server/internal/auth"
)

const secondReviewUserID auth.UserID = "00000000-0000-4000-8000-000000000002"

func TestReviewStateIsIsolatedByUser(t *testing.T) {
	db, service := newTestService(t, testNow)
	if _, err := db.Exec(`
		INSERT INTO users (id, display_name, created_at, updated_at)
		VALUES (?, 'Second learner', ?, ?)
	`, secondReviewUserID, formatTime(testNow), formatTime(testNow)); err != nil {
		t.Fatalf("insert second learner: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO lesson_progress (user_id, course_id, module_id, lesson_id, completed, completed_at, updated_at)
		VALUES (?, 'course', 'module', 'lesson', 1, ?, ?)
	`, secondReviewUserID, formatTime(testNow), formatTime(testNow)); err != nil {
		t.Fatalf("complete second learner source lesson: %v", err)
	}

	ctx := context.Background()
	if _, err := service.Submit(ctx, reviewTestUserID, "course", "module", "first", RatingGood); err != nil {
		t.Fatalf("submit first learner review: %v", err)
	}
	secondCards, err := service.Cards(ctx, secondReviewUserID, "course", false)
	if err != nil || len(secondCards) != 2 || !secondCards[0].Virtual {
		t.Fatalf("second learner observed first learner card: %#v, %v", secondCards, err)
	}
	secondHistory, err := service.History(ctx, secondReviewUserID, "course", "module", "first", 10)
	if err != nil || len(secondHistory) != 0 {
		t.Fatalf("second learner observed first learner history: %#v, %v", secondHistory, err)
	}
	if _, err := service.Submit(ctx, secondReviewUserID, "course", "module", "first", RatingAgain); err != nil {
		t.Fatalf("submit second learner review with same identity: %v", err)
	}

	var cards, logs int
	if err := db.QueryRow("SELECT COUNT(*) FROM review_cards WHERE review_item_id = 'first'").Scan(&cards); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM review_logs WHERE review_item_id = 'first'").Scan(&logs); err != nil {
		t.Fatal(err)
	}
	if cards != 2 || logs != 2 {
		t.Fatalf("same review identity was not independently stored: cards=%d logs=%d", cards, logs)
	}
}
