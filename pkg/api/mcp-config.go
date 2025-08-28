package api

// MCPConfig is the interface for MCP config providers (editors supporintg MCP servers configured in a file)
type MCPConfig interface {
	// GetFile returns the path to the MCP config file
	GetFile() string
	// GetConfig returns the MCP config for the given tools in format expected by the editor
	GetConfig(tools []ToolsProvider) ([]byte, error)
}
