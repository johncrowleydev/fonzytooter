package tutor

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/johncrowleydev/fonzytooter/server/internal/database"
)

func TestConversationStoreCreateReadAndStableList(t *testing.T) {
	store, _ := newTestConversationStore(t)
	base := time.Date(2026, time.August, 21, 12, 0, 0, 0, time.UTC)
	times := []time.Time{base, base.Add(time.Minute), base.Add(2 * time.Minute)}
	store.now = nextTime(t, times...)
	store.newID = nextID(t, "conversation-1", "conversation-2", "message-1")

	first, err := store.CreateConversation(context.Background(), CreateConversationParams{CourseID: "ai-ml", Title: "Gradient descent"})
	if err != nil {
		t.Fatalf("create first conversation: %v", err)
	}
	second, err := store.CreateConversation(context.Background(), CreateConversationParams{Title: "General question"})
	if err != nil {
		t.Fatalf("create second conversation: %v", err)
	}
	if _, err := store.AppendTextMessage(context.Background(), first.ID, MessageRoleUser, "Why does the gradient point uphill?"); err != nil {
		t.Fatalf("append first conversation message: %v", err)
	}

	loaded, err := store.Conversation(context.Background(), first.ID)
	if err != nil {
		t.Fatalf("read conversation: %v", err)
	}
	if loaded.ID != first.ID || loaded.CourseID != "ai-ml" || loaded.Title != "Gradient descent" {
		t.Fatalf("unexpected loaded conversation: %#v", loaded)
	}
	if !loaded.CreatedAt.Equal(base) || !loaded.UpdatedAt.Equal(base.Add(2*time.Minute)) {
		t.Fatalf("unexpected conversation timestamps: %#v", loaded)
	}

	conversations, err := store.ListConversations(context.Background())
	if err != nil {
		t.Fatalf("list conversations: %v", err)
	}
	if len(conversations) != 2 || conversations[0].ID != first.ID || conversations[1].ID != second.ID {
		t.Fatalf("expected most recently updated stable order, got %#v", conversations)
	}
}

func TestConversationStoreMessagesAndRecentWindow(t *testing.T) {
	store, _ := newTestConversationStore(t)
	store.newID = nextID(t, "conversation", "message-1", "message-2", "message-3", "message-4")
	conversation, err := store.CreateConversation(context.Background(), CreateConversationParams{CourseID: "ai-ml"})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	roles := []MessageRole{MessageRoleUser, MessageRoleAssistant, MessageRoleUser, MessageRoleAssistant}
	for index, role := range roles {
		parts := []ContentPart{{Kind: ContentKindText, Text: fmt.Sprintf("part %d-a", index)}, {Kind: ContentKindText, Text: fmt.Sprintf("part %d-b", index)}}
		message, err := store.AppendMessage(context.Background(), conversation.ID, role, parts)
		if err != nil {
			t.Fatalf("append message %d: %v", index, err)
		}
		if message.Sequence != index+1 {
			t.Fatalf("expected sequence %d, got %d", index+1, message.Sequence)
		}
	}

	messages, err := store.Messages(context.Background(), conversation.ID)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(messages) != 4 {
		t.Fatalf("expected four messages, got %#v", messages)
	}
	for index, message := range messages {
		if message.Sequence != index+1 || message.Role != roles[index] || len(message.Parts) != 2 {
			t.Fatalf("unexpected message %d: %#v", index, message)
		}
	}

	recent, err := store.RecentMessages(context.Background(), conversation.ID, 2)
	if err != nil {
		t.Fatalf("list recent messages: %v", err)
	}
	if len(recent) != 2 || recent[0].Sequence != 3 || recent[1].Sequence != 4 {
		t.Fatalf("expected canonical order for recent tail, got %#v", recent)
	}
}

func TestConversationStoreToolCallCompletionAndFailure(t *testing.T) {
	store, _ := newTestConversationStore(t)
	store.newID = nextID(t, "conversation", "assistant-message", "tool-1", "tool-2")
	conversation, err := store.CreateConversation(context.Background(), CreateConversationParams{CourseID: "ai-ml"})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	message, err := store.AppendMessage(context.Background(), conversation.ID, MessageRoleAssistant, nil)
	if err != nil {
		t.Fatalf("append assistant message: %v", err)
	}

	first, err := store.RecordToolCall(context.Background(), RecordToolCallParams{ConversationID: conversation.ID, MessageID: message.ID, RequestID: "provider-call-1", Name: "get_objective_state", Arguments: json.RawMessage(`{"objectiveIds":["python.functions"]}`)})
	if err != nil {
		t.Fatalf("record first tool call: %v", err)
	}
	second, err := store.RecordToolCall(context.Background(), RecordToolCallParams{ConversationID: conversation.ID, MessageID: message.ID, RequestID: "provider-call-2", Name: "search_curriculum", Arguments: json.RawMessage(`{"query":"composition"}`)})
	if err != nil {
		t.Fatalf("record second tool call: %v", err)
	}
	completed, err := store.CompleteToolCall(context.Background(), first.ID, json.RawMessage(`{"introduced":true}`), "")
	if err != nil {
		t.Fatalf("complete first tool call: %v", err)
	}
	failed, err := store.CompleteToolCall(context.Background(), second.ID, nil, "curriculum unavailable")
	if err != nil {
		t.Fatalf("fail second tool call: %v", err)
	}
	if completed.RequestID != "provider-call-1" || completed.Status != ToolCallCompleted || string(completed.Result) != `{"introduced":true}` || completed.CompletedAt == nil {
		t.Fatalf("unexpected completed tool call: %#v", completed)
	}
	if failed.Status != ToolCallFailed || failed.Error != "curriculum unavailable" || failed.CompletedAt == nil {
		t.Fatalf("unexpected failed tool call: %#v", failed)
	}

	toolCalls, err := store.ToolCalls(context.Background(), conversation.ID)
	if err != nil {
		t.Fatalf("list tool calls: %v", err)
	}
	if len(toolCalls) != 2 || toolCalls[0].ID != first.ID || toolCalls[1].ID != second.ID {
		t.Fatalf("unexpected tool call order: %#v", toolCalls)
	}
	if _, err := store.CompleteToolCall(context.Background(), first.ID, json.RawMessage(`null`), ""); !errors.Is(err, ErrToolCallAlreadyCompleted) {
		t.Fatalf("expected already-completed error, got %v", err)
	}
}

func TestConversationStoreRejectsUnknownAndInvalidInput(t *testing.T) {
	store, _ := newTestConversationStore(t)
	ctx := context.Background()

	if _, err := store.Conversation(ctx, "missing"); !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("expected missing conversation error, got %v", err)
	}
	if _, err := store.Messages(ctx, "missing"); !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("expected missing conversation messages error, got %v", err)
	}
	if _, err := store.AppendTextMessage(ctx, "missing", MessageRoleUser, "hello"); !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("expected append missing conversation error, got %v", err)
	}
	conversation, err := store.CreateConversation(ctx, CreateConversationParams{})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	message, err := store.AppendTextMessage(ctx, conversation.ID, MessageRoleUser, "hello")
	if err != nil {
		t.Fatalf("append message: %v", err)
	}
	if _, err := store.AppendTextMessage(ctx, conversation.ID, MessageRole("provider-role"), "hello"); !errors.Is(err, ErrInvalidMessageRole) {
		t.Fatalf("expected invalid role error, got %v", err)
	}
	if _, err := store.AppendTextMessage(ctx, conversation.ID, MessageRoleUser, "  "); !errors.Is(err, ErrInvalidMessageContentPart) {
		t.Fatalf("expected invalid content error, got %v", err)
	}
	if _, err := store.RecentMessages(ctx, conversation.ID, 0); err == nil {
		t.Fatal("expected invalid recent limit error")
	}
	if _, err := store.RecentMessages(ctx, conversation.ID, MaxRecentMessageLimit+1); err == nil {
		t.Fatal("expected excessive recent limit error")
	}
	if _, err := store.RecordToolCall(ctx, RecordToolCallParams{ConversationID: conversation.ID, MessageID: message.ID, RequestID: "call", Name: "tool", Arguments: json.RawMessage(`not-json`)}); !errors.Is(err, ErrInvalidToolArguments) {
		t.Fatalf("expected invalid arguments error, got %v", err)
	}
	if _, err := store.RecordToolCall(ctx, RecordToolCallParams{ConversationID: conversation.ID, MessageID: "missing", RequestID: "call", Name: "tool", Arguments: json.RawMessage(`{}`)}); !errors.Is(err, ErrMessageNotFound) {
		t.Fatalf("expected missing message error, got %v", err)
	}
	if _, err := store.CompleteToolCall(ctx, "missing", json.RawMessage(`null`), ""); !errors.Is(err, ErrToolCallNotFound) {
		t.Fatalf("expected missing tool call error, got %v", err)
	}
}

func TestConversationStoreEnforcesForeignKeysAndCascades(t *testing.T) {
	store, db := newTestConversationStore(t)
	store.newID = nextID(t, "conversation-1", "conversation-2", "message", "cross-tool", "tool")
	first, err := store.CreateConversation(context.Background(), CreateConversationParams{})
	if err != nil {
		t.Fatalf("create first conversation: %v", err)
	}
	second, err := store.CreateConversation(context.Background(), CreateConversationParams{})
	if err != nil {
		t.Fatalf("create second conversation: %v", err)
	}
	message, err := store.AppendTextMessage(context.Background(), first.ID, MessageRoleAssistant, "checking")
	if err != nil {
		t.Fatalf("append message: %v", err)
	}
	if _, err := store.RecordToolCall(context.Background(), RecordToolCallParams{ConversationID: second.ID, MessageID: message.ID, RequestID: "cross-call", Name: "tool", Arguments: json.RawMessage(`{}`)}); !errors.Is(err, ErrMessageNotFound) {
		t.Fatalf("expected cross-conversation message rejection, got %v", err)
	}
	if _, err := store.RecordToolCall(context.Background(), RecordToolCallParams{ConversationID: first.ID, MessageID: message.ID, RequestID: "call", Name: "tool", Arguments: json.RawMessage(`{}`)}); err != nil {
		t.Fatalf("record tool call: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO tutor_messages (id, conversation_id, sequence, role, created_at) VALUES ('orphan', 'missing', 1, 'user', '2026-08-21T12:00:00Z')`); err == nil {
		t.Fatal("expected direct foreign-key violation")
	}
	if _, err := db.Exec(`DELETE FROM tutor_conversations WHERE id = ?`, first.ID); err != nil {
		t.Fatalf("delete conversation: %v", err)
	}
	for _, table := range []string{"tutor_messages", "tutor_message_parts", "tutor_tool_calls"} {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&count); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("expected cascade to clear %s, got %d rows", table, count)
		}
	}
}

func TestConversationStoreAppendIsAtomic(t *testing.T) {
	store, db := newTestConversationStore(t)
	store.newID = nextID(t, "conversation", "message")
	conversation, err := store.CreateConversation(context.Background(), CreateConversationParams{})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	if _, err := db.Exec(`CREATE TRIGGER reject_tutor_part BEFORE INSERT ON tutor_message_parts BEGIN SELECT RAISE(ABORT, 'reject part'); END`); err != nil {
		t.Fatalf("create rejection trigger: %v", err)
	}
	if _, err := store.AppendTextMessage(context.Background(), conversation.ID, MessageRoleUser, "hello"); err == nil {
		t.Fatal("expected append error")
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM tutor_messages WHERE conversation_id = ?`, conversation.ID).Scan(&count); err != nil {
		t.Fatalf("count messages: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected transaction rollback, got %d messages", count)
	}
}

func TestConversationStorePersistsAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fonzytooter.db")
	firstDB, err := database.Open(context.Background(), path)
	if err != nil {
		t.Fatalf("open first database: %v", err)
	}
	firstStore := NewConversationStore(firstDB)
	conversation, err := firstStore.CreateConversation(context.Background(), CreateConversationParams{CourseID: "ai-ml"})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	if _, err := firstStore.AppendTextMessage(context.Background(), conversation.ID, MessageRoleUser, "persist me"); err != nil {
		t.Fatalf("append message: %v", err)
	}
	if err := firstDB.Close(); err != nil {
		t.Fatalf("close first database: %v", err)
	}

	secondDB, err := database.Open(context.Background(), path)
	if err != nil {
		t.Fatalf("reopen database: %v", err)
	}
	t.Cleanup(func() { _ = secondDB.Close() })
	messages, err := NewConversationStore(secondDB).Messages(context.Background(), conversation.ID)
	if err != nil {
		t.Fatalf("load reopened messages: %v", err)
	}
	if len(messages) != 1 || messages[0].Parts[0].Text != "persist me" {
		t.Fatalf("unexpected reopened messages: %#v", messages)
	}
}

func newTestConversationStore(t *testing.T) (*ConversationStore, *sql.DB) {
	t.Helper()
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "fonzytooter.db"))
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewConversationStore(db), db
}

func nextID(t *testing.T, ids ...string) func() (string, error) {
	t.Helper()
	index := 0
	return func() (string, error) {
		if index >= len(ids) {
			t.Fatalf("unexpected ID request after %d IDs", len(ids))
		}
		id := ids[index]
		index++
		return id, nil
	}
}

func nextTime(t *testing.T, times ...time.Time) func() time.Time {
	t.Helper()
	index := 0
	return func() time.Time {
		if index >= len(times) {
			t.Fatalf("unexpected time request after %d values", len(times))
		}
		value := times[index]
		index++
		return value
	}
}
