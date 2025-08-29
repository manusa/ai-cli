package fs

import (
	"context"
	"encoding/json"
	"os"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/policies"
	"github.com/manusa/ai-cli/pkg/tools"
)

type Provider struct {
	api.BasicToolsProvider
}

var _ api.ToolsProvider = &Provider{}

type FsPolicies struct {
	policies.ToolPolicies
}

func (p *Provider) Initialize(_ *config.Config, toolPolicies any) {
	// TODO: This should probably be generalized to all tools and inference providers
	if !policies.IsEnabledByPolicies(toolPolicies) {
		p.IsAvailableReason = "filesystem is not authorized by policies"
		return
	}
	// ReadOnly is not considered for fs, as all operations are read-only
	p.Available = true
	p.IsAvailableReason = "filesystem is accessible"
}

func (p *Provider) GetTools(_ context.Context, _ *config.Config) ([]*api.Tool, error) {
	return []*api.Tool{
		FileList,
	}, nil
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

func (p *Provider) GetDefaultPolicies() map[string]any {
	var policies = FsPolicies{}
	jsonBody, err := json.Marshal(policies)
	if err != nil {
		return nil
	}
	var policiesMap map[string]any
	err = json.Unmarshal(jsonBody, &policiesMap)
	if err != nil {
		return nil
	}
	return policiesMap
}

var instance = &Provider{
	api.BasicToolsProvider{
		BasicToolsAttributes: api.BasicToolsAttributes{
			BasicFeatureAttributes: api.BasicFeatureAttributes{
				FeatureName:        "fs",
				FeatureDescription: "Provides access to the local filesystem, allowing listing of files and directories.",
			},
		},
	},
}

func init() {
	tools.Register(instance)
}
