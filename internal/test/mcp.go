package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/manusa/ai-cli/pkg/api"
)

var mcpSettings *api.McpSettings

func McpServer() *api.McpSettings {
	if mcpSettings != nil {
		return mcpSettings
	}
	projectRoot, err := os.Getwd()
	for len(projectRoot) > 0 && err == nil {
		if _, statErr := os.Stat(filepath.Join(projectRoot, "go.mod")); statErr == nil {
			break
		}
		projectRoot = filepath.Dir(projectRoot)
	}
	if err != nil || len(projectRoot) == 0 {
		panic(fmt.Sprintf("failed to find project root: %s", err))
	}
	binDir := filepath.Join(projectRoot, ".work", "testdata")
	err = os.MkdirAll(binDir, 0755)
	if err != nil {
		panic(fmt.Sprintf("failed to create testdata bin dir: %s", err))
	}
	binary := "test-mcp-server"
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	output, err := exec.
		Command("go", "build", "-o", filepath.Join(binDir, binary),
			filepath.Join(projectRoot, "internal", "test", "mcp", "main.go")).
		CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("failed to build mcp server binary: %s, output: %s", err, string(output)))
	}
	mcpSettings = &api.McpSettings{
		Type:    api.McpTypeStdio,
		Command: filepath.Join(binDir, binary),
		Args:    []string{},
	}
	return mcpSettings
}
