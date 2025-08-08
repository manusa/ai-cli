package tools

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/ai"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
)

var providers = map[string]Provider{}

type BasicToolsProvider struct {
	api.BasicFeatureProvider
}

func (p *BasicToolsProvider) SetReason(reason string) {
	p.Reason = reason
}

type Attributes struct {
	api.BasicFeatureAttributes
	*api.ModelAttributes
}

type Data struct {
	api.BasicFeatureData
}

type Report struct {
	Attributes
	Data
}

type Provider interface {
	api.Feature[Attributes, Data]
	GetTools(ctx context.Context, cfg *config.Config) ([]*api.Tool, error)
	MarshalJSON() ([]byte, error)
	SetReason(reason string)
}

// Register a new tools provider
func Register(provider Provider) {
	if _, ok := providers[provider.Attributes().Name()]; ok {
		panic(fmt.Sprintf("tool provider already registered: %s", provider.Attributes().Name()))
	}
	providers[provider.Attributes().Name()] = provider
}

// Clear the registered tools providers (Exposed for testing purposes)
func Clear() {
	providers = map[string]Provider{}
}

// Discover the available tools based on the user preferences
func Discover(cfg *config.Config) (availableTools []Provider, notAvailableTools []Provider) {
	for _, provider := range providers {
		if provider.Attributes().ModelAttributes != nil {
			continue
		}
		if provider.IsAvailable(cfg) {
			availableTools = append(availableTools, provider)
		} else {
			notAvailableTools = append(notAvailableTools, provider)
		}
	}
	slices.SortFunc(availableTools, func(a, b Provider) int {
		return strings.Compare(a.Attributes().Name(), b.Attributes().Name())
	})
	slices.SortFunc(notAvailableTools, func(a, b Provider) int {
		return strings.Compare(a.Attributes().Name(), b.Attributes().Name())
	})
	return availableTools, notAvailableTools
}

func DiscoverWithModel(ctx context.Context, cfg *config.Config, llm model.ToolCallingChatModel, tools []Provider) (availableTools []Provider, notAvailableTools []Provider) {
nextProvider:
	for _, provider := range providers {
		if provider.Attributes().ModelAttributes == nil {
			continue
		}
		var allTools []*api.Tool

		availableToolsNames := []string{}
		for _, toolProvider := range tools {
			availableToolsNames = append(availableToolsNames, toolProvider.Attributes().Name())
		}

		// Check that all needed tools are available
		for _, neededTool := range provider.Attributes().NeededTools {
			if !slices.Contains(availableToolsNames, neededTool) {
				provider.SetReason(fmt.Sprintf("The necessary tool '%s' is not available", neededTool))
				notAvailableTools = append(notAvailableTools, provider)
				continue nextProvider
			}
		}

		for _, toolProvider := range tools {
			if !slices.Contains(provider.Attributes().NeededTools, toolProvider.Attributes().Name()) {
				continue
			}
			tools, err := toolProvider.GetTools(ctx, cfg)
			if err != nil {
				continue
			}
			allTools = append(allTools, tools...)
		}
		aiAgent := ai.New(llm, allTools, cfg)
		if err := aiAgent.Run(ctx); err != nil {
			return nil, nil
		}
		aiAgent.Input <- api.NewUserMessage(provider.Attributes().Prompt)
		for {
			<-aiAgent.Output
			if !aiAgent.Session().IsRunning() {
				break
			}
		}
		messages := aiAgent.Session().Messages()
		last := messages[len(messages)-1].Text
		if strings.TrimSpace(last) == "Yes" {
			provider.SetReason("model replied Yes to the prompt")
			availableTools = append(availableTools, provider)
		} else {
			provider.SetReason(fmt.Sprintf("model response:'%s'", last))
			notAvailableTools = append(notAvailableTools, provider)
		}
	}
	slices.SortFunc(availableTools, func(a, b Provider) int {
		return strings.Compare(a.Attributes().Name(), b.Attributes().Name())
	})
	slices.SortFunc(notAvailableTools, func(a, b Provider) int {
		return strings.Compare(a.Attributes().Name(), b.Attributes().Name())
	})
	return availableTools, notAvailableTools
}
