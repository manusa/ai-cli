package api

import "github.com/getkin/kin-openapi/openapi3"

type Tool struct {
	Name        string
	Description string
	// Parameters in OpenAPI format
	ParametersSchema *openapi3.Schema
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
