package tutor

import (
	"context"
	"encoding/json"
)

type EventType string

const (
	EventTextDelta     EventType = "text_delta"
	EventToolStarted   EventType = "tool_started"
	EventToolCompleted EventType = "tool_completed"
	EventCitation      EventType = "citation"
	EventUsage         EventType = "usage"
	EventCompleted     EventType = "completed"
	EventError         EventType = "error"
)

type Event struct {
	Type     EventType `json:"type"`
	Text     string    `json:"text,omitempty"`
	Tool     string    `json:"tool,omitempty"`
	SourceID string    `json:"sourceId,omitempty"`
	Error    string    `json:"error,omitempty"`
}

type TurnRequest struct {
	ConversationID string          `json:"conversationId,omitempty"`
	Message        string          `json:"message"`
	Mode           string          `json:"mode,omitempty"`
	PageContext    json.RawMessage `json:"pageContext,omitempty"`
}

type Provider interface {
	StreamTurn(ctx context.Context, request TurnRequest) (<-chan Event, error)
}

type Service struct {
	provider Provider
}

func NewService(provider Provider) *Service {
	return &Service{provider: provider}
}

func (s *Service) StreamTurn(ctx context.Context, request TurnRequest) (<-chan Event, error) {
	return s.provider.StreamTurn(ctx, request)
}
