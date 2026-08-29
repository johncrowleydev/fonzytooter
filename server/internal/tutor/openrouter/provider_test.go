package openrouter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/johncrowleydev/helix-academy/server/internal/tutor"
)

func TestNewRequiresConfiguration(t *testing.T) {
	if _, err := New(Config{Model: "vendor/model"}); err == nil || !strings.Contains(err.Error(), "API key") {
		t.Fatalf("expected missing API key error, got %v", err)
	}
	if _, err := New(Config{APIKey: "secret"}); err == nil || !strings.Contains(err.Error(), "model") {
		t.Fatalf("expected missing model error, got %v", err)
	}
	if _, err := New(Config{APIKey: "secret", Model: "vendor/model", BaseURL: "://bad"}); err == nil {
		t.Fatal("expected invalid base URL error")
	}
}

func TestProviderStreamsTextUsageAndMapsReasoning(t *testing.T) {
	var captured chatRequest
	provider, server := newFixtureProvider(t, func(w http.ResponseWriter, r *http.Request) {
		assertRequestHeaders(t, r)
		decodeRequest(t, r, &captured)
		writeSSE(t, w,
			`{"choices":[{"delta":{"content":"Hello"},"finish_reason":null}]}`,
			`{"choices":[{"delta":{"content":" tutor"},"finish_reason":null}]}`,
			`{"choices":[{"delta":{},"finish_reason":"stop"}]}`,
			`{"choices":[],"usage":{"prompt_tokens":11,"completion_tokens":5,"total_tokens":16,"prompt_tokens_details":{"cached_tokens":3},"completion_tokens_details":{"reasoning_tokens":2}}}`,
			"[DONE]",
		)
	})
	defer server.Close()

	events, err := provider.Stream(context.Background(), tutor.ModelRequest{
		Messages:  []tutor.ModelMessage{{Role: tutor.ModelRoleUser, Parts: []tutor.ModelContentPart{{Kind: tutor.ModelContentText, Text: "Hi"}}}},
		Reasoning: tutor.ReasoningHigh, MaxOutputTokens: 321,
	})
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	collected := collectEvents(events)
	assertEventTypes(t, collected, tutor.ProviderEventTextDelta, tutor.ProviderEventTextDelta, tutor.ProviderEventUsage, tutor.ProviderEventCompleted)
	if collected[0].Text+collected[1].Text != "Hello tutor" {
		t.Fatalf("unexpected text chunks: %#v", collected)
	}
	if usage := collected[2].Usage; usage == nil || usage.InputTokens != 11 || usage.OutputTokens != 5 || usage.TotalTokens != 16 || usage.CachedTokens != 3 || usage.ReasoningTokens != 2 {
		t.Fatalf("unexpected usage: %#v", usage)
	}
	if captured.Model != "vendor/model" || !captured.Stream || captured.MaxTokens != 321 {
		t.Fatalf("unexpected request metadata: %#v", captured)
	}
	if captured.Reasoning == nil || captured.Reasoning.Effort != "high" || captured.Reasoning.Exclude {
		t.Fatalf("unexpected reasoning mapping: %#v", captured.Reasoning)
	}
}

func TestProviderReconstructsMultipleFragmentedToolCalls(t *testing.T) {
	provider, server := newFixtureProvider(t, func(w http.ResponseWriter, _ *http.Request) {
		writeSSE(t, w,
			`{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call-a","type":"function","function":{"name":"lookup","arguments":"{\"query\":"}},{"index":1,"id":"call-b","type":"function","function":{"name":"state","arguments":"{\"id\":"}}]},"finish_reason":null}]}`,
			`{"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"gradient\"}"}},{"index":1,"function":{"arguments":"\"obj-1\"}"}}]},"finish_reason":null}]}`,
			`{"choices":[{"delta":{},"finish_reason":"tool_calls"}]}`,
			`{"choices":[],"usage":{"prompt_tokens":20,"completion_tokens":8,"total_tokens":28}}`,
			"[DONE]",
		)
	})
	defer server.Close()

	events, err := provider.Stream(context.Background(), tutor.ModelRequest{Messages: textMessages("find it")})
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	collected := collectEvents(events)
	assertEventTypes(t, collected, tutor.ProviderEventToolCall, tutor.ProviderEventToolCall, tutor.ProviderEventUsage, tutor.ProviderEventCompleted)
	if call := collected[0].ToolCall; call == nil || call.ID != "call-a" || call.Name != "lookup" || string(call.Arguments) != `{"query":"gradient"}` {
		t.Fatalf("unexpected first tool call: %#v", call)
	}
	if call := collected[1].ToolCall; call == nil || call.ID != "call-b" || call.Name != "state" || string(call.Arguments) != `{"id":"obj-1"}` {
		t.Fatalf("unexpected second tool call: %#v", call)
	}
}

func TestProviderPreservesOpaqueReasoningDetailsAcrossToolContinuation(t *testing.T) {
	var mu sync.Mutex
	var requests []chatRequest
	provider, server := newFixtureProvider(t, func(w http.ResponseWriter, r *http.Request) {
		var captured chatRequest
		decodeRequest(t, r, &captured)
		mu.Lock()
		requests = append(requests, captured)
		requestNumber := len(requests)
		mu.Unlock()
		if requestNumber == 1 {
			writeSSE(t, w,
				`{"choices":[{"delta":{"reasoning_details":[{"type":"reasoning.summary","summary":"Need evidence","id":"summary-1","format":"anthropic-claude-v1","index":0}]} ,"finish_reason":null}]}`,
				`{"choices":[{"delta":{"reasoning_details":[{"type":"reasoning.encrypted","data":"opaque-data","id":"encrypted-1","format":"anthropic-claude-v1","index":1}],"tool_calls":[{"index":0,"id":"call-1","type":"function","function":{"name":"lookup","arguments":"{}"}}]},"finish_reason":null}]}`,
				`{"choices":[{"delta":{},"finish_reason":"tool_calls"}]}`,
				"[DONE]",
			)
			return
		}
		writeSSE(t, w, `{"choices":[{"delta":{"content":"continued"},"finish_reason":"stop"}]}`, "[DONE]")
	})
	defer server.Close()

	first, err := provider.Stream(context.Background(), tutor.ModelRequest{Messages: textMessages("Find evidence"), Reasoning: tutor.ReasoningHigh})
	if err != nil {
		t.Fatalf("start tool request: %v", err)
	}
	firstEvents := collectEvents(first)
	assertEventTypes(t, firstEvents, tutor.ProviderEventState, tutor.ProviderEventToolCall, tutor.ProviderEventCompleted)
	state := firstEvents[0].Continuation
	if state == nil || state.Provider != providerID || state.Model != "vendor/model" || string(state.State) != `[{"type":"reasoning.summary","summary":"Need evidence","id":"summary-1","format":"anthropic-claude-v1","index":0},{"type":"reasoning.encrypted","data":"opaque-data","id":"encrypted-1","format":"anthropic-claude-v1","index":1}]` {
		t.Fatalf("unexpected opaque continuation state: %#v", state)
	}

	second, err := provider.Stream(context.Background(), tutor.ModelRequest{Messages: []tutor.ModelMessage{
		{Role: tutor.ModelRoleUser, Parts: []tutor.ModelContentPart{{Kind: tutor.ModelContentText, Text: "Find evidence"}}},
		{Role: tutor.ModelRoleAssistant, ToolCalls: []tutor.ToolCallRequest{{ID: "call-1", Name: "lookup", Arguments: json.RawMessage(`{}`)}}, Continuation: state},
		{Role: tutor.ModelRoleTool, ToolCallID: "call-1", ToolName: "lookup", Parts: []tutor.ModelContentPart{{Kind: tutor.ModelContentText, Text: `{"found":true}`}}},
	}})
	if err != nil {
		t.Fatalf("start continuation request: %v", err)
	}
	assertEventTypes(t, collectEvents(second), tutor.ProviderEventTextDelta, tutor.ProviderEventCompleted)
	mu.Lock()
	defer mu.Unlock()
	if len(requests) != 2 || string(requests[1].Messages[1].ReasoningDetails) != string(state.State) {
		t.Fatalf("reasoning details were not echoed unchanged: %#v", requests)
	}
}

func TestProviderSerializesToolsResultsAndImages(t *testing.T) {
	var captured chatRequest
	provider, server := newFixtureProvider(t, func(w http.ResponseWriter, r *http.Request) {
		decodeRequest(t, r, &captured)
		writeSSE(t, w, `{"choices":[{"delta":{},"finish_reason":"stop"}]}`, "[DONE]")
	})
	defer server.Close()

	request := tutor.ModelRequest{
		Messages: []tutor.ModelMessage{
			{Role: tutor.ModelRoleUser, Parts: []tutor.ModelContentPart{
				{Kind: tutor.ModelContentText, Text: "Inspect this"},
				{Kind: tutor.ModelContentImage, URL: "data:image/png;base64,AA==", MediaType: "image/png"},
			}},
			{Role: tutor.ModelRoleAssistant, ToolCalls: []tutor.ToolCallRequest{{ID: "call-1", Name: "lookup", Arguments: json.RawMessage(`{"query":"x"}`)}}},
			{Role: tutor.ModelRoleTool, ToolCallID: "call-1", ToolName: "lookup", Parts: []tutor.ModelContentPart{{Kind: tutor.ModelContentText, Text: `{"result":"ok"}`}}},
		},
		Tools:     []tutor.ToolDefinition{{Name: "lookup", Description: "Look up evidence", InputSchema: json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}},"required":["query"]}`)}},
		Reasoning: tutor.ReasoningNone,
	}
	events, err := provider.Stream(context.Background(), request)
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	assertEventTypes(t, collectEvents(events), tutor.ProviderEventCompleted)
	if len(captured.Messages) != 3 || len(captured.Tools) != 1 || captured.ParallelToolCalls == nil || !*captured.ParallelToolCalls {
		t.Fatalf("unexpected tool request: %#v", captured)
	}
	parts, ok := captured.Messages[0].Content.([]any)
	if !ok || len(parts) != 2 {
		t.Fatalf("expected multimodal content array, got %#v", captured.Messages[0].Content)
	}
	encoded, _ := json.Marshal(parts[1])
	if !strings.Contains(string(encoded), "image_url") || !strings.Contains(string(encoded), "data:image/png") {
		t.Fatalf("image part was not serialized: %s", encoded)
	}
	if captured.Messages[1].ToolCalls[0].ID != "call-1" || captured.Messages[2].ToolCallID != "call-1" || captured.Messages[2].Content != `{"result":"ok"}` {
		t.Fatalf("tool correlation was not preserved: %#v", captured.Messages)
	}
	if captured.Reasoning == nil || captured.Reasoning.Effort != "none" {
		t.Fatalf("unexpected none reasoning mapping: %#v", captured.Reasoning)
	}
}

func TestProviderRetriesWithoutUnsupportedReasoning(t *testing.T) {
	var mu sync.Mutex
	var requests []chatRequest
	provider, server := newFixtureProvider(t, func(w http.ResponseWriter, r *http.Request) {
		var captured chatRequest
		decodeRequest(t, r, &captured)
		mu.Lock()
		requests = append(requests, captured)
		count := len(requests)
		mu.Unlock()
		if count == 1 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"reasoning is not supported by this model","code":400}}`))
			return
		}
		writeSSE(t, w, `{"choices":[{"delta":{"content":"ok"},"finish_reason":"stop"}]}`, "[DONE]")
	})
	defer server.Close()

	events, err := provider.Stream(context.Background(), tutor.ModelRequest{Messages: textMessages("hello"), Reasoning: tutor.ReasoningLow})
	if err != nil {
		t.Fatalf("fallback stream: %v", err)
	}
	assertEventTypes(t, collectEvents(events), tutor.ProviderEventTextDelta, tutor.ProviderEventCompleted)
	if len(requests) != 2 || requests[0].Reasoning == nil || requests[1].Reasoning != nil {
		t.Fatalf("expected one retry without reasoning, got %#v", requests)
	}
}

func TestProviderSanitizesHTTPAndStreamErrors(t *testing.T) {
	const secret = "super-secret-key"
	t.Run("HTTP error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = fmt.Fprintf(w, `{"error":{"message":"quota failed for Bearer %s","code":429}}`, secret)
		}))
		defer server.Close()
		provider := mustProvider(t, Config{APIKey: secret, Model: "vendor/model", BaseURL: server.URL})
		_, err := provider.Stream(context.Background(), tutor.ModelRequest{Messages: textMessages("hello")})
		if err == nil || !strings.Contains(err.Error(), "HTTP 429") || strings.Contains(err.Error(), secret) {
			t.Fatalf("expected sanitized HTTP error, got %v", err)
		}
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"error":{"message":"upstream provider unavailable","code":502}}`))
		}))
		defer server.Close()
		provider := mustProvider(t, Config{APIKey: secret, Model: "vendor/model", BaseURL: server.URL})
		_, err := provider.Stream(context.Background(), tutor.ModelRequest{Messages: textMessages("hello")})
		if err == nil || !strings.Contains(err.Error(), "HTTP 502") || !strings.Contains(err.Error(), "upstream provider unavailable") {
			t.Fatalf("expected useful server error, got %v", err)
		}
	})

	t.Run("stream error", func(t *testing.T) {
		provider, server := newFixtureProviderWithKey(t, secret, func(w http.ResponseWriter, _ *http.Request) {
			writeSSE(t, w, fmt.Sprintf(`{"error":{"message":"provider rejected Bearer %s","code":502}}`, secret))
		})
		defer server.Close()
		events, err := provider.Stream(context.Background(), tutor.ModelRequest{Messages: textMessages("hello")})
		if err != nil {
			t.Fatalf("start stream: %v", err)
		}
		collected := collectEvents(events)
		if len(collected) != 1 || collected[0].Type != tutor.ProviderEventError || strings.Contains(collected[0].Error, secret) {
			t.Fatalf("expected sanitized stream error, got %#v", collected)
		}
	})
}

func TestReasoningPolicyMapping(t *testing.T) {
	tests := map[tutor.ReasoningPolicy]string{
		tutor.ReasoningNone: "none", tutor.ReasoningLow: "low",
		tutor.ReasoningMedium: "medium", tutor.ReasoningHigh: "high",
	}
	for policy, effort := range tests {
		mapped := mapReasoning(policy)
		if mapped == nil || mapped.Effort != effort || mapped.Exclude {
			t.Fatalf("policy %q mapped incorrectly: %#v", policy, mapped)
		}
	}
	if mapped := mapReasoning(tutor.ReasoningPolicy("invalid")); mapped != nil {
		t.Fatalf("invalid policy should be omitted, got %#v", mapped)
	}
}

func TestProviderRejectsMalformedAndPartialStreams(t *testing.T) {
	tests := []struct {
		name    string
		fixture []string
		want    string
	}{
		{name: "invalid JSON", fixture: []string{`{"choices":`}, want: "decode OpenRouter"},
		{name: "missing done", fixture: []string{`{"choices":[{"delta":{},"finish_reason":"stop"}]}`}, want: "before data: [DONE]"},
		{name: "missing finish", fixture: []string{"[DONE]"}, want: "without a finish reason"},
		{name: "partial tool", fixture: []string{`{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call","function":{"name":"lookup","arguments":"{\"x\":"}}]},"finish_reason":"tool_calls"}]}`}, want: "incomplete tool call"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider, server := newFixtureProvider(t, func(w http.ResponseWriter, _ *http.Request) {
				writeSSE(t, w, test.fixture...)
			})
			defer server.Close()
			events, err := provider.Stream(context.Background(), tutor.ModelRequest{Messages: textMessages("hello")})
			if err != nil {
				t.Fatalf("start stream: %v", err)
			}
			collected := collectEvents(events)
			if len(collected) == 0 || collected[len(collected)-1].Type != tutor.ProviderEventError || !strings.Contains(collected[len(collected)-1].Error, test.want) {
				t.Fatalf("expected error containing %q, got %#v", test.want, collected)
			}
		})
	}
}

func TestProviderHonorsCancellationAndTimeout(t *testing.T) {
	t.Run("stream cancellation", func(t *testing.T) {
		started := make(chan struct{})
		provider, server := newFixtureProvider(t, func(w http.ResponseWriter, r *http.Request) {
			flusher := w.(http.Flusher)
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			flusher.Flush()
			close(started)
			<-r.Context().Done()
		})
		defer server.Close()
		ctx, cancel := context.WithCancel(context.Background())
		events, err := provider.Stream(ctx, tutor.ModelRequest{Messages: textMessages("hello")})
		if err != nil {
			t.Fatalf("start stream: %v", err)
		}
		<-started
		cancel()
		select {
		case _, ok := <-events:
			if ok {
				t.Fatal("expected cancelled stream to close")
			}
		case <-time.After(2 * time.Second):
			t.Fatal("cancelled stream did not close")
		}
	})

	t.Run("HTTP timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			time.Sleep(100 * time.Millisecond)
		}))
		defer server.Close()
		provider := mustProvider(t, Config{
			APIKey: "test-key", Model: "vendor/model", BaseURL: server.URL,
			Client: &http.Client{Timeout: 20 * time.Millisecond},
		})
		_, err := provider.Stream(context.Background(), tutor.ModelRequest{Messages: textMessages("hello")})
		if err == nil || !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected timeout error, got %v", err)
		}
	})
}

func TestProviderRejectsUnsupportedCanonicalContent(t *testing.T) {
	provider := mustProvider(t, Config{APIKey: "test-key", Model: "vendor/model"})
	_, err := provider.Stream(context.Background(), tutor.ModelRequest{Messages: []tutor.ModelMessage{{
		Role: tutor.ModelRoleUser, Parts: []tutor.ModelContentPart{{Kind: tutor.ModelContentDocument, URL: "https://example.test/file.pdf"}},
	}}})
	if err == nil || !strings.Contains(err.Error(), "document content") {
		t.Fatalf("expected unsupported document error, got %v", err)
	}
}

func TestProviderRejectsInvalidOpaqueContinuationState(t *testing.T) {
	provider := mustProvider(t, Config{APIKey: "test-key", Model: "vendor/model"})
	_, err := provider.Stream(context.Background(), tutor.ModelRequest{Messages: []tutor.ModelMessage{{
		Role: tutor.ModelRoleAssistant, Continuation: &tutor.ProviderContinuation{
			Provider: providerID, Model: "vendor/model", State: json.RawMessage(`{"not":"an array"}`),
		},
	}}})
	if err == nil || !strings.Contains(err.Error(), "continuation state") {
		t.Fatalf("expected invalid continuation state error, got %v", err)
	}
}

func TestProviderOmitsContinuationStateFromAnotherProviderOrModel(t *testing.T) {
	for _, continuation := range []*tutor.ProviderContinuation{
		{Provider: "other-provider", Model: "vendor/model", State: json.RawMessage(`[{"opaque":true}]`)},
		{Provider: providerID, Model: "other-model", State: json.RawMessage(`[{"opaque":true}]`)},
	} {
		encoded, err := encodeRequest("vendor/model", tutor.ModelRequest{Messages: []tutor.ModelMessage{{
			Role: tutor.ModelRoleAssistant, Parts: []tutor.ModelContentPart{{Kind: tutor.ModelContentText, Text: "prior"}}, Continuation: continuation,
		}}}, true)
		if err != nil {
			t.Fatalf("encode request: %v", err)
		}
		if len(encoded.Messages[0].ReasoningDetails) != 0 {
			t.Fatalf("incompatible continuation state was replayed: %#v", encoded.Messages[0])
		}
	}
}

func newFixtureProvider(t *testing.T, handler http.HandlerFunc) (*Provider, *httptest.Server) {
	return newFixtureProviderWithKey(t, "test-key", handler)
}

func newFixtureProviderWithKey(t *testing.T, key string, handler http.HandlerFunc) (*Provider, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	provider := mustProvider(t, Config{APIKey: key, Model: "vendor/model", BaseURL: server.URL, Client: server.Client()})
	return provider, server
}

func mustProvider(t *testing.T, config Config) *Provider {
	t.Helper()
	provider, err := New(config)
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	return provider
}

func writeSSE(t *testing.T, w http.ResponseWriter, payloads ...string) {
	t.Helper()
	w.Header().Set("Content-Type", "text/event-stream")
	for _, payload := range payloads {
		if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
			t.Fatalf("write SSE: %v", err)
		}
	}
}

func decodeRequest(t *testing.T, r *http.Request, target *chatRequest) {
	t.Helper()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		t.Fatalf("decode request: %v", err)
	}
}

func assertRequestHeaders(t *testing.T, r *http.Request) {
	t.Helper()
	if r.URL.Path != "/chat/completions" || r.Header.Get("Authorization") != "Bearer test-key" || r.Header.Get("Accept") != "text/event-stream" {
		t.Fatalf("unexpected request: %s %#v", r.URL.Path, r.Header)
	}
}

func textMessages(text string) []tutor.ModelMessage {
	return []tutor.ModelMessage{{Role: tutor.ModelRoleUser, Parts: []tutor.ModelContentPart{{Kind: tutor.ModelContentText, Text: text}}}}
}

func collectEvents(events <-chan tutor.ProviderEvent) []tutor.ProviderEvent {
	var collected []tutor.ProviderEvent
	for event := range events {
		collected = append(collected, event)
	}
	return collected
}

func assertEventTypes(t *testing.T, events []tutor.ProviderEvent, expected ...tutor.ProviderEventType) {
	t.Helper()
	if len(events) != len(expected) {
		t.Fatalf("expected event types %v, got %#v", expected, events)
	}
	for index, eventType := range expected {
		if events[index].Type != eventType {
			t.Fatalf("event %d: expected %q, got %#v", index, eventType, events[index])
		}
	}
}
