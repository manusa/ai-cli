package mcpconfig

import (
	"fmt"
	"os"
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
	if _, err := config.FileSystem.Stat(configFile); err == nil {
		// file exists, output config to stdout
		// and message to stderr
		fmt.Fprintf(os.Stderr, "MCP config file %s already exists, outputting config to stdout\n", configFile)
		_, err := fmt.Println(string(content))
		return err
	} else {
		// file does not exist, create it
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
		fmt.Printf("MCP config file %s has been created\n", configFile)
	}
	return nil
}
