package test

import (
	"os"
	"os/exec"
	"path"
	"runtime"
	"testing"

	"github.com/manusa/ai-cli/pkg/api"
)

func McpServer(t *testing.T) *api.McpSettings {
	t.Helper()
	binDir := t.TempDir()
	binary := "test-mcp-server"
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	projectRoot, err := os.Getwd()
	for len(projectRoot) > 0 && err == nil {
		if _, err := os.Stat(path.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		projectRoot = path.Dir(projectRoot)
	}
	output, err := exec.
		Command("go", "build", "-o", path.Join(binDir, binary),
			path.Join(projectRoot, "internal", "test", "mcp", "main.go")).
		CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build mcp server binary: %s, output: %s", err, string(output))
	}
	return &api.McpSettings{
		Type:    api.McpTypeStdio,
		Command: path.Join(binDir, binary),
		Args:    []string{},
	}
}
