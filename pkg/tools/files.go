package tools

import (
	"encoding/json"
	"github.com/manusa/ai-cli/pkg/api"
	"os"
)

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
