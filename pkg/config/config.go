package config

import "os"

type Config struct {
	GoogleApiKey string // TODO: will likely be removed
	GeminiModel  string // TODO: will likely be removed
}

func New() *Config {
	return &Config{
		GoogleApiKey: os.Getenv("GEMINI_API_KEY"),
		GeminiModel:  "gemini-2.0-flash",
	}
}
