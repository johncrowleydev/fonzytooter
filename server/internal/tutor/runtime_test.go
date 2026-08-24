package tutor

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

type providerScript struct {
	events  []ProviderEvent
	err     error
	channel <-chan ProviderEvent
}

type scriptedProvider struct {
	mu       sync.Mutex
	scripts  []providerScript
	requests []ModelRequest
}

func (p *scriptedProvider) Stream(_ context.Context, request ModelRequest) (<-chan ProviderEvent, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.requests = append(p.requests, cloneModelRequest(request))
	if len(p.scripts) == 0 {
		return nil, errors.New("unexpected provider request")
	}
	script := p.scripts[0]
	p.scripts = p.scripts[1:]
	if script.err != nil {
		return nil, script.err
	}
	if script.channel != nil {
		return script.channel, nil
	}
	stream := make(chan ProviderEvent, len(script.events))
	for _, event := range script.events {
		stream <- event
	}
	close(stream)
	return stream, nil
}

type staticTurnContextBuilder struct {
	context TurnContext
	err     error
}

func (b staticTurnContextBuilder) Build(context.Context, testAuthUserID, TurnRequest) (TurnContext, error) {
	return b.context, b.err
}

func TestRuntimeDirectResponsePersistsConversationAndEmitsNormalizedEvents(t *testing.T) {
	provider := &scriptedProvider{scripts: []providerScript{{events: []ProviderEvent{
		{Type: ProviderEventTextDelta, Text: "A gradient points uphill."},
		{Type: ProviderEventUsage, Usage: &Usage{InputTokens: 10, OutputTokens: 5, TotalTokens: 15}},
		{Type: ProviderEventCompleted},
	}}}}
	service, store := newRuntimeForTest(t, provider, nil, DefaultMaxModelRounds)
	conversation := createConversationWithMessages(t, store)
	events, err := service.StreamTurn(context.Background(), testUserID, TurnRequest{
		ConversationID: conversation.ID,
		Message:        "What does a gradient mean?",
		PageContext:    &PageContext{Type: "lesson", CourseID: "ai-ml", LessonID: "gradients", LessonTitle: "Client-supplied stale title"},
	})
	if err != nil {
		t.Fatalf("stream turn: %v", err)
	}
	collected := collectTutorEvents(t, events)
	assertEventTypes(t, collected, EventTextDelta, EventUsage, EventCompleted)
	if collected[0].ConversationID != conversation.ID || collected[0].Text != "A gradient points uphill." {
		t.Fatalf("unexpected text event: %#v", collected[0])
	}
	messages, err := store.Messages(context.Background(), testUserID, conversation.ID)
	if err != nil {
		t.Fatalf("load messages: %v", err)
	}
	if len(messages) != 2 || messages[0].Role != MessageRoleUser || messages[1].Role != MessageRoleAssistant || messages[1].Parts[0].Text != "A gradient points uphill." {
		t.Fatalf("unexpected persisted conversation: %#v", messages)
	}
	if strings.Contains(messages[0].Parts[0].Text, "stale title") {
		t.Fatal("page context leaked into persisted user message")
	}
	if len(provider.requests) != 1 || provider.requests[0].MaxOutputTokens <= 0 || provider.requests[0].Reasoning != ReasoningLow {
		t.Fatalf("unexpected provider request: %#v", provider.requests)
	}
	requestText := joinedModelText(provider.requests[0].Messages)
	if !strings.Contains(requestText, "fresh-page") || !strings.Contains(requestText, "deterministic-context") {
		t.Fatalf("fresh and deterministic context missing: %s", requestText)
	}
}

func TestRuntimePersistsContinuationStateAcrossUserTurns(t *testing.T) {
	continuation := &ProviderContinuation{
		Provider: "test-provider", Model: "test-model", State: json.RawMessage(`[{"type":"reasoning.encrypted","data":"opaque-turn-state"}]`),
	}
	provider := &scriptedProvider{scripts: []providerScript{
		{events: []ProviderEvent{
			{Type: ProviderEventState, Continuation: continuation},
			{Type: ProviderEventTextDelta, Text: "First answer."},
			{Type: ProviderEventCompleted},
		}},
		{events: []ProviderEvent{
			{Type: ProviderEventTextDelta, Text: "Second answer."},
			{Type: ProviderEventCompleted},
		}},
	}}
	service, store := newRuntimeForTest(t, provider, nil, DefaultMaxModelRounds)
	conversation := createConversationWithMessages(t, store)
	first, err := service.StreamTurn(context.Background(), testUserID, TurnRequest{ConversationID: conversation.ID, Message: "First question."})
	if err != nil {
		t.Fatalf("start first turn: %v", err)
	}
	collectTutorEvents(t, first)

	persisted, err := store.Messages(context.Background(), testUserID, conversation.ID)
	if err != nil {
		t.Fatalf("load persisted first turn: %v", err)
	}
	if len(persisted) != 2 || persisted[1].Continuation == nil || persisted[1].Continuation.Provider != continuation.Provider || persisted[1].Continuation.Model != continuation.Model || string(persisted[1].Continuation.State) != string(continuation.State) {
		t.Fatalf("continuation state was not persisted: %#v", persisted)
	}

	second, err := service.StreamTurn(context.Background(), testUserID, TurnRequest{ConversationID: conversation.ID, Message: "Second question."})
	if err != nil {
		t.Fatalf("start second turn: %v", err)
	}
	collectTutorEvents(t, second)
	if len(provider.requests) != 2 {
		t.Fatalf("expected two provider requests, got %d", len(provider.requests))
	}
	var replayed *ProviderContinuation
	for _, message := range provider.requests[1].Messages {
		if message.Role == ModelRoleAssistant && modelMessageText(message) == "First answer." {
			replayed = message.Continuation
		}
	}
	if replayed == nil || replayed.Provider != continuation.Provider || replayed.Model != continuation.Model || string(replayed.State) != string(continuation.State) {
		t.Fatalf("persisted continuation state was not replayed: %#v", provider.requests[1])
	}
}

func TestRuntimeCreatesApplicationOwnedConversation(t *testing.T) {
	provider := &scriptedProvider{scripts: []providerScript{{events: []ProviderEvent{
		{Type: ProviderEventTextDelta, Text: "hello"},
		{Type: ProviderEventCompleted},
	}}}}
	service, store := newRuntimeForTest(t, provider, nil, DefaultMaxModelRounds)
	events, err := service.StreamTurn(context.Background(), testUserID, TurnRequest{Message: "start", PageContext: &PageContext{CourseID: "ai-ml"}})
	if err != nil {
		t.Fatalf("stream turn: %v", err)
	}
	collected := collectTutorEvents(t, events)
	conversationID := collected[len(collected)-1].ConversationID
	if conversationID == "" {
		t.Fatal("created conversation ID was not emitted")
	}
	conversation, err := store.Conversation(context.Background(), testUserID, conversationID)
	if err != nil {
		t.Fatalf("load created conversation: %v", err)
	}
	if conversation.CourseID != "ai-ml" {
		t.Fatalf("expected explicit course ownership, got %#v", conversation)
	}
}

func TestRuntimeSingleToolRoundTripPersistsCorrelationAndResult(t *testing.T) {
	tool := mustEchoTool(t, nil)
	provider := &scriptedProvider{scripts: []providerScript{
		{events: []ProviderEvent{
			{Type: ProviderEventState, Continuation: &ProviderContinuation{
				Provider: "test-provider", Model: "test-model", State: json.RawMessage(`[{"type":"reasoning.encrypted","data":"opaque"}]`),
			}},
			{Type: ProviderEventToolCall, ToolCall: &ToolCallRequest{ID: "call-1", Name: "echo", Arguments: json.RawMessage(`{"text":"evidence"}`)}},
			{Type: ProviderEventCompleted},
		}},
		{events: []ProviderEvent{
			{Type: ProviderEventTextDelta, Text: "The evidence is ready."},
			{Type: ProviderEventUsage, Usage: &Usage{InputTokens: 20, OutputTokens: 5, TotalTokens: 25}},
			{Type: ProviderEventCompleted},
		}},
	}}
	service, store := newRuntimeForTest(t, provider, []Tool{tool}, DefaultMaxModelRounds)
	conversation := createConversationWithMessages(t, store)
	events, err := service.StreamTurn(context.Background(), testUserID, TurnRequest{ConversationID: conversation.ID, Message: "Check the evidence."})
	if err != nil {
		t.Fatalf("stream turn: %v", err)
	}
	collected := collectTutorEvents(t, events)
	assertEventTypes(t, collected, EventToolStarted, EventToolCompleted, EventTextDelta, EventUsage, EventCompleted)
	if collected[0].ToolCallID != "call-1" || collected[1].Error != "" {
		t.Fatalf("unexpected tool events: %#v", collected[:2])
	}
	toolCalls, err := store.ToolCalls(context.Background(), testUserID, conversation.ID)
	if err != nil {
		t.Fatalf("load tool calls: %v", err)
	}
	if len(toolCalls) != 1 || toolCalls[0].RequestID != "call-1" || toolCalls[0].Status != ToolCallCompleted || string(toolCalls[0].Result) != `{"echo":"evidence"}` {
		t.Fatalf("unexpected persisted tool call: %#v", toolCalls)
	}
	if len(provider.requests) != 2 {
		t.Fatalf("expected two provider requests, got %d", len(provider.requests))
	}
	continuation := provider.requests[1].Messages
	if len(continuation) < 2 || continuation[len(continuation)-2].ToolCalls[0].ID != "call-1" || continuation[len(continuation)-2].Continuation == nil || string(continuation[len(continuation)-2].Continuation.State) != `[{"type":"reasoning.encrypted","data":"opaque"}]` || continuation[len(continuation)-1].ToolCallID != "call-1" || modelMessageText(continuation[len(continuation)-1]) != `{"echo":"evidence"}` {
		t.Fatalf("tool correlation was not preserved: %#v", continuation)
	}
}

func TestRuntimeRecoversInterruptedPendingCallsBeforeNextTurn(t *testing.T) {
	provider := &scriptedProvider{scripts: []providerScript{{events: []ProviderEvent{
		{Type: ProviderEventTextDelta, Text: "Continuing after the interruption."},
		{Type: ProviderEventCompleted},
	}}}}
	service, store := newRuntimeForTest(t, provider, nil, DefaultMaxModelRounds)
	conversation := createConversationWithMessages(t, store)
	response, err := store.AppendAssistantResponse(context.Background(), testUserID, conversation.ID, nil, []ToolCallInput{{
		RequestID: "interrupted-call",
		Name:      "echo",
		Arguments: json.RawMessage(`{"text":"unfinished"}`),
	}})
	if err != nil {
		t.Fatalf("persist interrupted assistant response: %v", err)
	}

	events, err := service.StreamTurn(context.Background(), testUserID, TurnRequest{ConversationID: conversation.ID, Message: "Continue."})
	if err != nil {
		t.Fatalf("stream next turn: %v", err)
	}
	collected := collectTutorEvents(t, events)
	assertEventTypes(t, collected, EventTextDelta, EventCompleted)

	calls, err := store.ToolCalls(context.Background(), testUserID, conversation.ID)
	if err != nil {
		t.Fatalf("load recovered tool calls: %v", err)
	}
	if len(calls) != 1 || calls[0].ID != response.ToolCalls[0].ID || calls[0].Status != ToolCallFailed || calls[0].Error != InterruptedToolCallError {
		t.Fatalf("unexpected recovered call: %#v", calls)
	}
	if len(provider.requests) != 1 || !strings.Contains(joinedModelText(provider.requests[0].Messages), InterruptedToolCallError) {
		t.Fatalf("recovered failure was not replayed to the provider: %#v", provider.requests)
	}
}

func TestRuntimeSupportsMultipleToolsAndRounds(t *testing.T) {
	var executions int
	tool := mustEchoTool(t, func() { executions++ })
	provider := &scriptedProvider{scripts: []providerScript{
		{events: []ProviderEvent{
			{Type: ProviderEventToolCall, ToolCall: &ToolCallRequest{ID: "call-1", Name: "echo", Arguments: json.RawMessage(`{"text":"one"}`)}},
			{Type: ProviderEventToolCall, ToolCall: &ToolCallRequest{ID: "call-2", Name: "echo", Arguments: json.RawMessage(`{"text":"two"}`)}},
			{Type: ProviderEventCompleted},
		}},
		{events: []ProviderEvent{
			{Type: ProviderEventToolCall, ToolCall: &ToolCallRequest{ID: "call-3", Name: "echo", Arguments: json.RawMessage(`{"text":"three"}`)}},
			{Type: ProviderEventCompleted},
		}},
		{events: []ProviderEvent{{Type: ProviderEventTextDelta, Text: "done"}, {Type: ProviderEventCompleted}}},
	}}
	service, store := newRuntimeForTest(t, provider, []Tool{tool}, DefaultMaxModelRounds)
	conversation := createConversationWithMessages(t, store)
	events, err := service.StreamTurn(context.Background(), testUserID, TurnRequest{ConversationID: conversation.ID, Message: "Use several tools."})
	if err != nil {
		t.Fatalf("stream turn: %v", err)
	}
	collected := collectTutorEvents(t, events)
	if executions != 3 || len(provider.requests) != 3 {
		t.Fatalf("expected three tool executions across three provider calls, got executions=%d requests=%d", executions, len(provider.requests))
	}
	if collected[len(collected)-1].Type != EventCompleted {
		t.Fatalf("expected completion, got %#v", collected)
	}
	toolCalls, err := store.ToolCalls(context.Background(), testUserID, conversation.ID)
	if err != nil || len(toolCalls) != 3 {
		t.Fatalf("expected three persisted tool calls, got %#v, %v", toolCalls, err)
	}
}

func TestRuntimeReturnsToolFailuresToModelAndPersistsThem(t *testing.T) {
	executionFailure := errors.New("tool execution failed")
	failingTool := mustEchoToolWithError(t, executionFailure)
	tests := []struct {
		name      string
		tools     []Tool
		call      ToolCallRequest
		errorPart string
	}{
		{name: "unknown", call: ToolCallRequest{ID: "unknown", Name: "missing", Arguments: json.RawMessage(`{}`)}, errorPart: "unknown tutor tool"},
		{name: "invalid arguments", tools: []Tool{mustEchoTool(t, nil)}, call: ToolCallRequest{ID: "invalid", Name: "echo", Arguments: json.RawMessage(`{"text":""}`)}, errorPart: "invalid tutor tool arguments"},
		{name: "execution error", tools: []Tool{failingTool}, call: ToolCallRequest{ID: "failed", Name: "echo", Arguments: json.RawMessage(`{"text":"ok"}`)}, errorPart: executionFailure.Error()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := &scriptedProvider{scripts: []providerScript{
				{events: []ProviderEvent{{Type: ProviderEventToolCall, ToolCall: &test.call}, {Type: ProviderEventCompleted}}},
				{events: []ProviderEvent{{Type: ProviderEventTextDelta, Text: "recovered"}, {Type: ProviderEventCompleted}}},
			}}
			service, store := newRuntimeForTest(t, provider, test.tools, DefaultMaxModelRounds)
			conversation := createConversationWithMessages(t, store)
			events, err := service.StreamTurn(context.Background(), testUserID, TurnRequest{ConversationID: conversation.ID, Message: "try"})
			if err != nil {
				t.Fatalf("stream turn: %v", err)
			}
			collected := collectTutorEvents(t, events)
			if collected[1].Type != EventToolCompleted || !strings.Contains(collected[1].Error, test.errorPart) || collected[len(collected)-1].Type != EventCompleted {
				t.Fatalf("unexpected failure/recovery events: %#v", collected)
			}
			calls, err := store.ToolCalls(context.Background(), testUserID, conversation.ID)
			if err != nil || len(calls) != 1 || calls[0].Status != ToolCallFailed || !strings.Contains(calls[0].Error, test.errorPart) {
				t.Fatalf("unexpected failed tool persistence: %#v, %v", calls, err)
			}
			if !strings.Contains(modelMessageText(provider.requests[1].Messages[len(provider.requests[1].Messages)-1]), test.errorPart) {
				t.Fatalf("tool error was not returned to model: %#v", provider.requests[1])
			}
		})
	}
}

func TestRuntimeEnforcesMaximumRounds(t *testing.T) {
	var executions int
	tool := mustEchoTool(t, func() { executions++ })
	provider := &scriptedProvider{scripts: []providerScript{
		{events: []ProviderEvent{{Type: ProviderEventToolCall, ToolCall: &ToolCallRequest{ID: "call-1", Name: "echo", Arguments: json.RawMessage(`{"text":"one"}`)}}, {Type: ProviderEventCompleted}}},
		{events: []ProviderEvent{{Type: ProviderEventToolCall, ToolCall: &ToolCallRequest{ID: "call-2", Name: "echo", Arguments: json.RawMessage(`{"text":"two"}`)}}, {Type: ProviderEventCompleted}}},
	}}
	service, store := newRuntimeForTest(t, provider, []Tool{tool}, 2)
	conversation := createConversationWithMessages(t, store)
	events, err := service.StreamTurn(context.Background(), testUserID, TurnRequest{ConversationID: conversation.ID, Message: "loop"})
	if err != nil {
		t.Fatalf("stream turn: %v", err)
	}
	collected := collectTutorEvents(t, events)
	if executions != 1 || collected[len(collected)-1].Type != EventError || collected[len(collected)-1].Error != ErrMaxModelRounds.Error() {
		t.Fatalf("maximum rounds not enforced: executions=%d events=%#v", executions, collected)
	}
}

func TestRuntimeCancellationStopsStream(t *testing.T) {
	providerStream := make(chan ProviderEvent)
	provider := &scriptedProvider{scripts: []providerScript{{channel: providerStream}}}
	service, store := newRuntimeForTest(t, provider, nil, DefaultMaxModelRounds)
	conversation := createConversationWithMessages(t, store)
	ctx, cancel := context.WithCancel(context.Background())
	events, err := service.StreamTurn(ctx, testUserID, TurnRequest{ConversationID: conversation.ID, Message: "wait"})
	if err != nil {
		t.Fatalf("stream turn: %v", err)
	}
	cancel()
	select {
	case _, ok := <-events:
		if ok {
			for range events {
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancelled tutor stream did not close")
	}
	messages, err := store.Messages(context.Background(), testUserID, conversation.ID)
	if err != nil || len(messages) != 1 || messages[0].Role != MessageRoleUser {
		t.Fatalf("unexpected cancellation persistence: %#v, %v", messages, err)
	}
}

func TestRuntimeProviderErrors(t *testing.T) {
	initialError := errors.New("provider unavailable")
	initial := &scriptedProvider{scripts: []providerScript{{err: initialError}}}
	service, store := newRuntimeForTest(t, initial, nil, DefaultMaxModelRounds)
	conversation := createConversationWithMessages(t, store)
	if _, err := service.StreamTurn(context.Background(), testUserID, TurnRequest{ConversationID: conversation.ID, Message: "hello"}); !errors.Is(err, initialError) {
		t.Fatalf("expected initial provider error, got %v", err)
	}

	tool := mustEchoTool(t, nil)
	later := &scriptedProvider{scripts: []providerScript{
		{events: []ProviderEvent{{Type: ProviderEventToolCall, ToolCall: &ToolCallRequest{ID: "call", Name: "echo", Arguments: json.RawMessage(`{"text":"ok"}`)}}, {Type: ProviderEventCompleted}}},
		{err: initialError},
	}}
	service, store = newRuntimeForTest(t, later, []Tool{tool}, DefaultMaxModelRounds)
	conversation = createConversationWithMessages(t, store)
	events, err := service.StreamTurn(context.Background(), testUserID, TurnRequest{ConversationID: conversation.ID, Message: "hello"})
	if err != nil {
		t.Fatalf("start tool turn: %v", err)
	}
	collected := collectTutorEvents(t, events)
	if collected[len(collected)-1].Type != EventError || !strings.Contains(collected[len(collected)-1].Error, initialError.Error()) {
		t.Fatalf("expected streamed later provider error, got %#v", collected)
	}
}

func TestRuntimeRejectsMalformedProviderStreams(t *testing.T) {
	tests := []struct {
		name    string
		events  []ProviderEvent
		message string
	}{
		{
			name:    "missing completion",
			events:  []ProviderEvent{{Type: ProviderEventTextDelta, Text: "partial"}},
			message: "ended before completion",
		},
		{
			name: "event after completion",
			events: []ProviderEvent{
				{Type: ProviderEventCompleted},
				{Type: ProviderEventTextDelta, Text: "late"},
			},
			message: "event after completion",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := &scriptedProvider{scripts: []providerScript{{events: test.events}}}
			service, store := newRuntimeForTest(t, provider, nil, DefaultMaxModelRounds)
			conversation := createConversationWithMessages(t, store)
			events, err := service.StreamTurn(context.Background(), testUserID, TurnRequest{ConversationID: conversation.ID, Message: "hello"})
			if err != nil {
				t.Fatalf("start turn: %v", err)
			}
			collected := collectTutorEvents(t, events)
			if len(collected) == 0 || collected[len(collected)-1].Type != EventError || !strings.Contains(collected[len(collected)-1].Error, test.message) {
				t.Fatalf("expected terminal stream error containing %q, got %#v", test.message, collected)
			}
		})
	}
}

func newRuntimeForTest(t *testing.T, provider Provider, tools []Tool, maxRounds int) (*Service, *ConversationStore) {
	t.Helper()
	store, _ := newTestConversationStore(t)
	registry, err := NewToolRegistry(tools...)
	if err != nil {
		t.Fatalf("create tool registry: %v", err)
	}
	config := DefaultContextManagerConfig()
	config.ContextWindowTokens = 100_000
	config.OutputReserveTokens = 2_000
	config.ToolReserveTokens = 1_000
	config.CompactionTriggerTokens = 90_000
	manager, err := NewContextManager(store, ConservativeTokenEstimator{}, RuleBasedCompactor{}, config)
	if err != nil {
		t.Fatalf("create context manager: %v", err)
	}
	service, err := NewRuntimeService(RuntimeConfig{
		Provider:       provider,
		Store:          store,
		Tools:          registry,
		ContextManager: manager,
		ContextBuilder: staticTurnContextBuilder{context: TurnContext{
			SystemPolicy:         "test-policy",
			CurrentPageContext:   "fresh-page",
			DeterministicContext: "deterministic-context",
			Reasoning:            ReasoningLow,
		}},
		MaxModelRounds: maxRounds,
		CostGate:       testCostGate(t, CostGateConfig{Entitled: true, MonthlyTurnLimit: 100}),
	})
	if err != nil {
		t.Fatalf("create runtime: %v", err)
	}
	return service, store
}

func mustEchoTool(t *testing.T, onExecute func()) Tool {
	t.Helper()
	tool, err := NewTypedTool[echoToolArguments, echoToolResult]("echo", "Echo text.", func(arguments echoToolArguments) error {
		if strings.TrimSpace(arguments.Text) == "" {
			return errors.New("text must not be blank")
		}
		return nil
	}, func(_ context.Context, _ testAuthUserID, arguments echoToolArguments) (echoToolResult, error) {
		if onExecute != nil {
			onExecute()
		}
		return echoToolResult{Echo: arguments.Text}, nil
	})
	if err != nil {
		t.Fatalf("create echo tool: %v", err)
	}
	return tool
}

func mustEchoToolWithError(t *testing.T, executionError error) Tool {
	t.Helper()
	tool, err := NewTypedTool[echoToolArguments, echoToolResult]("echo", "Echo text.", nil, func(context.Context, testAuthUserID, echoToolArguments) (echoToolResult, error) {
		return echoToolResult{}, executionError
	})
	if err != nil {
		t.Fatalf("create failing tool: %v", err)
	}
	return tool
}

func collectTutorEvents(t *testing.T, events <-chan Event) []Event {
	t.Helper()
	collected := make([]Event, 0)
	for event := range events {
		collected = append(collected, event)
	}
	return collected
}

func assertEventTypes(t *testing.T, events []Event, expected ...EventType) {
	t.Helper()
	if len(events) != len(expected) {
		t.Fatalf("expected %d events, got %#v", len(expected), events)
	}
	for index, eventType := range expected {
		if events[index].Type != eventType {
			t.Fatalf("expected event %d to be %s, got %#v", index, eventType, events[index])
		}
	}
}

func cloneModelRequest(request ModelRequest) ModelRequest {
	request.Messages = cloneModelMessages(request.Messages)
	request.Tools = append([]ToolDefinition(nil), request.Tools...)
	for index := range request.Tools {
		request.Tools[index].InputSchema = cloneJSON(request.Tools[index].InputSchema)
	}
	return request
}
