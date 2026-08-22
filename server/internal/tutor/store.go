package tutor

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	DefaultRecentMessageLimit = 20
	MaxRecentMessageLimit     = 200
	InterruptedToolCallError  = "tutor turn interrupted before tool call completion"
	sortableTimeLayout        = "2006-01-02T15:04:05.000000000Z"
)

var (
	ErrConversationNotFound       = errors.New("tutor conversation not found")
	ErrMessageNotFound            = errors.New("tutor message not found")
	ErrToolCallNotFound           = errors.New("tutor tool call not found")
	ErrToolCallAlreadyCompleted   = errors.New("tutor tool call already completed")
	ErrInvalidMessageRole         = errors.New("invalid tutor message role")
	ErrInvalidMessageContentPart  = errors.New("invalid tutor message content part")
	ErrInvalidToolArguments       = errors.New("invalid tutor tool arguments")
	ErrInvalidToolResult          = errors.New("invalid tutor tool result")
	ErrCompactionMarkerRegression = errors.New("tutor compaction marker cannot move backward")
)

type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)

type ContentKind string

const ContentKindText ContentKind = "text"

type ToolCallStatus string

const (
	ToolCallPending   ToolCallStatus = "pending"
	ToolCallCompleted ToolCallStatus = "completed"
	ToolCallFailed    ToolCallStatus = "failed"
)

type Conversation struct {
	ID        string
	CourseID  string
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

const ConversationMemoryFormatVersion = 1

type StructuredMemory struct {
	LearnerGoal              string   `json:"learnerGoal,omitempty"`
	EstablishedUnderstanding []string `json:"establishedUnderstanding,omitempty"`
	Misconceptions           []string `json:"misconceptions,omitempty"`
	UnsuccessfulApproaches   []string `json:"unsuccessfulApproaches,omitempty"`
	UnresolvedQuestions      []string `json:"unresolvedQuestions,omitempty"`
	ActiveContext            string   `json:"activeContext,omitempty"`
	SourceIDs                []string `json:"sourceIds,omitempty"`
	ToolFindings             []string `json:"toolFindings,omitempty"`
}

type ConversationMemory struct {
	ConversationID            string
	Summary                   string
	Structured                StructuredMemory
	SummarizedThroughSequence int
	FormatVersion             int
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

type ContentPart struct {
	Kind ContentKind
	Text string
}

type Message struct {
	ID             string
	ConversationID string
	Sequence       int
	Role           MessageRole
	Parts          []ContentPart
	Continuation   *ProviderContinuation
	CreatedAt      time.Time
}

type ToolCall struct {
	ID             string
	ConversationID string
	MessageID      string
	Sequence       int
	RequestID      string
	Name           string
	Arguments      json.RawMessage
	Status         ToolCallStatus
	Result         json.RawMessage
	Error          string
	CreatedAt      time.Time
	CompletedAt    *time.Time
}

type CreateConversationParams struct {
	CourseID string
	Title    string
}

type RecordToolCallParams struct {
	ConversationID string
	MessageID      string
	RequestID      string
	Name           string
	Arguments      json.RawMessage
}

type ToolCallInput struct {
	RequestID string
	Name      string
	Arguments json.RawMessage
}

type AssistantResponse struct {
	Message   Message
	ToolCalls []ToolCall
}

type ConversationStore struct {
	db    *sql.DB
	now   func() time.Time
	newID func() (string, error)
}

func NewConversationStore(db *sql.DB) *ConversationStore {
	if db == nil {
		panic("tutor.NewConversationStore: nil database")
	}
	return &ConversationStore{db: db, now: time.Now, newID: randomID}
}

func (s *ConversationStore) CreateConversation(ctx context.Context, params CreateConversationParams) (Conversation, error) {
	if strings.TrimSpace(params.CourseID) != params.CourseID {
		return Conversation{}, errors.New("course ID must not contain surrounding whitespace")
	}
	id, err := s.newID()
	if err != nil {
		return Conversation{}, fmt.Errorf("generate tutor conversation ID: %w", err)
	}
	now := s.now().UTC()
	conversation := Conversation{ID: id, CourseID: params.CourseID, Title: params.Title, CreatedAt: now, UpdatedAt: now}

	var courseID any
	if params.CourseID != "" {
		courseID = params.CourseID
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO tutor_conversations (id, course_id, title, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, id, courseID, params.Title, formatTime(now), formatTime(now)); err != nil {
		return Conversation{}, fmt.Errorf("create tutor conversation: %w", err)
	}
	return conversation, nil
}

func (s *ConversationStore) Conversation(ctx context.Context, id string) (Conversation, error) {
	return scanConversation(s.db.QueryRowContext(ctx, `
		SELECT id, course_id, title, created_at, updated_at
		FROM tutor_conversations
		WHERE id = ?
	`, id))
}

func (s *ConversationStore) ListConversations(ctx context.Context) ([]Conversation, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, course_id, title, created_at, updated_at
		FROM tutor_conversations
	`)
	if err != nil {
		return nil, fmt.Errorf("list tutor conversations: %w", err)
	}
	defer rows.Close()

	conversations := make([]Conversation, 0)
	for rows.Next() {
		conversation, err := scanConversation(rows)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, conversation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list tutor conversations: %w", err)
	}
	sort.Slice(conversations, func(i, j int) bool {
		if !conversations[i].UpdatedAt.Equal(conversations[j].UpdatedAt) {
			return conversations[i].UpdatedAt.After(conversations[j].UpdatedAt)
		}
		return conversations[i].ID > conversations[j].ID
	})
	return conversations, nil
}

func (s *ConversationStore) AppendTextMessage(ctx context.Context, conversationID string, role MessageRole, text string) (Message, error) {
	return s.AppendMessage(ctx, conversationID, role, []ContentPart{{Kind: ContentKindText, Text: text}})
}

func (s *ConversationStore) AppendMessage(ctx context.Context, conversationID string, role MessageRole, parts []ContentPart) (Message, error) {
	response, err := s.appendMessage(ctx, conversationID, role, parts, nil, nil)
	return response.Message, err
}

// AppendAssistantResponse atomically persists one canonical assistant message
// and the complete ordered set of tool calls requested by that model response.
func (s *ConversationStore) AppendAssistantResponse(ctx context.Context, conversationID string, parts []ContentPart, calls []ToolCallInput) (AssistantResponse, error) {
	return s.AppendAssistantResponseWithContinuation(ctx, conversationID, parts, calls, nil)
}

func (s *ConversationStore) AppendAssistantResponseWithContinuation(ctx context.Context, conversationID string, parts []ContentPart, calls []ToolCallInput, continuation *ProviderContinuation) (AssistantResponse, error) {
	return s.appendMessage(ctx, conversationID, MessageRoleAssistant, parts, calls, continuation)
}

func (s *ConversationStore) appendMessage(ctx context.Context, conversationID string, role MessageRole, parts []ContentPart, calls []ToolCallInput, continuation *ProviderContinuation) (AssistantResponse, error) {
	if !validMessageRole(role) {
		return AssistantResponse{}, ErrInvalidMessageRole
	}
	if len(parts) == 0 && role != MessageRoleAssistant {
		return AssistantResponse{}, fmt.Errorf("%w: at least one part is required", ErrInvalidMessageContentPart)
	}
	for _, part := range parts {
		if part.Kind != ContentKindText || strings.TrimSpace(part.Text) == "" {
			return AssistantResponse{}, ErrInvalidMessageContentPart
		}
	}
	if role != MessageRoleAssistant && len(calls) > 0 {
		return AssistantResponse{}, errors.New("only assistant messages can contain tutor tool calls")
	}
	if continuation != nil {
		if role != MessageRoleAssistant || strings.TrimSpace(continuation.Provider) == "" || strings.TrimSpace(continuation.Model) == "" || !json.Valid(continuation.State) {
			return AssistantResponse{}, errors.New("invalid tutor provider continuation state")
		}
	}
	validatedCalls := make([]ToolCallInput, len(calls))
	seenRequestIDs := make(map[string]struct{}, len(calls))
	for index, call := range calls {
		requestID, name, err := validateToolCallInput(call.RequestID, call.Name, call.Arguments)
		if err != nil {
			return AssistantResponse{}, err
		}
		if _, duplicate := seenRequestIDs[requestID]; duplicate {
			return AssistantResponse{}, fmt.Errorf("duplicate tutor tool request ID %q", requestID)
		}
		seenRequestIDs[requestID] = struct{}{}
		validatedCalls[index] = ToolCallInput{RequestID: requestID, Name: name, Arguments: cloneJSON(call.Arguments)}
	}

	id, err := s.newID()
	if err != nil {
		return AssistantResponse{}, fmt.Errorf("generate tutor message ID: %w", err)
	}
	toolCallIDs := make([]string, len(validatedCalls))
	for index := range validatedCalls {
		toolCallIDs[index], err = s.newID()
		if err != nil {
			return AssistantResponse{}, fmt.Errorf("generate tutor tool call ID: %w", err)
		}
	}
	now := s.now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AssistantResponse{}, fmt.Errorf("begin append tutor message: %w", err)
	}
	defer tx.Rollback()

	var sequence int
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(sequence), 0) + 1
		FROM tutor_messages
		WHERE conversation_id = ?
	`, conversationID).Scan(&sequence); err != nil {
		return AssistantResponse{}, fmt.Errorf("choose tutor message sequence: %w", err)
	}
	var continuationProvider, continuationModel, continuationState any
	if continuation != nil {
		continuationProvider = continuation.Provider
		continuationModel = continuation.Model
		continuationState = string(continuation.State)
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO tutor_messages (
			id, conversation_id, sequence, role, created_at,
			continuation_provider, continuation_model, continuation_state_json
		)
		SELECT ?, id, ?, ?, ?, ?, ?, ? FROM tutor_conversations WHERE id = ?
	`, id, sequence, role, formatTime(now), continuationProvider, continuationModel, continuationState, conversationID)
	if err != nil {
		return AssistantResponse{}, fmt.Errorf("append tutor message: %w", err)
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return AssistantResponse{}, fmt.Errorf("inspect appended tutor message: %w", err)
	}
	if inserted == 0 {
		return AssistantResponse{}, ErrConversationNotFound
	}
	for index, part := range parts {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO tutor_message_parts (message_id, sequence, kind, text_content)
			VALUES (?, ?, ?, ?)
		`, id, index+1, part.Kind, part.Text); err != nil {
			return AssistantResponse{}, fmt.Errorf("append tutor message part: %w", err)
		}
	}
	storedCalls := make([]ToolCall, len(validatedCalls))
	for index, call := range validatedCalls {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO tutor_tool_calls (
				id, conversation_id, message_id, sequence, request_id, name, arguments_json, status, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, 'pending', ?)
		`, toolCallIDs[index], conversationID, id, index+1, call.RequestID, call.Name, string(call.Arguments), formatTime(now)); err != nil {
			return AssistantResponse{}, fmt.Errorf("append assistant tutor tool call: %w", err)
		}
		storedCalls[index] = ToolCall{
			ID: toolCallIDs[index], ConversationID: conversationID, MessageID: id, Sequence: index + 1,
			RequestID: call.RequestID, Name: call.Name, Arguments: cloneJSON(call.Arguments), Status: ToolCallPending, CreatedAt: now,
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE tutor_conversations SET updated_at = ? WHERE id = ?`, formatTime(now), conversationID); err != nil {
		return AssistantResponse{}, fmt.Errorf("update tutor conversation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return AssistantResponse{}, fmt.Errorf("commit tutor message: %w", err)
	}

	return AssistantResponse{
		Message: Message{
			ID: id, ConversationID: conversationID, Sequence: sequence, Role: role, Parts: cloneParts(parts),
			Continuation: cloneProviderContinuation(continuation), CreatedAt: now,
		},
		ToolCalls: storedCalls,
	}, nil
}

func (s *ConversationStore) Messages(ctx context.Context, conversationID string) ([]Message, error) {
	return s.messages(ctx, conversationID, 0)
}

func (s *ConversationStore) RecentMessages(ctx context.Context, conversationID string, limit int) ([]Message, error) {
	if limit <= 0 || limit > MaxRecentMessageLimit {
		return nil, fmt.Errorf("recent tutor message limit must be between 1 and %d", MaxRecentMessageLimit)
	}
	return s.messages(ctx, conversationID, limit)
}

func (s *ConversationStore) RecordToolCall(ctx context.Context, params RecordToolCallParams) (ToolCall, error) {
	requestID, name, err := validateToolCallInput(params.RequestID, params.Name, params.Arguments)
	if err != nil {
		return ToolCall{}, err
	}
	id, err := s.newID()
	if err != nil {
		return ToolCall{}, fmt.Errorf("generate tutor tool call ID: %w", err)
	}
	now := s.now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ToolCall{}, fmt.Errorf("begin tutor tool call: %w", err)
	}
	defer tx.Rollback()

	var sequence int
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(sequence), 0) + 1
		FROM tutor_tool_calls
		WHERE message_id = ?
	`, params.MessageID).Scan(&sequence); err != nil {
		return ToolCall{}, fmt.Errorf("choose tutor tool call sequence: %w", err)
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO tutor_tool_calls (
			id, conversation_id, message_id, sequence, request_id, name, arguments_json, status, created_at
		)
		SELECT ?, conversation_id, id, ?, ?, ?, ?, 'pending', ?
		FROM tutor_messages
		WHERE id = ? AND conversation_id = ?
	`, id, sequence, requestID, name, string(params.Arguments), formatTime(now), params.MessageID, params.ConversationID)
	if err != nil {
		return ToolCall{}, fmt.Errorf("record tutor tool call: %w", err)
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return ToolCall{}, fmt.Errorf("inspect recorded tutor tool call: %w", err)
	}
	if inserted == 0 {
		return ToolCall{}, ErrMessageNotFound
	}
	if _, err := tx.ExecContext(ctx, `UPDATE tutor_conversations SET updated_at = ? WHERE id = ?`, formatTime(now), params.ConversationID); err != nil {
		return ToolCall{}, fmt.Errorf("update tutor conversation for tool call: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return ToolCall{}, fmt.Errorf("commit tutor tool call: %w", err)
	}

	return ToolCall{ID: id, ConversationID: params.ConversationID, MessageID: params.MessageID, Sequence: sequence, RequestID: requestID, Name: name, Arguments: cloneJSON(params.Arguments), Status: ToolCallPending, CreatedAt: now}, nil
}

func (s *ConversationStore) CompleteToolCall(ctx context.Context, id string, result json.RawMessage, executionError string) (ToolCall, error) {
	if executionError != "" && len(result) != 0 {
		return ToolCall{}, errors.New("failed tutor tool call cannot have a result")
	}
	if len(result) != 0 && !json.Valid(result) {
		return ToolCall{}, ErrInvalidToolResult
	}
	now := s.now().UTC()
	status := ToolCallCompleted
	var storedResult, storedError any
	if executionError != "" {
		status = ToolCallFailed
		storedError = executionError
	} else if len(result) != 0 {
		storedResult = string(result)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ToolCall{}, fmt.Errorf("begin complete tutor tool call: %w", err)
	}
	defer tx.Rollback()
	update, err := tx.ExecContext(ctx, `
		UPDATE tutor_tool_calls
		SET status = ?, result_json = ?, error = ?, completed_at = ?
		WHERE id = ? AND status = 'pending'
	`, status, storedResult, storedError, formatTime(now), id)
	if err != nil {
		return ToolCall{}, fmt.Errorf("complete tutor tool call: %w", err)
	}
	changed, err := update.RowsAffected()
	if err != nil {
		return ToolCall{}, fmt.Errorf("inspect completed tutor tool call: %w", err)
	}
	if changed == 0 {
		var existingStatus string
		err := tx.QueryRowContext(ctx, `SELECT status FROM tutor_tool_calls WHERE id = ?`, id).Scan(&existingStatus)
		if errors.Is(err, sql.ErrNoRows) {
			return ToolCall{}, ErrToolCallNotFound
		}
		if err != nil {
			return ToolCall{}, fmt.Errorf("read tutor tool call status: %w", err)
		}
		return ToolCall{}, ErrToolCallAlreadyCompleted
	}
	toolCall, err := scanToolCall(tx.QueryRowContext(ctx, `
		SELECT id, conversation_id, message_id, sequence, request_id, name, arguments_json,
		       status, result_json, error, created_at, completed_at
		FROM tutor_tool_calls
		WHERE id = ?
	`, id))
	if err != nil {
		return ToolCall{}, fmt.Errorf("read completed tutor tool call: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE tutor_conversations SET updated_at = ? WHERE id = ?`, formatTime(now), toolCall.ConversationID); err != nil {
		return ToolCall{}, fmt.Errorf("update tutor conversation for tool result: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return ToolCall{}, fmt.Errorf("commit completed tutor tool call: %w", err)
	}

	return toolCall, nil
}

// RecoverPendingToolCalls marks every still-pending call in a conversation as
// failed. The runtime invokes this before accepting a new turn so a process
// interruption is represented explicitly and canonical history is replayable.
func (s *ConversationStore) RecoverPendingToolCalls(ctx context.Context, conversationID string) (int64, error) {
	now := s.now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin recover pending tutor tool calls: %w", err)
	}
	defer tx.Rollback()
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM tutor_conversations WHERE id = ?`, conversationID).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return 0, ErrConversationNotFound
	} else if err != nil {
		return 0, fmt.Errorf("read tutor conversation for recovery: %w", err)
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE tutor_tool_calls
		SET status = 'failed', error = ?, completed_at = ?
		WHERE conversation_id = ? AND status = 'pending'
	`, InterruptedToolCallError, formatTime(now), conversationID)
	if err != nil {
		return 0, fmt.Errorf("recover pending tutor tool calls: %w", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("inspect recovered tutor tool calls: %w", err)
	}
	if changed > 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE tutor_conversations SET updated_at = ? WHERE id = ?`, formatTime(now), conversationID); err != nil {
			return 0, fmt.Errorf("update tutor conversation for recovery: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit recovered tutor tool calls: %w", err)
	}
	return changed, nil
}

func (s *ConversationStore) ToolCalls(ctx context.Context, conversationID string) ([]ToolCall, error) {
	if _, err := s.Conversation(ctx, conversationID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT tc.id, tc.conversation_id, tc.message_id, tc.sequence, tc.request_id, tc.name,
		       tc.arguments_json, tc.status, tc.result_json, tc.error,
		       tc.created_at, tc.completed_at
		FROM tutor_tool_calls tc
		JOIN tutor_messages m ON m.id = tc.message_id
		WHERE tc.conversation_id = ?
		ORDER BY m.sequence, tc.sequence
	`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("list tutor tool calls: %w", err)
	}
	defer rows.Close()

	toolCalls := make([]ToolCall, 0)
	for rows.Next() {
		toolCall, err := scanToolCall(rows)
		if err != nil {
			return nil, err
		}
		toolCalls = append(toolCalls, toolCall)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list tutor tool calls: %w", err)
	}
	return toolCalls, nil
}

func (s *ConversationStore) ConversationMemory(ctx context.Context, conversationID string) (ConversationMemory, error) {
	if _, err := s.Conversation(ctx, conversationID); err != nil {
		return ConversationMemory{}, err
	}
	var memory ConversationMemory
	var structuredJSON, createdAt, updatedAt string
	err := s.db.QueryRowContext(ctx, `
		SELECT conversation_id, summary, structured_json, summarized_through_sequence,
		       format_version, created_at, updated_at
		FROM tutor_conversation_memory
		WHERE conversation_id = ?
	`, conversationID).Scan(
		&memory.ConversationID, &memory.Summary, &structuredJSON,
		&memory.SummarizedThroughSequence, &memory.FormatVersion, &createdAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ConversationMemory{ConversationID: conversationID, FormatVersion: ConversationMemoryFormatVersion}, nil
	}
	if err != nil {
		return ConversationMemory{}, fmt.Errorf("read tutor conversation memory: %w", err)
	}
	if err := json.Unmarshal([]byte(structuredJSON), &memory.Structured); err != nil {
		return ConversationMemory{}, fmt.Errorf("decode tutor conversation memory: %w", err)
	}
	memory.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return ConversationMemory{}, fmt.Errorf("parse tutor conversation memory created time: %w", err)
	}
	memory.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return ConversationMemory{}, fmt.Errorf("parse tutor conversation memory updated time: %w", err)
	}
	return memory, nil
}

func (s *ConversationStore) SaveConversationMemory(ctx context.Context, memory ConversationMemory) (ConversationMemory, error) {
	if memory.SummarizedThroughSequence <= 0 {
		return ConversationMemory{}, errors.New("tutor compaction marker must be positive")
	}
	if memory.FormatVersion == 0 {
		memory.FormatVersion = ConversationMemoryFormatVersion
	}
	structuredJSON, err := json.Marshal(memory.Structured)
	if err != nil {
		return ConversationMemory{}, fmt.Errorf("encode tutor conversation memory: %w", err)
	}
	now := s.now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ConversationMemory{}, fmt.Errorf("begin save tutor conversation memory: %w", err)
	}
	defer tx.Rollback()
	var markerExists int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM tutor_messages
		WHERE conversation_id = ? AND sequence = ?
	`, memory.ConversationID, memory.SummarizedThroughSequence).Scan(&markerExists); err != nil {
		return ConversationMemory{}, fmt.Errorf("validate tutor compaction marker: %w", err)
	}
	if markerExists == 0 {
		var conversationExists int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM tutor_conversations WHERE id = ?`, memory.ConversationID).Scan(&conversationExists); err != nil {
			return ConversationMemory{}, fmt.Errorf("validate tutor compaction conversation: %w", err)
		}
		if conversationExists == 0 {
			return ConversationMemory{}, ErrConversationNotFound
		}
		return ConversationMemory{}, ErrMessageNotFound
	}

	var currentMarker int
	err = tx.QueryRowContext(ctx, `
		SELECT summarized_through_sequence
		FROM tutor_conversation_memory
		WHERE conversation_id = ?
	`, memory.ConversationID).Scan(&currentMarker)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return ConversationMemory{}, fmt.Errorf("read tutor compaction marker: %w", err)
	}
	if err == nil && memory.SummarizedThroughSequence < currentMarker {
		return ConversationMemory{}, ErrCompactionMarkerRegression
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO tutor_conversation_memory (
			conversation_id, summary, structured_json, summarized_through_sequence,
			format_version, created_at, updated_at
		)
		SELECT id, ?, ?, ?, ?, ?, ?
		FROM tutor_conversations
		WHERE id = ?
		ON CONFLICT (conversation_id) DO UPDATE SET
			summary = excluded.summary,
			structured_json = excluded.structured_json,
			summarized_through_sequence = excluded.summarized_through_sequence,
			format_version = excluded.format_version,
			updated_at = excluded.updated_at
	`, memory.Summary, string(structuredJSON), memory.SummarizedThroughSequence,
		memory.FormatVersion, formatTime(now), formatTime(now), memory.ConversationID)
	if err != nil {
		return ConversationMemory{}, fmt.Errorf("save tutor conversation memory: %w", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return ConversationMemory{}, fmt.Errorf("inspect saved tutor conversation memory: %w", err)
	}
	if changed == 0 {
		return ConversationMemory{}, ErrConversationNotFound
	}
	if _, err := tx.ExecContext(ctx, `UPDATE tutor_conversations SET updated_at = ? WHERE id = ?`, formatTime(now), memory.ConversationID); err != nil {
		return ConversationMemory{}, fmt.Errorf("update tutor conversation for compaction: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return ConversationMemory{}, fmt.Errorf("commit tutor conversation memory: %w", err)
	}
	return s.ConversationMemory(ctx, memory.ConversationID)
}

func (s *ConversationStore) messages(ctx context.Context, conversationID string, limit int) ([]Message, error) {
	if _, err := s.Conversation(ctx, conversationID); err != nil {
		return nil, err
	}
	query := `
		SELECT m.id, m.conversation_id, m.sequence, m.role, m.created_at,
		       m.continuation_provider, m.continuation_model, m.continuation_state_json,
		       p.kind, p.text_content
		FROM tutor_messages m
		LEFT JOIN tutor_message_parts p ON p.message_id = m.id
		WHERE m.conversation_id = ?
		ORDER BY m.sequence, p.sequence`
	args := []any{conversationID}
	if limit > 0 {
		query = `
			WITH recent_messages AS (
				SELECT id, conversation_id, sequence, role, created_at,
				       continuation_provider, continuation_model, continuation_state_json
				FROM tutor_messages
				WHERE conversation_id = ?
				ORDER BY sequence DESC
				LIMIT ?
			)
			SELECT m.id, m.conversation_id, m.sequence, m.role, m.created_at,
			       m.continuation_provider, m.continuation_model, m.continuation_state_json,
			       p.kind, p.text_content
			FROM recent_messages m
			LEFT JOIN tutor_message_parts p ON p.message_id = m.id
			ORDER BY m.sequence, p.sequence`
		args = append(args, limit)
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tutor messages: %w", err)
	}
	defer rows.Close()

	messages := make([]Message, 0)
	for rows.Next() {
		var message Message
		var role, createdAt string
		var continuationProvider, continuationModel, continuationState sql.NullString
		var partKind, partText sql.NullString
		if err := rows.Scan(
			&message.ID, &message.ConversationID, &message.Sequence, &role, &createdAt,
			&continuationProvider, &continuationModel, &continuationState, &partKind, &partText,
		); err != nil {
			return nil, fmt.Errorf("scan tutor message: %w", err)
		}
		if len(messages) > 0 && messages[len(messages)-1].ID == message.ID {
			if partKind.Valid {
				messages[len(messages)-1].Parts = append(messages[len(messages)-1].Parts, ContentPart{Kind: ContentKind(partKind.String), Text: partText.String})
			}
			continue
		}
		message.Role = MessageRole(role)
		if continuationProvider.Valid || continuationModel.Valid || continuationState.Valid {
			if !continuationProvider.Valid || !continuationModel.Valid || !continuationState.Valid || !json.Valid([]byte(continuationState.String)) {
				return nil, errors.New("scan tutor message: invalid provider continuation state")
			}
			message.Continuation = &ProviderContinuation{
				Provider: continuationProvider.String, Model: continuationModel.String, State: json.RawMessage(continuationState.String),
			}
		}
		message.CreatedAt, err = parseTime(createdAt)
		if err != nil {
			return nil, fmt.Errorf("scan tutor message: %w", err)
		}
		message.Parts = make([]ContentPart, 0)
		if partKind.Valid {
			message.Parts = append(message.Parts, ContentPart{Kind: ContentKind(partKind.String), Text: partText.String})
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list tutor messages: %w", err)
	}
	return messages, nil
}

func (s *ConversationStore) toolCall(ctx context.Context, id string) (ToolCall, error) {
	return scanToolCall(s.db.QueryRowContext(ctx, `
		SELECT id, conversation_id, message_id, sequence, request_id, name, arguments_json,
		       status, result_json, error, created_at, completed_at
		FROM tutor_tool_calls
		WHERE id = ?
	`, id))
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanConversation(row rowScanner) (Conversation, error) {
	var conversation Conversation
	var courseID sql.NullString
	var createdAt, updatedAt string
	if err := row.Scan(&conversation.ID, &courseID, &conversation.Title, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Conversation{}, ErrConversationNotFound
		}
		return Conversation{}, fmt.Errorf("scan tutor conversation: %w", err)
	}
	conversation.CourseID = courseID.String
	var err error
	conversation.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return Conversation{}, fmt.Errorf("scan tutor conversation created time: %w", err)
	}
	conversation.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return Conversation{}, fmt.Errorf("scan tutor conversation updated time: %w", err)
	}
	return conversation, nil
}

func scanToolCall(row rowScanner) (ToolCall, error) {
	var toolCall ToolCall
	var arguments string
	var status string
	var result, executionError, completedAt sql.NullString
	var createdAt string
	if err := row.Scan(
		&toolCall.ID, &toolCall.ConversationID, &toolCall.MessageID, &toolCall.Sequence,
		&toolCall.RequestID, &toolCall.Name, &arguments, &status, &result, &executionError, &createdAt, &completedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ToolCall{}, ErrToolCallNotFound
		}
		return ToolCall{}, fmt.Errorf("scan tutor tool call: %w", err)
	}
	toolCall.Arguments = json.RawMessage(arguments)
	toolCall.Status = ToolCallStatus(status)
	if result.Valid {
		toolCall.Result = json.RawMessage(result.String)
	}
	toolCall.Error = executionError.String
	var err error
	toolCall.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return ToolCall{}, fmt.Errorf("scan tutor tool call created time: %w", err)
	}
	if completedAt.Valid {
		parsed, err := parseTime(completedAt.String)
		if err != nil {
			return ToolCall{}, fmt.Errorf("scan tutor tool call completed time: %w", err)
		}
		toolCall.CompletedAt = &parsed
	}
	return toolCall, nil
}

func validMessageRole(role MessageRole) bool {
	return role == MessageRoleUser || role == MessageRoleAssistant
}

func validateToolCallInput(rawRequestID, rawName string, arguments json.RawMessage) (string, string, error) {
	requestID := strings.TrimSpace(rawRequestID)
	if requestID == "" {
		return "", "", errors.New("tutor tool request ID is empty")
	}
	name := strings.TrimSpace(rawName)
	if name == "" {
		return "", "", errors.New("tutor tool name is empty")
	}
	if !json.Valid(arguments) {
		return "", "", ErrInvalidToolArguments
	}
	return requestID, name, nil
}

func randomID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	encoded := make([]byte, 36)
	hex.Encode(encoded[0:8], value[0:4])
	encoded[8] = '-'
	hex.Encode(encoded[9:13], value[4:6])
	encoded[13] = '-'
	hex.Encode(encoded[14:18], value[6:8])
	encoded[18] = '-'
	hex.Encode(encoded[19:23], value[8:10])
	encoded[23] = '-'
	hex.Encode(encoded[24:36], value[10:16])
	return string(encoded), nil
}

func formatTime(value time.Time) string {
	return value.UTC().Format(sortableTimeLayout)
}

func parseTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, value)
}

func cloneParts(parts []ContentPart) []ContentPart {
	return append([]ContentPart(nil), parts...)
}

func cloneJSON(value json.RawMessage) json.RawMessage {
	return append(json.RawMessage(nil), value...)
}

func cloneProviderContinuation(value *ProviderContinuation) *ProviderContinuation {
	if value == nil {
		return nil
	}
	return &ProviderContinuation{Provider: value.Provider, Model: value.Model, State: cloneJSON(value.State)}
}
