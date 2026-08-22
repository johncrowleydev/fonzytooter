package tutor

import (
	"context"
	"encoding/json"
)

type ModelRole string

const (
	ModelRoleSystem    ModelRole = "system"
	ModelRoleUser      ModelRole = "user"
	ModelRoleAssistant ModelRole = "assistant"
	ModelRoleTool      ModelRole = "tool"
)

type ModelContentKind string

const (
	ModelContentText     ModelContentKind = "text"
	ModelContentImage    ModelContentKind = "image"
	ModelContentDocument ModelContentKind = "document"
)

type ModelContentPart struct {
	Kind      ModelContentKind
	Text      string
	URL       string
	MediaType string
}

type ModelMessage struct {
	Role       ModelRole
	Parts      []ModelContentPart
	ToolCalls  []ToolCallRequest
	ToolCallID string
	ToolName   string
}

type ToolCallRequest struct {
	ID        string
	Name      string
	Arguments json.RawMessage
}

type ToolDefinition struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

type ReasoningPolicy string

const (
	ReasoningNone   ReasoningPolicy = "none"
	ReasoningLow    ReasoningPolicy = "low"
	ReasoningMedium ReasoningPolicy = "medium"
	ReasoningHigh   ReasoningPolicy = "high"
)

type ModelRequest struct {
	Messages        []ModelMessage
	Tools           []ToolDefinition
	Reasoning       ReasoningPolicy
	MaxOutputTokens int
}

type ProviderEventType string

const (
	ProviderEventTextDelta ProviderEventType = "text_delta"
	ProviderEventToolCall  ProviderEventType = "tool_call"
	ProviderEventUsage     ProviderEventType = "usage"
	ProviderEventCompleted ProviderEventType = "completed"
	ProviderEventError     ProviderEventType = "error"
)

type ProviderEvent struct {
	Type     ProviderEventType
	Text     string
	ToolCall *ToolCallRequest
	Usage    *Usage
	Error    string
}

type Provider interface {
	Stream(ctx context.Context, request ModelRequest) (<-chan ProviderEvent, error)
}
