package tutor

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type characterEstimator struct{}

func (characterEstimator) EstimateText(text string) int { return len(text) }

type capturingCompactor struct {
	requests []CompactionRequest
	memory   ConversationMemory
}

func (c *capturingCompactor) Compact(_ context.Context, request CompactionRequest) (ConversationMemory, error) {
	c.requests = append(c.requests, request)
	memory := c.memory
	if memory.Summary == "" {
		memory = request.Previous
		memory.Summary = strings.TrimSpace(memory.Summary + " compacted")
	}
	if memory.Structured.LearnerGoal == "" {
		memory.Structured = request.Previous.Structured
	}
	return memory, nil
}

func TestContextManagerCompactsOldHistoryAndKeepsRecentTailWithoutDuplication(t *testing.T) {
	store, _ := newTestConversationStore(t)
	conversation := createConversationWithMessages(t, store,
		messageFixture{MessageRoleUser, "old question one"},
		messageFixture{MessageRoleAssistant, "old answer one"},
		messageFixture{MessageRoleUser, "old question two"},
		messageFixture{MessageRoleAssistant, "recent answer"},
		messageFixture{MessageRoleUser, "current question"},
	)
	compactor := &capturingCompactor{memory: ConversationMemory{
		Summary: "The learner is comparing function composition explanations.",
		Structured: StructuredMemory{
			LearnerGoal:            "Understand composition",
			Misconceptions:         []string{"Composition is commutative"},
			UnsuccessfulApproaches: []string{"Symbolic notation alone"},
			UnresolvedQuestions:    []string{"Why does order matter?"},
			SourceIDs:              []string{"src.functions"},
		},
	}}
	config := DefaultContextManagerConfig()
	config.RecentMessageCount = 2
	manager, err := NewContextManager(store, ConservativeTokenEstimator{}, compactor, config)
	if err != nil {
		t.Fatalf("create context manager: %v", err)
	}
	prepared, err := manager.Prepare(context.Background(), conversation.ID, ContextInput{
		SystemPolicy:       "Always retain this policy.",
		CurrentPageContext: `{"lessonId":"fresh-page"}`,
	}, nil)
	if err != nil {
		t.Fatalf("prepare context: %v", err)
	}
	if len(compactor.requests) != 1 {
		t.Fatalf("expected one compaction, got %d", len(compactor.requests))
	}
	if len(compactor.requests[0].Messages) != 3 || compactor.requests[0].Messages[2].Sequence != 3 {
		t.Fatalf("unexpected compacted range: %#v", compactor.requests[0].Messages)
	}
	for _, compacted := range compactor.requests[0].Messages {
		if strings.Contains(modelMessageText(compacted.Message), "fresh-page") {
			t.Fatal("fresh page context leaked into durable compaction input")
		}
	}
	joined := joinedModelText(prepared.Messages)
	if strings.Contains(joined, "old question one") || strings.Count(joined, "recent answer") != 1 || strings.Count(joined, "current question") != 1 {
		t.Fatalf("history was duplicated or retained outside the compacted memory: %s", joined)
	}
	for _, required := range []string{"Always retain this policy.", "fresh-page", "Understand composition", "Composition is commutative", "src.functions"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("prepared context missing %q: %s", required, joined)
		}
	}
	if last := prepared.Messages[len(prepared.Messages)-1]; last.Role != ModelRoleUser || modelMessageText(last) != "current question" {
		t.Fatalf("current user message was not retained last: %#v", last)
	}
	memory, err := store.ConversationMemory(context.Background(), conversation.ID)
	if err != nil {
		t.Fatalf("load memory: %v", err)
	}
	if memory.SummarizedThroughSequence != 3 {
		t.Fatalf("expected summarized-through marker 3, got %#v", memory)
	}
}

func TestContextManagerRepeatedCompactionAdvancesMarkerAndPreservesMemory(t *testing.T) {
	store, _ := newTestConversationStore(t)
	conversation := createConversationWithMessages(t, store,
		messageFixture{MessageRoleUser, "q1"},
		messageFixture{MessageRoleAssistant, "a1"},
		messageFixture{MessageRoleUser, "q2"},
		messageFixture{MessageRoleAssistant, "a2"},
		messageFixture{MessageRoleUser, "q3"},
	)
	compactor := &capturingCompactor{memory: ConversationMemory{Summary: "salient", Structured: StructuredMemory{LearnerGoal: "goal", Misconceptions: []string{"m1"}}}}
	config := DefaultContextManagerConfig()
	config.RecentMessageCount = 2
	manager, err := NewContextManager(store, ConservativeTokenEstimator{}, compactor, config)
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	if _, err := manager.Prepare(context.Background(), conversation.ID, ContextInput{SystemPolicy: "policy"}, nil); err != nil {
		t.Fatalf("first prepare: %v", err)
	}
	if _, err := store.AppendTextMessage(context.Background(), conversation.ID, MessageRoleAssistant, "a3"); err != nil {
		t.Fatalf("append assistant: %v", err)
	}
	if _, err := store.AppendTextMessage(context.Background(), conversation.ID, MessageRoleUser, "q4"); err != nil {
		t.Fatalf("append user: %v", err)
	}
	compactor.memory = ConversationMemory{}
	prepared, err := manager.Prepare(context.Background(), conversation.ID, ContextInput{SystemPolicy: "policy"}, nil)
	if err != nil {
		t.Fatalf("second prepare: %v", err)
	}
	if len(compactor.requests) != 2 || compactor.requests[1].Previous.SummarizedThroughSequence != 3 {
		t.Fatalf("second compaction did not start from prior marker: %#v", compactor.requests)
	}
	memory, err := store.ConversationMemory(context.Background(), conversation.ID)
	if err != nil {
		t.Fatalf("load memory: %v", err)
	}
	if memory.SummarizedThroughSequence != 5 || !strings.Contains(memory.Summary, "salient") || len(memory.Structured.Misconceptions) != 1 {
		t.Fatalf("salient memory or marker was lost: %#v", memory)
	}
	if strings.Count(joinedModelText(prepared.Messages), "q4") != 1 {
		t.Fatalf("current message duplicated after repeated compaction: %#v", prepared.Messages)
	}
}

func TestContextManagerCompactsBeforeHardBudgetAndReservesHeadroom(t *testing.T) {
	store, _ := newTestConversationStore(t)
	conversation := createConversationWithMessages(t, store,
		messageFixture{MessageRoleUser, strings.Repeat("q", 25)},
		messageFixture{MessageRoleAssistant, strings.Repeat("a", 25)},
		messageFixture{MessageRoleUser, strings.Repeat("c", 25)},
	)
	compactor := &capturingCompactor{memory: ConversationMemory{Summary: "short memory"}}
	config := ContextManagerConfig{
		ContextWindowTokens:     260,
		OutputReserveTokens:     40,
		ToolReserveTokens:       20,
		CompactionTriggerTokens: 100,
		RecentMessageCount:      10,
		MaxMemoryCharacters:     30,
		MaxMemoryItems:          4,
		MaxMemoryItemCharacters: 20,
	}
	manager, err := NewContextManager(store, characterEstimator{}, compactor, config)
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	prepared, err := manager.Prepare(context.Background(), conversation.ID, ContextInput{SystemPolicy: "policy"}, nil)
	if err != nil {
		t.Fatalf("prepare context: %v", err)
	}
	if len(compactor.requests) != 1 {
		t.Fatalf("expected pre-limit compaction, got %d", len(compactor.requests))
	}
	if prepared.InputBudget != 200 || prepared.MaxOutputTokens != 40 || prepared.InputTokens > prepared.InputBudget {
		t.Fatalf("headroom or budget not enforced: %#v", prepared)
	}
}

func TestContextManagerProgressivelyShrinksRecentTailUntilItFits(t *testing.T) {
	store, _ := newTestConversationStore(t)
	conversation := createConversationWithMessages(t, store,
		messageFixture{MessageRoleUser, strings.Repeat("q", 60)},
		messageFixture{MessageRoleAssistant, strings.Repeat("a", 60)},
		messageFixture{MessageRoleUser, strings.Repeat("q", 60)},
		messageFixture{MessageRoleAssistant, strings.Repeat("a", 60)},
		messageFixture{MessageRoleUser, strings.Repeat("q", 60)},
		messageFixture{MessageRoleAssistant, strings.Repeat("a", 60)},
		messageFixture{MessageRoleUser, "current question"},
	)
	compactor := &capturingCompactor{memory: ConversationMemory{Summary: "bounded semantic memory"}}
	config := ContextManagerConfig{
		ContextWindowTokens: 280, OutputReserveTokens: 40, ToolReserveTokens: 20,
		CompactionTriggerTokens: 160, RecentMessageCount: 5,
		MaxMemoryCharacters: 40, MaxMemoryItems: 4, MaxMemoryItemCharacters: 20,
	}
	manager, err := NewContextManager(store, characterEstimator{}, compactor, config)
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	prepared, err := manager.Prepare(context.Background(), conversation.ID, ContextInput{SystemPolicy: "policy"}, nil)
	if err != nil {
		t.Fatalf("prepare progressively compacted context: %v", err)
	}
	if len(compactor.requests) < 2 {
		t.Fatalf("expected multiple compaction passes, got %d", len(compactor.requests))
	}
	if prepared.InputTokens > prepared.InputBudget {
		t.Fatalf("prepared context still exceeds its budget: %#v", prepared)
	}
	if last := prepared.Messages[len(prepared.Messages)-1]; last.Role != ModelRoleUser || modelMessageText(last) != "current question" {
		t.Fatalf("current user message was not retained verbatim: %#v", last)
	}
}

func TestContextManagerRejectsUnfitCurrentTurnAndToolDefinitions(t *testing.T) {
	store, _ := newTestConversationStore(t)
	conversation := createConversationWithMessages(t, store, messageFixture{MessageRoleUser, strings.Repeat("x", 200)})
	config := ContextManagerConfig{
		ContextWindowTokens: 100, OutputReserveTokens: 20, ToolReserveTokens: 10,
		CompactionTriggerTokens: 60, RecentMessageCount: 2,
		MaxMemoryCharacters: 30, MaxMemoryItems: 4, MaxMemoryItemCharacters: 20,
	}
	manager, err := NewContextManager(store, characterEstimator{}, &capturingCompactor{}, config)
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	if _, err := manager.Prepare(context.Background(), conversation.ID, ContextInput{SystemPolicy: "policy"}, nil); !errors.Is(err, ErrContextBudgetExceeded) {
		t.Fatalf("expected current-turn budget error, got %v", err)
	}

	second := createConversationWithMessages(t, store, messageFixture{MessageRoleUser, "short"})
	largeTool := ToolDefinition{Name: "large", Description: strings.Repeat("d", 80), InputSchema: json.RawMessage(`{"type":"object"}`)}
	if _, err := manager.Prepare(context.Background(), second.ID, ContextInput{SystemPolicy: "policy"}, []ToolDefinition{largeTool}); !errors.Is(err, ErrContextBudgetExceeded) {
		t.Fatalf("expected tool-definition budget error, got %v", err)
	}
}

type messageFixture struct {
	role MessageRole
	text string
}

func createConversationWithMessages(t *testing.T, store *ConversationStore, fixtures ...messageFixture) Conversation {
	t.Helper()
	conversation, err := store.CreateConversation(context.Background(), CreateConversationParams{CourseID: "ai-ml"})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	for _, fixture := range fixtures {
		if _, err := store.AppendTextMessage(context.Background(), conversation.ID, fixture.role, fixture.text); err != nil {
			t.Fatalf("append %s message: %v", fixture.role, err)
		}
	}
	return conversation
}

func joinedModelText(messages []ModelMessage) string {
	parts := make([]string, 0, len(messages))
	for _, message := range messages {
		parts = append(parts, modelMessageText(message))
	}
	return strings.Join(parts, "\n")
}
