package tutor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/johncrowleydev/fonzytooter/server/internal/auth"
)

const (
	DefaultMaxModelRounds = 4
	DefaultSystemPolicy   = "You are Fonzytooter's bounded technical learning tutor. Help the learner understand the current question, use available tools only when needed, and do not claim that conversation alone proves mastery."
)

var ErrMaxModelRounds = errors.New("tutor maximum model rounds exceeded")

type TurnContext struct {
	SystemPolicy         string
	CurrentPageContext   string
	DeterministicContext string
	AllowedTools         []string
	Reasoning            ReasoningPolicy
}

type TurnContextBuilder interface {
	Build(ctx context.Context, userID auth.UserID, request TurnRequest) (TurnContext, error)
}

type BasicTurnContextBuilder struct {
	SystemPolicy string
}

func (b BasicTurnContextBuilder) Build(_ context.Context, _ auth.UserID, request TurnRequest) (TurnContext, error) {
	policy := b.SystemPolicy
	if strings.TrimSpace(policy) == "" {
		policy = DefaultSystemPolicy
	}
	var pageContext string
	if request.PageContext != nil {
		encoded, err := json.Marshal(request.PageContext)
		if err != nil {
			return TurnContext{}, fmt.Errorf("encode current tutor page context: %w", err)
		}
		pageContext = string(encoded)
	}
	return TurnContext{SystemPolicy: policy, CurrentPageContext: pageContext, Reasoning: ReasoningLow}, nil
}

type RuntimeConfig struct {
	Provider       Provider
	Store          *ConversationStore
	Tools          *ToolRegistry
	ContextManager *ContextManager
	ContextBuilder TurnContextBuilder
	MaxModelRounds int
	CostGate       *CostGate
}

type Service struct {
	provider          Provider
	conversations     *ConversationStore
	tools             *ToolRegistry
	contextManager    *ContextManager
	contextBuilder    TurnContextBuilder
	maxModelRounds    int
	costGate          *CostGate
	conversationLocks sync.Map
}

func NewService(provider Provider, costGates ...*CostGate) *Service {
	if provider == nil {
		panic("tutor.NewService: nil provider")
	}
	if len(costGates) > 1 {
		panic("tutor.NewService: multiple cost gates")
	}
	var costGate *CostGate
	if len(costGates) == 1 {
		costGate = costGates[0]
	}
	tools, _ := NewToolRegistry()
	return &Service{
		provider:       provider,
		tools:          tools,
		contextBuilder: BasicTurnContextBuilder{},
		maxModelRounds: DefaultMaxModelRounds,
		costGate:       costGate,
	}
}

func NewPersistentService(provider Provider, conversations *ConversationStore) *Service {
	if conversations == nil {
		panic("tutor.NewPersistentService: nil conversation store")
	}
	tools, _ := NewToolRegistry()
	manager, err := NewContextManager(
		conversations,
		ConservativeTokenEstimator{},
		ModelCompactor{Provider: provider, Fallback: RuleBasedCompactor{}},
		DefaultContextManagerConfig(),
	)
	if err != nil {
		panic(fmt.Sprintf("tutor.NewPersistentService: %v", err))
	}
	service, err := NewRuntimeService(RuntimeConfig{
		Provider:       provider,
		Store:          conversations,
		Tools:          tools,
		ContextManager: manager,
		ContextBuilder: BasicTurnContextBuilder{},
		MaxModelRounds: DefaultMaxModelRounds,
	})
	if err != nil {
		panic(fmt.Sprintf("tutor.NewPersistentService: %v", err))
	}
	return service
}

func NewRuntimeService(config RuntimeConfig) (*Service, error) {
	if config.Provider == nil {
		return nil, errors.New("tutor runtime provider is nil")
	}
	if config.Store == nil {
		return nil, errors.New("tutor runtime conversation store is nil")
	}
	if config.Tools == nil {
		return nil, errors.New("tutor runtime tool registry is nil")
	}
	if config.ContextManager == nil {
		return nil, errors.New("tutor runtime context manager is nil")
	}
	if config.ContextBuilder == nil {
		return nil, errors.New("tutor runtime context builder is nil")
	}
	if config.MaxModelRounds <= 0 {
		return nil, errors.New("tutor runtime maximum model rounds must be positive")
	}
	return &Service{
		provider:       config.Provider,
		conversations:  config.Store,
		tools:          config.Tools,
		contextManager: config.ContextManager,
		contextBuilder: config.ContextBuilder,
		maxModelRounds: config.MaxModelRounds,
		costGate:       config.CostGate,
	}, nil
}

func (s *Service) StreamTurn(ctx context.Context, userID auth.UserID, request TurnRequest) (<-chan Event, error) {
	if s.costGate != nil {
		if _, err := s.costGate.ReserveTurn(ctx, userID); err != nil {
			return nil, err
		}
	}
	if s.conversations == nil {
		return s.streamUnpersisted(ctx, userID, request)
	}
	turnContext, err := s.contextBuilder.Build(ctx, userID, request)
	if err != nil {
		return nil, err
	}
	conversationID, err := s.resolveConversation(ctx, userID, request)
	if err != nil {
		return nil, err
	}
	unlock, err := s.lockConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	releaseLock := true
	defer func() {
		if releaseLock {
			unlock()
		}
	}()
	if _, err := s.conversations.RecoverPendingToolCalls(ctx, userID, conversationID); err != nil {
		return nil, err
	}
	if _, err := s.conversations.AppendTextMessage(ctx, userID, conversationID, MessageRoleUser, request.Message); err != nil {
		return nil, err
	}
	definitions, err := s.tools.Definitions(turnContext.AllowedTools)
	if err != nil {
		return nil, err
	}
	prepared, err := s.contextManager.Prepare(ctx, userID, conversationID, ContextInput{
		SystemPolicy:         turnContext.SystemPolicy,
		CurrentPageContext:   turnContext.CurrentPageContext,
		DeterministicContext: turnContext.DeterministicContext,
	}, definitions)
	if err != nil {
		return nil, err
	}
	reasoning := turnContext.Reasoning
	if reasoning == "" {
		reasoning = ReasoningLow
	}
	if !validReasoningPolicy(reasoning) {
		return nil, fmt.Errorf("invalid tutor reasoning policy %q", reasoning)
	}
	modelRequest := ModelRequest{
		Messages:        prepared.Messages,
		Tools:           definitions,
		Reasoning:       reasoning,
		MaxOutputTokens: prepared.MaxOutputTokens,
	}
	providerEvents, err := s.provider.Stream(ctx, modelRequest)
	if err != nil {
		return nil, err
	}
	events := make(chan Event, 16)
	releaseLock = false
	go s.run(ctx, events, userID, conversationID, turnContext.AllowedTools, modelRequest, providerEvents, unlock)
	return events, nil
}

func (s *Service) TutorAccess(ctx context.Context, userID auth.UserID) (Access, error) {
	if s.costGate == nil {
		return Access{Status: AccessUnavailable}, nil
	}
	return s.costGate.Access(ctx, userID)
}

func (s *Service) lockConversation(ctx context.Context, conversationID string) (func(), error) {
	created := make(chan struct{}, 1)
	created <- struct{}{}
	value, _ := s.conversationLocks.LoadOrStore(conversationID, created)
	lock := value.(chan struct{})
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-lock:
		return func() { lock <- struct{}{} }, nil
	}
}

func (s *Service) streamUnpersisted(ctx context.Context, userID auth.UserID, request TurnRequest) (<-chan Event, error) {
	turnContext, err := s.contextBuilder.Build(ctx, userID, request)
	if err != nil {
		return nil, err
	}
	modelRequest := ModelRequest{
		Messages: []ModelMessage{
			textModelMessage(ModelRoleSystem, turnContext.SystemPolicy),
			textModelMessage(ModelRoleUser, request.Message),
		},
		Reasoning:       ReasoningLow,
		MaxOutputTokens: DefaultContextManagerConfig().OutputReserveTokens,
	}
	providerEvents, err := s.provider.Stream(ctx, modelRequest)
	if err != nil {
		return nil, err
	}
	events := make(chan Event, 8)
	go func() {
		defer close(events)
		_, _, _, err := consumeProviderEvents(ctx, events, request.ConversationID, providerEvents)
		if err != nil {
			sendTutorEvent(ctx, events, Event{Type: EventError, ConversationID: request.ConversationID, Error: err.Error()})
			return
		}
		sendTutorEvent(ctx, events, Event{Type: EventCompleted, ConversationID: request.ConversationID})
	}()
	return events, nil
}

func (s *Service) resolveConversation(ctx context.Context, userID auth.UserID, request TurnRequest) (string, error) {
	if request.ConversationID == "" {
		courseID := ""
		if request.PageContext != nil {
			courseID = request.PageContext.CourseID
		}
		conversation, err := s.conversations.CreateConversation(ctx, userID, CreateConversationParams{CourseID: courseID})
		if err != nil {
			return "", err
		}
		return conversation.ID, nil
	}
	conversation, err := s.conversations.Conversation(ctx, userID, request.ConversationID)
	if err != nil {
		return "", err
	}
	if request.PageContext != nil && conversation.CourseID != "" && request.PageContext.CourseID != "" && conversation.CourseID != request.PageContext.CourseID {
		return "", errors.New("tutor conversation belongs to a different course")
	}
	return conversation.ID, nil
}

func (s *Service) run(
	ctx context.Context,
	events chan<- Event,
	userID auth.UserID,
	conversationID string,
	allowedTools []string,
	request ModelRequest,
	providerEvents <-chan ProviderEvent,
	unlock func(),
) {
	defer close(events)
	defer unlock()
	for round := 1; round <= s.maxModelRounds; round++ {
		text, toolCalls, continuation, err := consumeProviderEvents(ctx, events, conversationID, providerEvents)
		if err != nil {
			if ctx.Err() == nil {
				sendTutorEvent(ctx, events, Event{Type: EventError, ConversationID: conversationID, Error: err.Error()})
			}
			return
		}
		if len(toolCalls) == 0 {
			if strings.TrimSpace(text) == "" {
				sendTutorEvent(ctx, events, Event{Type: EventError, ConversationID: conversationID, Error: "tutor provider completed without a response"})
				return
			}
			if _, err := s.conversations.AppendAssistantResponseWithContinuation(
				ctx, userID, conversationID, []ContentPart{{Kind: ContentKindText, Text: text}}, nil, continuation,
			); err != nil {
				sendTutorEvent(ctx, events, Event{Type: EventError, ConversationID: conversationID, Error: err.Error()})
				return
			}
			sendTutorEvent(ctx, events, Event{Type: EventCompleted, ConversationID: conversationID})
			return
		}
		if round == s.maxModelRounds {
			sendTutorEvent(ctx, events, Event{Type: EventError, ConversationID: conversationID, Error: ErrMaxModelRounds.Error()})
			return
		}

		parts := make([]ContentPart, 0, 1)
		if strings.TrimSpace(text) != "" {
			parts = append(parts, ContentPart{Kind: ContentKindText, Text: text})
		}
		callInputs := make([]ToolCallInput, len(toolCalls))
		for index, call := range toolCalls {
			callInputs[index] = ToolCallInput{RequestID: call.ID, Name: call.Name, Arguments: call.Arguments}
		}
		assistant, err := s.conversations.AppendAssistantResponseWithContinuation(ctx, userID, conversationID, parts, callInputs, continuation)
		if err != nil {
			sendTutorEvent(ctx, events, Event{Type: EventError, ConversationID: conversationID, Error: err.Error()})
			return
		}
		assistantModel := ModelMessage{
			Role: ModelRoleAssistant, ToolCalls: cloneToolCallRequests(toolCalls), Continuation: cloneProviderContinuation(continuation),
		}
		if text != "" {
			assistantModel.Parts = []ModelContentPart{{Kind: ModelContentText, Text: text}}
		}
		request.Messages = append(request.Messages, assistantModel)
		for index, call := range toolCalls {
			stored := assistant.ToolCalls[index]
			if !sendTutorEvent(ctx, events, Event{Type: EventToolStarted, ConversationID: conversationID, Tool: call.Name, ToolCallID: call.ID}) {
				return
			}
			result, toolErr := s.tools.Execute(ctx, userID, call.Name, call.Arguments, allowedTools)
			resultForModel := result
			if toolErr != nil {
				resultForModel, _ = json.Marshal(map[string]string{"error": toolErr.Error()})
				if _, err := s.conversations.CompleteToolCall(ctx, userID, stored.ID, nil, toolErr.Error()); err != nil {
					sendTutorEvent(ctx, events, Event{Type: EventError, ConversationID: conversationID, Error: err.Error()})
					return
				}
			} else if _, err := s.conversations.CompleteToolCall(ctx, userID, stored.ID, result, ""); err != nil {
				sendTutorEvent(ctx, events, Event{Type: EventError, ConversationID: conversationID, Error: err.Error()})
				return
			}
			if !sendTutorEvent(ctx, events, Event{Type: EventToolCompleted, ConversationID: conversationID, Tool: call.Name, ToolCallID: call.ID, Error: errorText(toolErr)}) {
				return
			}
			if toolErr == nil {
				for _, sourceID := range s.tools.SourceIDs(call.Name, result) {
					if !sendTutorEvent(ctx, events, Event{Type: EventCitation, ConversationID: conversationID, Tool: call.Name, ToolCallID: call.ID, SourceID: sourceID}) {
						return
					}
				}
			}
			request.Messages = append(request.Messages, ModelMessage{
				Role:       ModelRoleTool,
				ToolCallID: call.ID,
				ToolName:   call.Name,
				Parts:      []ModelContentPart{{Kind: ModelContentText, Text: string(resultForModel)}},
			})
		}
		if err := s.contextManager.CheckRequest(request.Messages, request.Tools); err != nil {
			sendTutorEvent(ctx, events, Event{Type: EventError, ConversationID: conversationID, Error: err.Error()})
			return
		}
		providerEvents, err = s.provider.Stream(ctx, request)
		if err != nil {
			sendTutorEvent(ctx, events, Event{Type: EventError, ConversationID: conversationID, Error: err.Error()})
			return
		}
	}
}

func consumeProviderEvents(ctx context.Context, output chan<- Event, conversationID string, input <-chan ProviderEvent) (string, []ToolCallRequest, *ProviderContinuation, error) {
	var text strings.Builder
	toolCalls := make([]ToolCallRequest, 0)
	var continuation *ProviderContinuation
	completed := false
	for {
		select {
		case <-ctx.Done():
			return "", nil, nil, ctx.Err()
		case event, ok := <-input:
			if !ok {
				if !completed {
					return "", nil, nil, errors.New("tutor provider stream ended before completion")
				}
				return text.String(), toolCalls, continuation, nil
			}
			if completed {
				return "", nil, nil, errors.New("tutor provider emitted an event after completion")
			}
			switch event.Type {
			case ProviderEventTextDelta:
				text.WriteString(event.Text)
				if !sendTutorEvent(ctx, output, Event{Type: EventTextDelta, ConversationID: conversationID, Text: event.Text}) {
					return "", nil, nil, ctx.Err()
				}
			case ProviderEventToolCall:
				if event.ToolCall == nil || strings.TrimSpace(event.ToolCall.ID) == "" || strings.TrimSpace(event.ToolCall.Name) == "" || !json.Valid(event.ToolCall.Arguments) {
					return "", nil, nil, errors.New("tutor provider returned an invalid tool call")
				}
				toolCalls = append(toolCalls, ToolCallRequest{ID: event.ToolCall.ID, Name: event.ToolCall.Name, Arguments: cloneJSON(event.ToolCall.Arguments)})
			case ProviderEventState:
				if continuation != nil || event.Continuation == nil || strings.TrimSpace(event.Continuation.Provider) == "" || strings.TrimSpace(event.Continuation.Model) == "" || !json.Valid(event.Continuation.State) {
					return "", nil, nil, errors.New("tutor provider returned invalid continuation state")
				}
				continuation = cloneProviderContinuation(event.Continuation)
			case ProviderEventUsage:
				if event.Usage != nil && !sendTutorEvent(ctx, output, Event{Type: EventUsage, ConversationID: conversationID, Usage: cloneUsage(event.Usage)}) {
					return "", nil, nil, ctx.Err()
				}
			case ProviderEventCompleted:
				completed = true
			case ProviderEventError:
				return "", nil, nil, errors.New(event.Error)
			default:
				return "", nil, nil, fmt.Errorf("unknown tutor provider event %q", event.Type)
			}
		}
	}
}

func sendTutorEvent(ctx context.Context, output chan<- Event, event Event) bool {
	select {
	case <-ctx.Done():
		return false
	case output <- event:
		return true
	}
}

func cloneToolCallRequests(calls []ToolCallRequest) []ToolCallRequest {
	result := make([]ToolCallRequest, len(calls))
	for index, call := range calls {
		result[index] = call
		result[index].Arguments = cloneJSON(call.Arguments)
	}
	return result
}

func cloneUsage(usage *Usage) *Usage {
	if usage == nil {
		return nil
	}
	copy := *usage
	return &copy
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func validReasoningPolicy(policy ReasoningPolicy) bool {
	return policy == ReasoningNone || policy == ReasoningLow || policy == ReasoningMedium || policy == ReasoningHigh
}
