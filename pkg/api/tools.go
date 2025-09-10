package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type ToolsProvider interface {
	Feature[ToolsAttributes]
	// GetMcpSettings returns the MCP settings if the provider uses an MCP server, or nil otherwise
	GetMcpSettings() *McpSettings
	// GetTools returns the list of built-in, native, available tools
	GetTools(ctx context.Context) []*Tool
}

type ToolsAttributes interface {
	FeatureAttributes
}

// ToolsParameters parameters for a tool
// TODO: maybe shareable by policies
type ToolsParameters struct {
	Enabled            *bool `json:"-" toml:"enabled"`
	ReadOnly           *bool `json:"-" toml:"read-only"`
	DisableDestructive *bool `json:"-" toml:"disable-destructive"`
	//Local          *bool
}

type BasicToolsProvider struct {
	ToolsProvider `json:"-"`
	BasicToolsAttributes
	Available         bool         `json:"-"`
	IsAvailableReason string       `json:"reason"`
	McpSettings       *McpSettings `json:"mcp_settings,omitempty"`
	ToolsParameters
}

func (p *BasicToolsProvider) Attributes() ToolsAttributes {
	return &p.BasicToolsAttributes
}

func (p *BasicToolsProvider) IsAvailable() bool {
	return p.Available
}

func (p *BasicToolsProvider) Reason() string {
	return p.IsAvailableReason
}

func (p *BasicToolsProvider) GetMcpSettings() *McpSettings {
	return p.McpSettings
}

func (p *BasicToolsProvider) GetTools(_ context.Context) []*Tool {
	return nil
}

type BasicToolsAttributes struct {
	BasicFeatureAttributes
}

type Tool struct {
	Name        string
	Description string
	// Parameters in map format (if ParametersSchema is not set)
	Parameters map[string]ToolParameter
	Function   func(args map[string]interface{}) (string, error)
}

type ToolParameterType string

const (
	String ToolParameterType = "string"
	// TODO: add more types as needed
)

type ToolParameter struct {
	Type        ToolParameterType
	Description string
	Required    bool
}

type McpType int

const (
	McpTypeStdio McpType = iota
	McpTypeSse
	McpTypeStreamableHttp
)

var McpTypes = [...]string{
	"stdio",
	"sse",
	"http",
}

func (t *McpType) String() string {
	return McpTypes[*t]
}

func (t *McpType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

func (t *McpType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("unmarshaling McpType: %w", err)
	}
	for i, v := range McpTypes {
		if strings.EqualFold(v, s) {
			*t = McpType(i)
			return nil
		}
	}
	return fmt.Errorf("invalid McpType: %s", s)
}

type McpSettings struct {
	Type    McpType           `json:"type"`              // Type of MCP (STDIO, SSE, or HTTP)
	Command string            `json:"command,omitempty"` // Command to run for STDIO type
	Args    []string          `json:"args,omitempty"`    // Arguments for the command (STDIO)
	Env     []string          `json:"env,omitempty"`     // Environment variables for the command (STDIO)
	Url     string            `json:"url,omitempty"`     // URL for SSE or HTTP type
	Headers map[string]string `json:"headers,omitempty"` // Headers for HTTP requests (SSE or HTTP)
}
