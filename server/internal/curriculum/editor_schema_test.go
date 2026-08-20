package curriculum

import (
	"bytes"
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestGenerateEditorSchemasIsDeterministic(t *testing.T) {
	first, err := GenerateEditorSchemas()
	if err != nil {
		t.Fatalf("generate editor schemas: %v", err)
	}
	second, err := GenerateEditorSchemas()
	if err != nil {
		t.Fatalf("generate editor schemas again: %v", err)
	}
	if len(first) != len(second) {
		t.Fatalf("generated schema count changed: %d != %d", len(first), len(second))
	}

	wantNames := []string{
		"course.schema.json",
		"exercise.schema.json",
		"lesson-frontmatter.schema.json",
		"module.schema.json",
		"review-item.schema.json",
		"sources.schema.json",
		"worksheet.schema.json",
	}
	for index, schema := range first {
		if schema.Filename != wantNames[index] {
			t.Errorf("schema %d filename = %q, want %q", index, schema.Filename, wantNames[index])
		}
		if schema.Filename != second[index].Filename || !bytes.Equal(schema.Data, second[index].Data) {
			t.Errorf("schema %q was not generated deterministically", schema.Filename)
		}
		var document map[string]any
		if err := json.Unmarshal(schema.Data, &document); err != nil {
			t.Errorf("schema %q is not valid JSON: %v", schema.Filename, err)
			continue
		}
		if document["$schema"] != editorSchemaDraft {
			t.Errorf("schema %q draft = %#v, want %q", schema.Filename, document["$schema"], editorSchemaDraft)
		}
		if !strings.Contains(document["$comment"].(string), "DO NOT EDIT") {
			t.Errorf("schema %q lacks generated-file warning", schema.Filename)
		}
	}
}

func TestGeneratedEditorSchemaRequiredFieldsMatchAuthoringValidation(t *testing.T) {
	tests := map[string][]string{
		"course.schema.json":             {"description", "id", "order", "title"},
		"exercise.schema.json":           {"id", "lessonId", "objectiveIds", "order", "prompt", "starterCode", "tests", "title"},
		"lesson-frontmatter.schema.json": {"id", "title"},
		"module.schema.json":             {"id", "order", "title"},
		"review-item.schema.json":        {"answer", "id", "objectiveIds", "order", "prompt", "sourceLessonId"},
		"sources.schema.json":            {"sources"},
		"worksheet.schema.json":          {"id", "instructions", "lessonId", "objectiveIds", "order", "problems", "title"},
	}

	for filename, want := range tests {
		t.Run(filename, func(t *testing.T) {
			root := generatedRootSchema(t, filename)
			if root["additionalProperties"] != false {
				t.Fatalf("root additionalProperties = %#v, want false", root["additionalProperties"])
			}
			got := stringArray(t, root["required"])
			slices.Sort(got)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Fatalf("required fields = %v, want %v", got, want)
			}
		})
	}
}

func TestGeneratedEditorSchemasExposeStructuralConstraints(t *testing.T) {
	module := generatedRootSchema(t, "module.schema.json")
	moduleProperties := objectValue(t, module["properties"])
	if moduleProperties["lessons"].(map[string]any)["uniqueItems"] != true {
		t.Fatal("module lessons are not marked unique")
	}
	if _, required := sliceSet(stringArray(t, module["required"]))["lessons"]; required {
		t.Fatal("module lessons should be optional for empty orientation modules")
	}

	exerciseDocument := generatedSchemaDocument(t, "exercise.schema.json")
	exercise := resolveDefinition(t, exerciseDocument, "ExerciseAuthoring")
	exerciseProperties := objectValue(t, exercise["properties"])
	if exerciseProperties["objectiveIds"].(map[string]any)["minItems"] != float64(1) {
		t.Fatal("exercise objectiveIds do not require at least one value")
	}
	testDefinition := resolveDefinition(t, exerciseDocument, "ExerciseTestAuthoring")
	visibility := objectValue(t, testDefinition["properties"])["visibility"].(map[string]any)
	if got := stringArray(t, visibility["enum"]); !slices.Equal(got, []string{"visible", "hidden"}) {
		t.Fatalf("exercise visibility enum = %v", got)
	}

	review := generatedRootSchema(t, "review-item.schema.json")
	if _, required := sliceSet(stringArray(t, review["required"]))["hint"]; required {
		t.Fatal("review hint should be optional")
	}
}

func TestAuthoringYAMLAndJSONFieldNamesMatch(t *testing.T) {
	visited := map[reflect.Type]bool{}
	for _, definition := range editorSchemaDefinitions {
		checkAuthoringTags(t, definition.authoringTy, visited)
	}
}

func checkAuthoringTags(t *testing.T, typ reflect.Type, visited map[reflect.Type]bool) {
	t.Helper()
	for typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Map {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct || visited[typ] {
		return
	}
	visited[typ] = true

	for index := range typ.NumField() {
		field := typ.Field(index)
		yamlName := strings.Split(field.Tag.Get("yaml"), ",")[0]
		jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
		if yamlName == "" || jsonName == "" {
			t.Errorf("%s.%s must declare both yaml and json field names", typ, field.Name)
		} else if yamlName != jsonName {
			t.Errorf("%s.%s yaml name %q does not match json name %q", typ, field.Name, yamlName, jsonName)
		}
		checkAuthoringTags(t, field.Type, visited)
	}
}

func generatedRootSchema(t *testing.T, filename string) map[string]any {
	t.Helper()
	document := generatedSchemaDocument(t, filename)
	ref, ok := document["$ref"].(string)
	if !ok {
		t.Fatalf("schema %q has no root $ref", filename)
	}
	return resolveDefinition(t, document, strings.TrimPrefix(ref, "#/$defs/"))
}

func generatedSchemaDocument(t *testing.T, filename string) map[string]any {
	t.Helper()
	schemas, err := GenerateEditorSchemas()
	if err != nil {
		t.Fatalf("generate editor schemas: %v", err)
	}
	for _, schema := range schemas {
		if schema.Filename != filename {
			continue
		}
		var document map[string]any
		if err := json.Unmarshal(schema.Data, &document); err != nil {
			t.Fatalf("decode %q: %v", filename, err)
		}
		return document
	}
	t.Fatalf("schema %q was not generated", filename)
	return nil
}

func resolveDefinition(t *testing.T, document map[string]any, name string) map[string]any {
	t.Helper()
	definitions := objectValue(t, document["$defs"])
	definition, ok := definitions[name]
	if !ok {
		t.Fatalf("schema definition %q was not generated", name)
	}
	return objectValue(t, definition)
}

func objectValue(t *testing.T, value any) map[string]any {
	t.Helper()
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("value %#v is not an object", value)
	}
	return object
}

func stringArray(t *testing.T, value any) []string {
	t.Helper()
	values, ok := value.([]any)
	if !ok {
		t.Fatalf("value %#v is not an array", value)
	}
	result := make([]string, len(values))
	for index, value := range values {
		text, ok := value.(string)
		if !ok {
			t.Fatalf("array value %#v is not a string", value)
		}
		result[index] = text
	}
	return result
}

func sliceSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}
