package config

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/afero"
)

// FileSystem TODO: Properly inject the file system in the future (see _discover/registration)
var FileSystem afero.Fs = afero.NewOsFs()

// Os is the operating system
var Os = runtime.GOOS
var LookPath = exec.LookPath

// CommandExists checks if a command exists in the system's PATH
func CommandExists(cmd string) bool {
	_, err := LookPath(cmd)
	return err == nil
}

func IsDesktop() bool {
	switch system := Os; system {
	case "darwin", "windows":
		return true
	case "linux":
		if v := os.Getenv("DISPLAY"); v != "" {
			return true
		}
		return false
	default:
		return false
	}
}

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
