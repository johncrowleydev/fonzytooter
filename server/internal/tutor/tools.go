package tutor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/danielgtaylor/huma/v2"
	"github.com/johncrowleydev/helix-academy/server/internal/auth"
)

var (
	ErrDuplicateTool        = errors.New("duplicate tutor tool")
	ErrUnknownTool          = errors.New("unknown tutor tool")
	ErrToolNotAllowed       = errors.New("tutor tool is not allowed for this turn")
	ErrToolArgumentsInvalid = errors.New("invalid tutor tool arguments")
)

type Tool interface {
	Definition() ToolDefinition
	Execute(ctx context.Context, userID auth.UserID, arguments json.RawMessage) (json.RawMessage, error)
}

type ProvenanceTool interface {
	Tool
	SourceIDs(result json.RawMessage) []string
}

type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

func NewToolRegistry(tools ...Tool) (*ToolRegistry, error) {
	registry := &ToolRegistry{tools: make(map[string]Tool, len(tools))}
	for _, tool := range tools {
		if err := registry.Register(tool); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func (r *ToolRegistry) Register(tool Tool) error {
	if tool == nil {
		return errors.New("register nil tutor tool")
	}
	definition := tool.Definition()
	if err := validateToolDefinition(definition); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tools[definition.Name]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateTool, definition.Name)
	}
	r.tools[definition.Name] = tool
	return nil
}

func (r *ToolRegistry) Definitions(allowed []string) ([]ToolDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names, err := r.allowedNames(allowed)
	if err != nil {
		return nil, err
	}
	definitions := make([]ToolDefinition, 0, len(names))
	for _, name := range names {
		definition := r.tools[name].Definition()
		definition.InputSchema = cloneJSON(definition.InputSchema)
		definitions = append(definitions, definition)
	}
	return definitions, nil
}

func (r *ToolRegistry) Execute(ctx context.Context, userID auth.UserID, name string, arguments json.RawMessage, allowed []string) (json.RawMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrUnknownTool, name)
	}
	if allowed != nil && !containsString(allowed, name) {
		return nil, fmt.Errorf("%w: %s", ErrToolNotAllowed, name)
	}
	return tool.Execute(ctx, userID, arguments)
}

func (r *ToolRegistry) SourceIDs(name string, result json.RawMessage) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name].(ProvenanceTool)
	if !ok {
		return nil
	}
	values := tool.SourceIDs(result)
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, duplicate := seen[value]; duplicate {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}
	sort.Strings(unique)
	return unique
}

func (r *ToolRegistry) allowedNames(allowed []string) ([]string, error) {
	if allowed == nil {
		names := make([]string, 0, len(r.tools))
		for name := range r.tools {
			names = append(names, name)
		}
		sort.Strings(names)
		return names, nil
	}
	seen := make(map[string]struct{}, len(allowed))
	names := make([]string, 0, len(allowed))
	for _, name := range allowed {
		if _, duplicate := seen[name]; duplicate {
			continue
		}
		if _, exists := r.tools[name]; !exists {
			return nil, fmt.Errorf("%w: %s", ErrUnknownTool, name)
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

type typedTool[Arguments any, Result any] struct {
	definition ToolDefinition
	registry   huma.Registry
	schema     *huma.Schema
	validate   func(Arguments) error
	execute    func(context.Context, auth.UserID, Arguments) (Result, error)
	provenance func(Result) []string
}

func NewTypedTool[Arguments any, Result any](
	name string,
	description string,
	validate func(Arguments) error,
	execute func(context.Context, auth.UserID, Arguments) (Result, error),
) (Tool, error) {
	return newTypedTool(name, description, validate, execute, nil)
}

func NewTypedToolWithProvenance[Arguments any, Result any](
	name string,
	description string,
	validate func(Arguments) error,
	execute func(context.Context, auth.UserID, Arguments) (Result, error),
	provenance func(Result) []string,
) (Tool, error) {
	return newTypedTool(name, description, validate, execute, provenance)
}

func newTypedTool[Arguments any, Result any](
	name string,
	description string,
	validate func(Arguments) error,
	execute func(context.Context, auth.UserID, Arguments) (Result, error),
	provenance func(Result) []string,
) (Tool, error) {
	if execute == nil {
		return nil, errors.New("typed tutor tool execute function is nil")
	}
	registry := huma.NewMapRegistry("#/$defs/", huma.DefaultSchemaNamer)
	argumentType := reflect.TypeFor[Arguments]()
	schema := huma.SchemaFromType(registry, argumentType)
	rawSchema, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("encode tutor tool schema: %w", err)
	}
	definition := ToolDefinition{Name: strings.TrimSpace(name), Description: strings.TrimSpace(description), InputSchema: rawSchema}
	if err := validateToolDefinition(definition); err != nil {
		return nil, err
	}
	return &typedTool[Arguments, Result]{definition: definition, registry: registry, schema: schema, validate: validate, execute: execute, provenance: provenance}, nil
}

func (t *typedTool[Arguments, Result]) SourceIDs(raw json.RawMessage) []string {
	if t.provenance == nil {
		return nil
	}
	var result Result
	if json.Unmarshal(raw, &result) != nil {
		return nil
	}
	return t.provenance(result)
}

func (t *typedTool[Arguments, Result]) Definition() ToolDefinition {
	definition := t.definition
	definition.InputSchema = cloneJSON(definition.InputSchema)
	return definition
}

func (t *typedTool[Arguments, Result]) Execute(ctx context.Context, userID auth.UserID, raw json.RawMessage) (json.RawMessage, error) {
	if !json.Valid(raw) {
		return nil, fmt.Errorf("%w: malformed JSON", ErrToolArgumentsInvalid)
	}
	var untyped any
	if err := json.Unmarshal(raw, &untyped); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrToolArgumentsInvalid, err)
	}
	path := huma.NewPathBuffer(nil, 0)
	validation := &huma.ValidateResult{}
	huma.Validate(t.registry, t.schema, path, huma.ModeWriteToServer, untyped, validation)
	if len(validation.Errors) > 0 {
		return nil, fmt.Errorf("%w: %v", ErrToolArgumentsInvalid, validation.Errors[0])
	}

	var arguments Arguments
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&arguments); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrToolArgumentsInvalid, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("%w: trailing JSON", ErrToolArgumentsInvalid)
	}
	if t.validate != nil {
		if err := t.validate(arguments); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrToolArgumentsInvalid, err)
		}
	}
	result, err := t.execute(ctx, userID, arguments)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("encode tutor tool result: %w", err)
	}
	return encoded, nil
}

func validateToolDefinition(definition ToolDefinition) error {
	if definition.Name == "" {
		return errors.New("tutor tool name is empty")
	}
	if definition.Description == "" {
		return fmt.Errorf("tutor tool %q description is empty", definition.Name)
	}
	if !json.Valid(definition.InputSchema) {
		return fmt.Errorf("tutor tool %q schema is invalid JSON", definition.Name)
	}
	return nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
