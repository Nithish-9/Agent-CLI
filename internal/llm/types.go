package llm

import (
	"encoding/json"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/sashabaranov/go-openai"
)

type LLMModels struct {
	Models   map[string]*LLMModel
	Planner  *LLMModel
	Executor *LLMModel
}

type LLMModel struct {
	Model  string
	Client *openai.Client
}

func generateSchema(input any) (json.RawMessage, error) {
	reflector := &jsonschema.Reflector{}
	rootSchema := reflector.Reflect(input)

	schemaBytes, err := json.Marshal(rootSchema)
	if err != nil {
		return nil, err
	}

	var raw map[string]any
	if err := json.Unmarshal(schemaBytes, &raw); err != nil {
		return nil, err
	}

	defs, _ := raw["$defs"].(map[string]any)

	if ref, ok := raw["$ref"].(string); ok {
		defName := strings.TrimPrefix(ref, "#/$defs/")
		if def, ok := defs[defName]; ok {
			defMap, _ := def.(map[string]any)
			raw = defMap
		}
	}

	resolved := resolveRefs(raw, defs)
	delete(resolved, "$defs")
	delete(resolved, "$schema")
	delete(resolved, "$ref")

	return json.Marshal(resolved)
}

func resolveRefs(schema map[string]any, defs map[string]any) map[string]any {
	if ref, ok := schema["$ref"].(string); ok {
		defName := strings.TrimPrefix(ref, "#/$defs/")
		if def, ok := defs[defName]; ok {
			defMap, _ := json.Marshal(def)
			var resolved map[string]any
			json.Unmarshal(defMap, &resolved)
			return resolveRefs(resolved, defs)
		}
	}

	for key, val := range schema {
		switch v := val.(type) {
		case map[string]any:
			schema[key] = resolveRefs(v, defs)
		case []any:
			for i, item := range v {
				if itemMap, ok := item.(map[string]any); ok {
					v[i] = resolveRefs(itemMap, defs)
				}
			}
		}
	}
	return schema
}
