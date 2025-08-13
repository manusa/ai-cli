package config

import (
	"os"

	"github.com/spf13/afero"
)

// FileSystem TODO: Properly inject the file system in the future (see _discover/registration)
var FileSystem afero.Fs = afero.NewOsFs()

type Config struct {
	GoogleApiKey string  // TODO: will likely be removed
	GeminiModel  string  // TODO: will likely be removed
	Inference    *string // An inference to use, if not set, the best inference will be used
	Model        *string // A model to use, if not set, the best model will be used
}

func New() *Config {
	return &Config{
		GoogleApiKey: os.Getenv("GEMINI_API_KEY"),
		GeminiModel:  "gemini-2.0-flash",
	}
}
