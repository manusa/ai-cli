package config

import "context"

type configCtxKeyType struct{}

var configCtxKey = configCtxKeyType{}

func WithConfig(ctx context.Context, config *Config) context.Context {
	return context.WithValue(ctx, configCtxKey, config)
}

func GetConfig(ctx context.Context) *Config {
	config, ok := ctx.Value(configCtxKey).(*Config)
	if !ok {
		return nil
	}
	return config
}
