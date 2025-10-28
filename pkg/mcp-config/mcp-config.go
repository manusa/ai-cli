package mcpconfig

import (
	"fmt"
	"path/filepath"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
)

func Save(provider api.MCPConfig, tools []api.ToolsProvider) error {
	content, err := provider.GetConfig(tools)
	if err != nil {
		return err
	}
	configFile := provider.GetFile()
	exists := false
	if _, err := config.FileSystem.Stat(configFile); err == nil {
		exists = true
	}
	err = config.FileSystem.MkdirAll(filepath.Dir(configFile), 0755)
	if err != nil {
		return err
	}
	file, err := config.FileSystem.Create(configFile)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	_, err = file.Write(content)
	if err != nil {
		return err
	}
	if !exists {
		fmt.Printf("MCP config file %s has been created\n", configFile)
	} else {
		fmt.Printf("MCP config file %s has been updated\n", configFile)
	}
	return nil
}
