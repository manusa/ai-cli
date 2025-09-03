package config

import (
	"context"

	"github.com/manusa/ai-cli/pkg/api"
)

type configCtxKeyType struct{}

var configCtxKey = configCtxKeyType{}

func WithConfig(ctx context.Context, config *api.Config) context.Context {
	return context.WithValue(ctx, configCtxKey, config)
}

func GetConfig(ctx context.Context) *api.Config {
	config, ok := ctx.Value(configCtxKey).(*api.Config)
	if !ok {
		return nil
	}
	return config
}
