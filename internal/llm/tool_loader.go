package llm

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed data/tools/*.json
var toolFS embed.FS

// LoadTool loads a single tool definition from its JSON schema file.
func LoadTool(name string) (ToolDef, error) {
	b, err := toolFS.ReadFile("data/tools/" + name + ".json")
	if err != nil {
		return ToolDef{}, fmt.Errorf("llm: load tool %s: %w", name, err)
	}
	return ParseToolDef(b)
}

// ParseToolDef parses a raw JSON tool schema into a ToolDef.
// The JSON file contains the function definition at the top level;
// we wrap it in {"type":"function","function":{...}} for the API.
func ParseToolDef(raw json.RawMessage) (ToolDef, error) {
	// Validate that it's valid JSON and extract the name for debug logging.
	var schema struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		return ToolDef{}, fmt.Errorf("llm: invalid tool schema: %w", err)
	}
	if schema.Name == "" {
		return ToolDef{}, fmt.Errorf("llm: tool schema missing name")
	}
	return ToolDef{
		Type:     "function",
		Function: raw,
	}, nil
}
