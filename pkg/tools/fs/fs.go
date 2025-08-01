package fs

import (
	"context"
	"encoding/json"
	"os"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/tools"
)

type Provider struct {
}

var _ tools.Provider = &Provider{}

func (p *Provider) Attributes() tools.Attributes {
	return tools.Attributes{
		BasicFeatureAttributes: api.BasicFeatureAttributes{
			FeatureName: "fs",
		},
	}
}

func (p *Provider) IsAvailable(_ *config.Config) bool {
	return true
}

func (p *Provider) GetTools(_ context.Context, _ *config.Config) ([]*api.Tool, error) {
	return []*api.Tool{
		FileList,
	}, nil
}

func (p *Provider) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Attributes())
}

var FileList = &api.Tool{
	Name: "file_list",
	Description: "List files in the provided directory or the current working directory if none is provided." +
		"Returns a JSON representation of the files, including their names and metadata.",
	Parameters: map[string]api.ToolParameter{
		"directory": {
			Type:        api.String,
			Description: "The directory to list files from. If not provided, the current working directory will be used.",
			Required:    false,
		},
	},
	Function: func(args map[string]interface{}) (string, error) {
		directory := "."
		d, ok := args["directory"].(string)
		if ok && d != "" {
			directory = d
		}
		files, err := os.ReadDir(directory)
		if err != nil {
			return "", err
		}
		var fileInfos []interface{}
		for _, file := range files {
			fileInfo := map[string]interface{}{
				"name": file.Name(),
				"type": file.Type().String(),
			}
			if fi, err := file.Info(); err == nil {
				fileInfo["size"] = fi.Size()
				fileInfo["mod_time"] = fi.ModTime().Format("2006-01-02 15:04:05")
			}
			fileInfos = append(fileInfos, fileInfo)
		}
		fileNamesJSON, err := json.Marshal(fileInfos)
		if err != nil {
			return "", err
		}
		return string(fileNamesJSON), nil
	},
}

var instance = &Provider{}

func init() {
	tools.Register(instance)
}
