package openrouter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
)

const (
	defaultBaseURL = "https://openrouter.ai/api/v1"
	providerID     = "openrouter"
)

type Config struct {
	APIKey  string
	Model   string
	BaseURL string
	Client  *http.Client
}

type Provider struct {
	apiKey   string
	model    string
	endpoint string
	client   *http.Client
}

type transportError struct {
	message string
	cause   error
}

func (e *transportError) Error() string { return e.message }
func (e *transportError) Unwrap() error { return e.cause }

func New(config Config) (*Provider, error) {
	if strings.TrimSpace(config.APIKey) == "" {
		return nil, errors.New("OpenRouter API key is required")
	}
	if strings.TrimSpace(config.Model) == "" {
		return nil, errors.New("OpenRouter model is required")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid OpenRouter base URL %q", baseURL)
	}
	client := config.Client
	if client == nil {
		client = http.DefaultClient
	}
	return &Provider{
		apiKey:   config.APIKey,
		model:    config.Model,
		endpoint: baseURL + "/chat/completions",
		client:   client,
	}, nil
}

func (p *Provider) Stream(ctx context.Context, request tutor.ModelRequest) (<-chan tutor.ProviderEvent, error) {
	body, err := encodeRequest(p.model, request, true)
	if err != nil {
		return nil, err
	}
	response, err := p.send(ctx, body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		providerErr := p.responseError(response)
		if response.StatusCode == http.StatusBadRequest && body.Reasoning != nil && isUnsupportedReasoning(providerErr.Error()) {
			body.Reasoning = nil
			response, err = p.send(ctx, body)
			if err != nil {
				return nil, err
			}
			if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
				return nil, p.responseError(response)
			}
		} else {
			return nil, providerErr
		}
	}

	events := make(chan tutor.ProviderEvent)
	go func() {
		defer close(events)
		defer response.Body.Close()
		if err := parseStream(ctx, response.Body, events, p.model); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			sendEvent(ctx, events, tutor.ProviderEvent{Type: tutor.ProviderEventError, Error: p.sanitize(err.Error())})
		}
	}()
	return events, nil
}

func (p *Provider) send(ctx context.Context, body chatRequest) (*http.Response, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("encode OpenRouter request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("create OpenRouter request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+p.apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "text/event-stream")
	response, err := p.client.Do(request)
	if err != nil {
		return nil, &transportError{message: "send OpenRouter request: " + p.sanitize(err.Error()), cause: err}
	}
	return response, nil
}

func (p *Provider) responseError(response *http.Response) error {
	defer response.Body.Close()
	limited, err := io.ReadAll(io.LimitReader(response.Body, 64<<10))
	if err != nil {
		return fmt.Errorf("OpenRouter HTTP %d", response.StatusCode)
	}
	message := providerErrorMessage(limited)
	if message == "" {
		message = http.StatusText(response.StatusCode)
	}
	return fmt.Errorf("OpenRouter HTTP %d: %s", response.StatusCode, p.sanitize(message))
}

func (p *Provider) sanitize(message string) string {
	message = strings.ReplaceAll(message, p.apiKey, "[REDACTED]")
	for _, prefix := range []string{"Authorization: Bearer ", "authorization: bearer ", "Bearer ", "bearer "} {
		searchFrom := 0
		for {
			relative := strings.Index(message[searchFrom:], prefix)
			if relative < 0 {
				break
			}
			index := searchFrom + relative
			end := index + len(prefix)
			for end < len(message) && message[end] != ' ' && message[end] != '\r' && message[end] != '\n' && message[end] != '"' {
				end++
			}
			credentialStart := index + len(prefix)
			if message[credentialStart:end] != "[REDACTED]" {
				message = message[:credentialStart] + "[REDACTED]" + message[end:]
				end = credentialStart + len("[REDACTED]")
			}
			searchFrom = end
		}
	}
	return message
}

type chatRequest struct {
	Model             string           `json:"model"`
	Messages          []chatMessage    `json:"messages"`
	Stream            bool             `json:"stream"`
	MaxTokens         int              `json:"max_tokens,omitempty"`
	ParallelToolCalls *bool            `json:"parallel_tool_calls,omitempty"`
	Tools             []chatTool       `json:"tools,omitempty"`
	Reasoning         *reasoningConfig `json:"reasoning,omitempty"`
}

type chatMessage struct {
	Role             string          `json:"role"`
	Content          any             `json:"content"`
	ToolCalls        []chatToolCall  `json:"tool_calls,omitempty"`
	ToolCallID       string          `json:"tool_call_id,omitempty"`
	Name             string          `json:"name,omitempty"`
	ReasoningDetails json.RawMessage `json:"reasoning_details,omitempty"`
}

type contentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *imageURL `json:"image_url,omitempty"`
}

type imageURL struct {
	URL string `json:"url"`
}

type chatTool struct {
	Type     string           `json:"type"`
	Function chatToolFunction `json:"function"`
}

type chatToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters"`
}

type chatToolCall struct {
	ID       string               `json:"id"`
	Type     string               `json:"type"`
	Function chatToolCallFunction `json:"function"`
}

type chatToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type reasoningConfig struct {
	Effort  string `json:"effort"`
	Exclude bool   `json:"exclude"`
}

func encodeRequest(model string, request tutor.ModelRequest, includeReasoning bool) (chatRequest, error) {
	messages := make([]chatMessage, 0, len(request.Messages))
	for _, message := range request.Messages {
		encoded, err := encodeMessage(model, message)
		if err != nil {
			return chatRequest{}, err
		}
		messages = append(messages, encoded)
	}
	tools := make([]chatTool, 0, len(request.Tools))
	for _, definition := range request.Tools {
		if strings.TrimSpace(definition.Name) == "" || !json.Valid(definition.InputSchema) {
			return chatRequest{}, fmt.Errorf("invalid canonical tool definition %q", definition.Name)
		}
		tools = append(tools, chatTool{Type: "function", Function: chatToolFunction{
			Name: definition.Name, Description: definition.Description, Parameters: definition.InputSchema,
		}})
	}
	result := chatRequest{
		Model: model, Messages: messages, Stream: true, MaxTokens: request.MaxOutputTokens, Tools: tools,
	}
	if len(tools) > 0 {
		enabled := true
		result.ParallelToolCalls = &enabled
	}
	if includeReasoning {
		result.Reasoning = mapReasoning(request.Reasoning)
	}
	return result, nil
}

func encodeMessage(model string, message tutor.ModelMessage) (chatMessage, error) {
	role := string(message.Role)
	if role != string(tutor.ModelRoleSystem) && role != string(tutor.ModelRoleUser) && role != string(tutor.ModelRoleAssistant) && role != string(tutor.ModelRoleTool) {
		return chatMessage{}, fmt.Errorf("unsupported canonical model role %q", message.Role)
	}
	content, err := encodeContent(message.Parts)
	if err != nil {
		return chatMessage{}, err
	}
	result := chatMessage{Role: role, Content: content, ToolCallID: message.ToolCallID, Name: message.ToolName}
	if message.Continuation != nil && message.Continuation.Provider == providerID && message.Continuation.Model == model {
		var details []json.RawMessage
		if message.Role != tutor.ModelRoleAssistant || json.Unmarshal(message.Continuation.State, &details) != nil || len(details) == 0 {
			return chatMessage{}, errors.New("canonical provider continuation state is invalid")
		}
		result.ReasoningDetails = append(json.RawMessage(nil), message.Continuation.State...)
	}
	for _, call := range message.ToolCalls {
		if strings.TrimSpace(call.ID) == "" || strings.TrimSpace(call.Name) == "" || !json.Valid(call.Arguments) {
			return chatMessage{}, fmt.Errorf("invalid canonical tool call %q", call.ID)
		}
		result.ToolCalls = append(result.ToolCalls, chatToolCall{
			ID: call.ID, Type: "function", Function: chatToolCallFunction{Name: call.Name, Arguments: string(call.Arguments)},
		})
	}
	if message.Role == tutor.ModelRoleTool && strings.TrimSpace(message.ToolCallID) == "" {
		return chatMessage{}, errors.New("canonical tool result is missing its tool call ID")
	}
	return result, nil
}

func encodeContent(parts []tutor.ModelContentPart) (any, error) {
	if len(parts) == 0 {
		return "", nil
	}
	allText := true
	var text strings.Builder
	for _, part := range parts {
		if part.Kind != tutor.ModelContentText {
			allText = false
			break
		}
		text.WriteString(part.Text)
	}
	if allText {
		return text.String(), nil
	}
	encoded := make([]contentPart, 0, len(parts))
	for _, part := range parts {
		switch part.Kind {
		case tutor.ModelContentText:
			encoded = append(encoded, contentPart{Type: "text", Text: part.Text})
		case tutor.ModelContentImage:
			if strings.TrimSpace(part.URL) == "" {
				return nil, errors.New("canonical image content is missing a URL")
			}
			encoded = append(encoded, contentPart{Type: "image_url", ImageURL: &imageURL{URL: part.URL}})
		case tutor.ModelContentDocument:
			return nil, errors.New("OpenRouter document content is not supported by the tutor adapter")
		default:
			return nil, fmt.Errorf("unsupported canonical content kind %q", part.Kind)
		}
	}
	return encoded, nil
}

func mapReasoning(policy tutor.ReasoningPolicy) *reasoningConfig {
	switch policy {
	case tutor.ReasoningNone:
		return &reasoningConfig{Effort: "none", Exclude: false}
	case tutor.ReasoningMedium:
		return &reasoningConfig{Effort: "medium", Exclude: false}
	case tutor.ReasoningHigh:
		return &reasoningConfig{Effort: "high", Exclude: false}
	case tutor.ReasoningLow:
		return &reasoningConfig{Effort: "low", Exclude: false}
	default:
		return nil
	}
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content          *string           `json:"content"`
			ToolCalls        []toolCallDelta   `json:"tool_calls"`
			ReasoningDetails []json.RawMessage `json:"reasoning_details"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Usage *streamUsage   `json:"usage"`
	Error *providerError `json:"error"`
}

type toolCallDelta struct {
	Index    int    `json:"index"`
	ID       string `json:"id"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type streamUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	PromptDetails    struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"prompt_tokens_details"`
	CompletionDetails struct {
		ReasoningTokens int `json:"reasoning_tokens"`
	} `json:"completion_tokens_details"`
}

type providerError struct {
	Message string          `json:"message"`
	Code    json.RawMessage `json:"code"`
}

type pendingToolCall struct {
	id        string
	name      string
	arguments strings.Builder
}

func parseStream(ctx context.Context, input io.Reader, output chan<- tutor.ProviderEvent, model string) error {
	reader := bufio.NewReader(input)
	pending := make(map[int]*pendingToolCall)
	reasoningDetails := make([]json.RawMessage, 0)
	stateEmitted := false
	sawFinish := false
	for {
		data, done, err := readSSEEvent(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return errors.New("OpenRouter stream ended before data: [DONE]")
			}
			return fmt.Errorf("read OpenRouter stream: %w", err)
		}
		if done {
			if !sawFinish {
				return errors.New("OpenRouter stream completed without a finish reason")
			}
			if len(pending) > 0 {
				if err := emitToolCalls(ctx, output, pending); err != nil {
					return err
				}
			}
			if !sendEvent(ctx, output, tutor.ProviderEvent{Type: tutor.ProviderEventCompleted}) {
				return ctx.Err()
			}
			return nil
		}
		if len(data) == 0 {
			continue
		}
		var chunk streamChunk
		if err := json.Unmarshal(data, &chunk); err != nil {
			return fmt.Errorf("decode OpenRouter stream event: %w", err)
		}
		if chunk.Error != nil {
			return errors.New(providerErrorText(*chunk.Error))
		}
		for _, choice := range chunk.Choices {
			for _, detail := range choice.Delta.ReasoningDetails {
				if !json.Valid(detail) {
					return errors.New("OpenRouter returned invalid reasoning details")
				}
				reasoningDetails = append(reasoningDetails, append(json.RawMessage(nil), detail...))
			}
			if choice.Delta.Content != nil && *choice.Delta.Content != "" {
				if !sendEvent(ctx, output, tutor.ProviderEvent{Type: tutor.ProviderEventTextDelta, Text: *choice.Delta.Content}) {
					return ctx.Err()
				}
			}
			for _, delta := range choice.Delta.ToolCalls {
				call := pending[delta.Index]
				if call == nil {
					call = &pendingToolCall{}
					pending[delta.Index] = call
				}
				if delta.ID != "" {
					if call.id != "" && call.id != delta.ID {
						return fmt.Errorf("OpenRouter changed tool call ID at index %d", delta.Index)
					}
					call.id = delta.ID
				}
				if delta.Function.Name != "" {
					call.name += delta.Function.Name
				}
				call.arguments.WriteString(delta.Function.Arguments)
			}
			if choice.FinishReason != nil {
				sawFinish = true
				if len(reasoningDetails) > 0 && !stateEmitted {
					state, err := json.Marshal(reasoningDetails)
					if err != nil {
						return fmt.Errorf("encode OpenRouter continuation state: %w", err)
					}
					if !sendEvent(ctx, output, tutor.ProviderEvent{Type: tutor.ProviderEventState, Continuation: &tutor.ProviderContinuation{
						Provider: providerID, Model: model, State: state,
					}}) {
						return ctx.Err()
					}
					stateEmitted = true
				}
				if len(pending) > 0 {
					if err := emitToolCalls(ctx, output, pending); err != nil {
						return err
					}
					pending = make(map[int]*pendingToolCall)
				}
			}
		}
		if chunk.Usage != nil {
			usage := &tutor.Usage{
				InputTokens: chunk.Usage.PromptTokens, OutputTokens: chunk.Usage.CompletionTokens,
				TotalTokens: chunk.Usage.TotalTokens, CachedTokens: chunk.Usage.PromptDetails.CachedTokens,
				ReasoningTokens: chunk.Usage.CompletionDetails.ReasoningTokens,
			}
			if !sendEvent(ctx, output, tutor.ProviderEvent{Type: tutor.ProviderEventUsage, Usage: usage}) {
				return ctx.Err()
			}
		}
	}
}

func emitToolCalls(ctx context.Context, output chan<- tutor.ProviderEvent, pending map[int]*pendingToolCall) error {
	indexes := make([]int, 0, len(pending))
	for index := range pending {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)
	for _, index := range indexes {
		call := pending[index]
		arguments := json.RawMessage(call.arguments.String())
		if strings.TrimSpace(call.id) == "" || strings.TrimSpace(call.name) == "" || !json.Valid(arguments) {
			return fmt.Errorf("OpenRouter returned an incomplete tool call at index %d", index)
		}
		if !sendEvent(ctx, output, tutor.ProviderEvent{Type: tutor.ProviderEventToolCall, ToolCall: &tutor.ToolCallRequest{
			ID: call.id, Name: call.name, Arguments: arguments,
		}}) {
			return ctx.Err()
		}
	}
	return nil
}

func readSSEEvent(reader *bufio.Reader) ([]byte, bool, error) {
	var data []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, false, err
		}
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if line == "" {
			if len(data) > 0 {
				joined := strings.Join(data, "\n")
				return []byte(joined), joined == "[DONE]", nil
			}
			if errors.Is(err, io.EOF) {
				return nil, false, io.EOF
			}
			continue
		}
		if strings.HasPrefix(line, "data:") {
			value := strings.TrimPrefix(line, "data:")
			value = strings.TrimPrefix(value, " ")
			data = append(data, value)
		}
		if errors.Is(err, io.EOF) {
			if len(data) > 0 {
				joined := strings.Join(data, "\n")
				return []byte(joined), joined == "[DONE]", nil
			}
			return nil, false, io.EOF
		}
	}
}

func sendEvent(ctx context.Context, output chan<- tutor.ProviderEvent, event tutor.ProviderEvent) bool {
	select {
	case <-ctx.Done():
		return false
	case output <- event:
		return true
	}
}

func providerErrorMessage(body []byte) string {
	var envelope struct {
		Error providerError `json:"error"`
	}
	if json.Unmarshal(body, &envelope) == nil {
		return providerErrorText(envelope.Error)
	}
	return ""
}

func providerErrorText(providerErr providerError) string {
	if providerErr.Message == "" {
		return "OpenRouter provider error"
	}
	if len(providerErr.Code) > 0 && string(providerErr.Code) != "null" {
		return fmt.Sprintf("%s (code %s)", providerErr.Message, providerErr.Code)
	}
	return providerErr.Message
}

func isUnsupportedReasoning(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "reasoning") && (strings.Contains(lower, "unsupported") || strings.Contains(lower, "not support") || strings.Contains(lower, "unknown"))
}
