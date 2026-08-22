package tutor

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type echoToolArguments struct {
	Text string `json:"text" minLength:"1"`
}

type echoToolResult struct {
	Echo string `json:"echo"`
}

func TestTypedToolSchemaValidationAndExecution(t *testing.T) {
	tool, err := NewTypedTool[echoToolArguments, echoToolResult](
		"echo",
		"Echo validated text.",
		func(arguments echoToolArguments) error {
			if strings.TrimSpace(arguments.Text) == "" {
				return errors.New("text must not be blank")
			}
			return nil
		},
		func(_ context.Context, arguments echoToolArguments) (echoToolResult, error) {
			return echoToolResult{Echo: arguments.Text}, nil
		},
	)
	if err != nil {
		t.Fatalf("create typed tool: %v", err)
	}
	definition := tool.Definition()
	if definition.Name != "echo" || !json.Valid(definition.InputSchema) || !strings.Contains(string(definition.InputSchema), `"text"`) {
		t.Fatalf("unexpected tool definition: %#v", definition)
	}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{"text":"hello"}`))
	if err != nil {
		t.Fatalf("execute typed tool: %v", err)
	}
	if string(result) != `{"echo":"hello"}` {
		t.Fatalf("unexpected tool result: %s", result)
	}
	for _, invalid := range []string{`{`, `{}`, `{"text":""}`, `{"text":"ok","extra":true}`, `{"text":"ok"} {}`} {
		if _, err := tool.Execute(context.Background(), json.RawMessage(invalid)); !errors.Is(err, ErrToolArgumentsInvalid) {
			t.Fatalf("expected invalid arguments for %s, got %v", invalid, err)
		}
	}
}

func TestToolRegistryRejectsDuplicatesUnknownAndDisallowedTools(t *testing.T) {
	first, err := NewTypedTool[echoToolArguments, echoToolResult]("echo", "Echo text.", nil, func(_ context.Context, arguments echoToolArguments) (echoToolResult, error) {
		return echoToolResult{Echo: arguments.Text}, nil
	})
	if err != nil {
		t.Fatalf("create first tool: %v", err)
	}
	second, err := NewTypedTool[echoToolArguments, echoToolResult]("second", "A second tool.", nil, func(_ context.Context, arguments echoToolArguments) (echoToolResult, error) {
		return echoToolResult{Echo: arguments.Text}, nil
	})
	if err != nil {
		t.Fatalf("create second tool: %v", err)
	}
	registry, err := NewToolRegistry(first, second)
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	if err := registry.Register(first); !errors.Is(err, ErrDuplicateTool) {
		t.Fatalf("expected duplicate error, got %v", err)
	}
	definitions, err := registry.Definitions([]string{"second", "echo", "echo"})
	if err != nil {
		t.Fatalf("get allowed definitions: %v", err)
	}
	if len(definitions) != 2 || definitions[0].Name != "echo" || definitions[1].Name != "second" {
		t.Fatalf("expected stable sorted definitions, got %#v", definitions)
	}
	if _, err := registry.Definitions([]string{"missing"}); !errors.Is(err, ErrUnknownTool) {
		t.Fatalf("expected unknown definition error, got %v", err)
	}
	if _, err := registry.Execute(context.Background(), "missing", json.RawMessage(`{}`), nil); !errors.Is(err, ErrUnknownTool) {
		t.Fatalf("expected unknown execution error, got %v", err)
	}
	if _, err := registry.Execute(context.Background(), "second", json.RawMessage(`{"text":"ok"}`), []string{"echo"}); !errors.Is(err, ErrToolNotAllowed) {
		t.Fatalf("expected disallowed execution error, got %v", err)
	}
}

func TestTypedToolPropagatesExecutionError(t *testing.T) {
	executionError := errors.New("deterministic failure")
	tool, err := NewTypedTool[echoToolArguments, echoToolResult]("echo", "Echo text.", nil, func(context.Context, echoToolArguments) (echoToolResult, error) {
		return echoToolResult{}, executionError
	})
	if err != nil {
		t.Fatalf("create tool: %v", err)
	}
	if _, err := tool.Execute(context.Background(), json.RawMessage(`{"text":"ok"}`)); !errors.Is(err, executionError) {
		t.Fatalf("expected execution error, got %v", err)
	}
}
