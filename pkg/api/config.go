package api

type Config struct {
	GoogleApiKey string  // TODO: will likely be removed
	GeminiModel  string  // TODO: will likely be removed
	Inference    *string // An inference to use, if not set, the best inference will be used
	Model        *string // A model to use, if not set, the best model will be used

	ToolsParameters map[string]ToolParameters
}
