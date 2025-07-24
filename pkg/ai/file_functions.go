package ai

//type Function interface {
//	// Name returns the name of the function.
//	Name() string
//}
//
//type FileList struct {
//	gollm.FunctionDefinition
//}
//
//func (f *FileList) Name() string {
//	return f.FunctionDefinition.Name
//}
//
//var _ Function = &FileList{}
//
//func NewFileList() *FileList {
//	return &FileList{
//		FunctionDefinition: gollm.FunctionDefinition{
//			Name:        "file_list",
//			Description: "List files in the current directory",
//			Parameters: &gollm.Schema{
//				Type: gollm.TypeObject,
//				Properties: map[string]*gollm.Schema{
//					"path": {
//						Type:        gollm.TypeString,
//						Description: "The path to the directory to list files from. Optional, if not provided, the current working directory will be used.",
//					},
//				},
//			},
//		},
//	}
//}
