package api

type Tool struct {
	Name        string
	Description string
	Parameters  map[string]ToolParameter
	Function    func(args map[string]interface{}) (string, error)
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
