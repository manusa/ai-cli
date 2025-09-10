package main

import (
	"context"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestFunc(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "test-works"},
		},
	}, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "testing-mcp-server", Version: "1.33.7"}, nil)
	server.AddTool(&mcp.Tool{Name: "test-func", Description: "A test tool", InputSchema: &jsonschema.Schema{
		Type: "object",
	}}, TestFunc)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		panic(err)
	}
}
