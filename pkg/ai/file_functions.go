package ai

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"os"
)

type fileList struct {
	schema.ToolInfo
}

var _ tool.InvokableTool = &fileList{}

var FileList = &fileList{
	ToolInfo: schema.ToolInfo{
		Name: "file_list",
		Desc: "List files in the provided directory or the current working directory if none is provided." +
			"Returns a JSON representation of the files, including their names and metadata.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"directory": {
				Type:     schema.String,
				Desc:     "The directory to list files from. If not provided, the current working directory will be used.",
				Required: false,
			},
		}),
	},
}

func (f fileList) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &f.ToolInfo, nil
}

func (f fileList) InvokableRun(_ context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	directory := "."
	// Parse the arguments from JSON
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err == nil {
		d, ok := args["directory"].(string)
		if ok && d != "" {
			directory = d
		}
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
}
