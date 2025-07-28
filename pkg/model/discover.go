package model

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/config"
)

func Discover(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error) {
	if isGeminiAvailable(cfg) {
		return GetGemini(ctx, cfg)
	}
	if isOllamaAvailable() {
		return GetOllama(ctx, cfg)
	}
	return nil, fmt.Errorf("no model found")
}
