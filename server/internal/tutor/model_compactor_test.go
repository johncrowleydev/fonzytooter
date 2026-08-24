package tutor

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type compactionTestProvider struct {
	request ModelRequest
	events  []ProviderEvent
	err     error
}

func (p *compactionTestProvider) Stream(_ context.Context, request ModelRequest) (<-chan ProviderEvent, error) {
	p.request = cloneModelRequest(request)
	if p.err != nil {
		return nil, p.err
	}
	stream := make(chan ProviderEvent, len(p.events))
	for _, event := range p.events {
		stream <- event
	}
	close(stream)
	return stream, nil
}

func TestModelCompactorProducesValidatedStructuredMemory(t *testing.T) {
	provider := &compactionTestProvider{events: []ProviderEvent{
		{Type: ProviderEventTextDelta, Text: `{"summary":"The learner is debugging composition.","structured":{"learnerGoal":"Understand function composition","establishedUnderstanding":["Order changes the result"],"misconceptions":["Composition is commutative"],"unsuccessfulApproaches":["Notation without examples"],"unresolvedQuestions":["How does this apply to pipelines?"],"activeContext":"Debugging a pipeline","sourceIds":["src.functions"],"toolFindings":["Exercise 2 failed its order test"]}}`},
		{Type: ProviderEventCompleted},
	}}
	compactor := ModelCompactor{Provider: provider}
	memory, err := compactor.Compact(context.Background(), CompactionRequest{
		Previous: ConversationMemory{Summary: "Prior fact", Structured: StructuredMemory{LearnerGoal: "Learn functions"}},
		Messages: []CompactionMessage{{Sequence: 3, Message: textModelMessage(ModelRoleUser, "Why does order matter?")}},
	})
	if err != nil {
		t.Fatalf("compact semantic memory: %v", err)
	}
	if memory.Structured.LearnerGoal != "Understand function composition" || len(memory.Structured.Misconceptions) != 1 || memory.Structured.SourceIDs[0] != "src.functions" {
		t.Fatalf("unexpected structured memory: %#v", memory)
	}
	requestText := joinedModelText(provider.request.Messages)
	for _, required := range []string{"Prior fact", "Learn functions", "Why does order matter?"} {
		if !strings.Contains(requestText, required) {
			t.Fatalf("semantic compaction request missing %q: %s", required, requestText)
		}
	}
	if len(provider.request.Tools) != 0 || provider.request.Reasoning != ReasoningLow || provider.request.MaxOutputTokens != DefaultCompactionOutputTokens {
		t.Fatalf("unexpected semantic compaction model request: %#v", provider.request)
	}
}

func TestModelCompactorFallsBackOnProviderOrValidationFailure(t *testing.T) {
	tests := []struct {
		name     string
		provider *compactionTestProvider
	}{
		{name: "provider error", provider: &compactionTestProvider{err: errors.New("offline")}},
		{name: "invalid document", provider: &compactionTestProvider{events: []ProviderEvent{{Type: ProviderEventTextDelta, Text: `{"summary":"","structured":{},"unexpected":true}`}, {Type: ProviderEventCompleted}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fallback := &capturingCompactor{memory: ConversationMemory{Summary: "deterministic fallback", Structured: StructuredMemory{LearnerGoal: "retain me"}}}
			memory, err := (ModelCompactor{Provider: test.provider, Fallback: fallback}).Compact(context.Background(), CompactionRequest{
				Messages: []CompactionMessage{{Sequence: 1, Message: textModelMessage(ModelRoleUser, "hello")}},
			})
			if err != nil {
				t.Fatalf("run fallback: %v", err)
			}
			if memory.Summary != "deterministic fallback" || memory.Structured.LearnerGoal != "retain me" || len(fallback.requests) != 1 {
				t.Fatalf("unexpected fallback memory: %#v", memory)
			}
		})
	}
}
