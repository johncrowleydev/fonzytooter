package tutor

import (
	"context"
	"errors"
	"testing"

	"github.com/johncrowleydev/helix-academy/server/internal/auth"
)

const secondTutorUserID auth.UserID = "00000000-0000-4000-8000-000000000002"

func TestConversationStoreIsIsolatedByUser(t *testing.T) {
	store, db := newTestConversationStore(t)
	if _, err := db.Exec(`
		INSERT INTO users (id, display_name, created_at, updated_at)
		VALUES (?, 'Second learner', '2026-08-21T00:00:00.000000000Z', '2026-08-21T00:00:00.000000000Z')
	`, secondTutorUserID); err != nil {
		t.Fatalf("insert second learner: %v", err)
	}
	store.newID = nextID(t, "first-conversation", "first-message", "rejected-message", "second-conversation", "second-message")
	ctx := context.Background()

	first, err := store.CreateConversation(ctx, testUserID, CreateConversationParams{CourseID: "course", Title: "First"})
	if err != nil {
		t.Fatalf("create first learner conversation: %v", err)
	}
	if _, err := store.AppendTextMessage(ctx, testUserID, first.ID, MessageRoleUser, "private first message"); err != nil {
		t.Fatalf("append first learner message: %v", err)
	}
	if _, err := store.Conversation(ctx, secondTutorUserID, first.ID); !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("second learner accessed first learner conversation: %v", err)
	}
	if _, err := store.AppendTextMessage(ctx, secondTutorUserID, first.ID, MessageRoleUser, "cross-user write"); !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("second learner wrote first learner conversation: %v", err)
	}
	secondList, err := store.ListConversations(ctx, secondTutorUserID)
	if err != nil || len(secondList) != 0 {
		t.Fatalf("second learner listed first learner conversation: %#v, %v", secondList, err)
	}

	second, err := store.CreateConversation(ctx, secondTutorUserID, CreateConversationParams{CourseID: "course", Title: "Second"})
	if err != nil {
		t.Fatalf("create second learner conversation: %v", err)
	}
	if _, err := store.AppendTextMessage(ctx, secondTutorUserID, second.ID, MessageRoleUser, "private second message"); err != nil {
		t.Fatalf("append second learner message: %v", err)
	}
	firstMessages, _ := store.Messages(ctx, testUserID, first.ID)
	secondMessages, _ := store.Messages(ctx, secondTutorUserID, second.ID)
	if len(firstMessages) != 1 || firstMessages[0].Parts[0].Text != "private first message" || len(secondMessages) != 1 || secondMessages[0].Parts[0].Text != "private second message" {
		t.Fatalf("conversation messages crossed users: first=%#v second=%#v", firstMessages, secondMessages)
	}
}
