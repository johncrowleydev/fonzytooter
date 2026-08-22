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

func TestConversationStoreListOrdersFractionalTimestampsChronologically(t *testing.T) {
	store, db := newTestConversationStore(t)
	base := time.Date(2026, time.August, 21, 12, 0, 0, 0, time.UTC)
	store.now = nextTime(t, base.Add(100*time.Millisecond), base.Add(110*time.Millisecond), base.Add(90*time.Millisecond))
	store.newID = nextID(t, "conversation-100ms", "conversation-110ms", "conversation-90ms")

	for _, title := range []string{"100ms", "110ms", "90ms"} {
		if _, err := store.CreateConversation(context.Background(), CreateConversationParams{Title: title}); err != nil {
			t.Fatalf("create %s conversation: %v", title, err)
		}
	}

	// Preserve compatibility with timestamps written before the fixed-width
	// formatter was introduced.
	if _, err := db.Exec(`UPDATE tutor_conversations SET updated_at = '2026-08-21T12:00:00.1Z' WHERE id = 'conversation-100ms'`); err != nil {
		t.Fatalf("write legacy variable-width timestamp: %v", err)
	}

	conversations, err := store.ListConversations(context.Background())
	if err != nil {
		t.Fatalf("list conversations: %v", err)
	}
	want := []string{"conversation-110ms", "conversation-100ms", "conversation-90ms"}
	if len(conversations) != len(want) {
		t.Fatalf("expected %d conversations, got %#v", len(want), conversations)
	}
	for index, id := range want {
		if conversations[index].ID != id {
			t.Fatalf("expected chronological order %v, got %#v", want, conversations)
		}
	}

	var stored string
	if err := db.QueryRow(`SELECT updated_at FROM tutor_conversations WHERE id = 'conversation-110ms'`).Scan(&stored); err != nil {
		t.Fatalf("read stored timestamp: %v", err)
	}
	if stored != "2026-08-21T12:00:00.110000000Z" {
		t.Fatalf("expected fixed-width timestamp, got %q", stored)
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

func TestConversationStoreAppendAssistantResponseIsAtomic(t *testing.T) {
	store, db := newTestConversationStore(t)
	store.newID = nextID(t, "conversation", "assistant-message", "tool-1", "tool-2")
	conversation, err := store.CreateConversation(context.Background(), CreateConversationParams{})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	if _, err := db.Exec(`
		CREATE TRIGGER reject_second_tutor_tool_call
		BEFORE INSERT ON tutor_tool_calls
		WHEN NEW.request_id = 'provider-call-2'
		BEGIN SELECT RAISE(ABORT, 'reject second tool call'); END
	`); err != nil {
		t.Fatalf("create rejection trigger: %v", err)
	}

	_, err = store.AppendAssistantResponse(context.Background(), conversation.ID,
		[]ContentPart{{Kind: ContentKindText, Text: "I will inspect both resources."}},
		[]ToolCallInput{
			{RequestID: "provider-call-1", Name: "first_tool", Arguments: json.RawMessage(`{"value":1}`)},
			{RequestID: "provider-call-2", Name: "second_tool", Arguments: json.RawMessage(`{"value":2}`)},
		},
	)
	if err == nil {
		t.Fatal("expected atomic assistant response append to fail")
	}
	queries := []struct {
		name  string
		query string
		arg   string
	}{
		{name: "messages", query: `SELECT COUNT(*) FROM tutor_messages WHERE conversation_id = ?`, arg: conversation.ID},
		{name: "message parts", query: `SELECT COUNT(*) FROM tutor_message_parts WHERE message_id = ?`, arg: "assistant-message"},
		{name: "tool calls", query: `SELECT COUNT(*) FROM tutor_tool_calls WHERE conversation_id = ?`, arg: conversation.ID},
	}
	for _, check := range queries {
		var count int
		if err := db.QueryRow(check.query, check.arg).Scan(&count); err != nil {
			t.Fatalf("count %s: %v", check.name, err)
		}
		if count != 0 {
			t.Fatalf("expected rollback to leave no %s, got %d rows", check.name, count)
		}
	}
}

func TestConversationStoreAppendAssistantResponseAndRecoverPendingCalls(t *testing.T) {
	store, _ := newTestConversationStore(t)
	base := time.Date(2026, time.August, 21, 12, 0, 0, 0, time.UTC)
	store.now = nextTime(t,
		base,
		base.Add(time.Second),
		base.Add(2*time.Second),
		base.Add(3*time.Second),
		base.Add(4*time.Second),
		base.Add(5*time.Second),
	)
	store.newID = nextID(t, "conversation", "assistant-message", "tool-1", "tool-2")
	conversation, err := store.CreateConversation(context.Background(), CreateConversationParams{})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	response, err := store.AppendAssistantResponse(context.Background(), conversation.ID, nil, []ToolCallInput{
		{RequestID: "provider-call-1", Name: "first_tool", Arguments: json.RawMessage(`{"value":1}`)},
		{RequestID: "provider-call-2", Name: "second_tool", Arguments: json.RawMessage(`{"value":2}`)},
	})
	if err != nil {
		t.Fatalf("append assistant response: %v", err)
	}
	if response.Message.Role != MessageRoleAssistant || len(response.ToolCalls) != 2 {
		t.Fatalf("unexpected assistant response: %#v", response)
	}
	if _, err := store.CompleteToolCall(context.Background(), response.ToolCalls[0].ID, json.RawMessage(`{"ok":true}`), ""); err != nil {
		t.Fatalf("complete first tool call: %v", err)
	}
	recovered, err := store.RecoverPendingToolCalls(context.Background(), conversation.ID)
	if err != nil {
		t.Fatalf("recover pending calls: %v", err)
	}
	if recovered != 1 {
		t.Fatalf("expected one recovered call, got %d", recovered)
	}

	calls, err := store.ToolCalls(context.Background(), conversation.ID)
	if err != nil {
		t.Fatalf("list tool calls: %v", err)
	}
	if calls[0].Status != ToolCallCompleted {
		t.Fatalf("expected completed call to remain completed, got %#v", calls[0])
	}
	if calls[1].Status != ToolCallFailed || calls[1].Error != InterruptedToolCallError || calls[1].CompletedAt == nil {
		t.Fatalf("expected pending call to be marked interrupted, got %#v", calls[1])
	}
	if recovered, err := store.RecoverPendingToolCalls(context.Background(), conversation.ID); err != nil || recovered != 0 {
		t.Fatalf("expected idempotent recovery, got count %d and error %v", recovered, err)
	}
	if _, err := store.RecoverPendingToolCalls(context.Background(), "missing"); !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("expected missing conversation error, got %v", err)
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

func TestConversationStoreDatabaseRejectsPartialProviderContinuationState(t *testing.T) {
	store, db := newTestConversationStore(t)
	store.newID = nextID(t, "conversation", "assistant-message")
	conversation, err := store.CreateConversation(context.Background(), CreateConversationParams{})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	message, err := store.AppendTextMessage(context.Background(), conversation.ID, MessageRoleAssistant, "answer")
	if err != nil {
		t.Fatalf("append assistant message: %v", err)
	}

	tests := []struct {
		name     string
		role     string
		provider any
		model    any
		state    any
	}{
		{name: "provider only", role: "assistant", provider: "openrouter"},
		{name: "model only", role: "assistant", model: "model"},
		{name: "state only", role: "assistant", state: `[]`},
		{name: "provider and model", role: "assistant", provider: "openrouter", model: "model"},
		{name: "provider and state", role: "assistant", provider: "openrouter", state: `[]`},
		{name: "model and state", role: "assistant", model: "model", state: `[]`},
		{name: "empty provider", role: "assistant", provider: "", model: "model", state: `[]`},
		{name: "blank model", role: "assistant", provider: "openrouter", model: "   ", state: `[]`},
		{name: "non-assistant", role: "user", provider: "openrouter", model: "model", state: `[]`},
	}
	for index, test := range tests {
		t.Run("insert/"+test.name, func(t *testing.T) {
			_, err := db.Exec(`
				INSERT INTO tutor_messages (
					id, conversation_id, sequence, role, created_at,
					continuation_provider, continuation_model, continuation_state_json
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			`, fmt.Sprintf("invalid-message-%d", index), conversation.ID, index+2, test.role,
				"2026-08-21T12:00:00.000000000Z", test.provider, test.model, test.state)
			if err == nil {
				t.Fatal("expected continuation insert trigger to reject malformed state")
			}
		})
		t.Run("update/"+test.name, func(t *testing.T) {
			_, err := db.Exec(`
				UPDATE tutor_messages
				SET role = ?, continuation_provider = ?, continuation_model = ?, continuation_state_json = ?
				WHERE id = ?
			`, test.role, test.provider, test.model, test.state, message.ID)
			if err == nil {
				t.Fatal("expected continuation update trigger to reject malformed state")
			}
		})
	}

	if _, err := db.Exec(`
		UPDATE tutor_messages
		SET continuation_provider = 'openrouter', continuation_model = 'model', continuation_state_json = '[]'
		WHERE id = ?
	`, message.ID); err != nil {
		t.Fatalf("write complete continuation state: %v", err)
	}
	if _, err := db.Exec(`
		UPDATE tutor_messages
		SET continuation_provider = NULL, continuation_model = NULL, continuation_state_json = NULL
		WHERE id = ?
	`, message.ID); err != nil {
		t.Fatalf("clear complete continuation state: %v", err)
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

func TestConversationStorePersistsCompactionMemoryAndAdvancesMarker(t *testing.T) {
	store, _ := newTestConversationStore(t)
	conversation, err := store.CreateConversation(context.Background(), CreateConversationParams{CourseID: "ai-ml"})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	for _, text := range []string{"question one", "answer one", "question two"} {
		role := MessageRoleUser
		if text == "answer one" {
			role = MessageRoleAssistant
		}
		if _, err := store.AppendTextMessage(context.Background(), conversation.ID, role, text); err != nil {
			t.Fatalf("append message: %v", err)
		}
	}
	memory, err := store.SaveConversationMemory(context.Background(), ConversationMemory{
		ConversationID:            conversation.ID,
		Summary:                   "The learner is comparing two explanations.",
		SummarizedThroughSequence: 2,
		Structured: StructuredMemory{
			LearnerGoal:            "Understand composition",
			Misconceptions:         []string{"Composition is commutative"},
			UnresolvedQuestions:    []string{"Why does order matter?"},
			SourceIDs:              []string{"src.functions"},
			UnsuccessfulApproaches: []string{"Only showing symbolic notation"},
		},
	})
	if err != nil {
		t.Fatalf("save conversation memory: %v", err)
	}
	if memory.FormatVersion != ConversationMemoryFormatVersion || memory.SummarizedThroughSequence != 2 || memory.Structured.LearnerGoal != "Understand composition" {
		t.Fatalf("unexpected memory: %#v", memory)
	}
	loaded, err := store.ConversationMemory(context.Background(), conversation.ID)
	if err != nil {
		t.Fatalf("load conversation memory: %v", err)
	}
	if loaded.Summary != memory.Summary || len(loaded.Structured.Misconceptions) != 1 {
		t.Fatalf("unexpected loaded memory: %#v", loaded)
	}
	if _, err := store.SaveConversationMemory(context.Background(), ConversationMemory{
		ConversationID:            conversation.ID,
		Summary:                   "stale",
		SummarizedThroughSequence: 1,
	}); !errors.Is(err, ErrCompactionMarkerRegression) {
		t.Fatalf("expected marker regression error, got %v", err)
	}
	advanced, err := store.SaveConversationMemory(context.Background(), ConversationMemory{
		ConversationID:            conversation.ID,
		Summary:                   "advanced",
		SummarizedThroughSequence: 3,
	})
	if err != nil {
		t.Fatalf("advance conversation memory: %v", err)
	}
	if advanced.SummarizedThroughSequence != 3 || advanced.Summary != "advanced" || !advanced.CreatedAt.Equal(memory.CreatedAt) {
		t.Fatalf("unexpected advanced memory: %#v", advanced)
	}
}

func TestConversationMemoryRequiresKnownConversationAndCascades(t *testing.T) {
	store, db := newTestConversationStore(t)
	if _, err := store.ConversationMemory(context.Background(), "missing"); !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("expected unknown conversation error, got %v", err)
	}
	conversation, err := store.CreateConversation(context.Background(), CreateConversationParams{})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	if _, err := store.AppendTextMessage(context.Background(), conversation.ID, MessageRoleUser, "remember this"); err != nil {
		t.Fatalf("append message: %v", err)
	}
	if _, err := store.SaveConversationMemory(context.Background(), ConversationMemory{ConversationID: conversation.ID, Summary: "memory", SummarizedThroughSequence: 1}); err != nil {
		t.Fatalf("save memory: %v", err)
	}
	if _, err := db.Exec(`DELETE FROM tutor_conversations WHERE id = ?`, conversation.ID); err != nil {
		t.Fatalf("delete conversation: %v", err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM tutor_conversation_memory`).Scan(&count); err != nil {
		t.Fatalf("count memory rows: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected memory cascade, got %d rows", count)
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
