package setup

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/v2/list"
	"github.com/charmbracelet/log"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/features"
	"github.com/manusa/ai-cli/pkg/ui/components/selector"
)

func Run(ctx context.Context) error {
	for { // setup inference
		discoveredFeatures := features.Discover(ctx)
		if discoveredFeatures.Inference == nil {
			err := selectInferenceProvider(ctx)
			if err != nil {
				return err
			}
		} else {
			inference := *discoveredFeatures.Inference
			fmt.Printf("✅ The %q inference provider will be used\n", inference.Attributes().Name())
			model, err := inference.GetModel(ctx)
			if err != nil {
				return err
			} else {
				fmt.Printf("✅ The model %q will be used\n", model)
				break
			}
		}
	}

	for {
		discoveredFeatures := features.Discover(ctx)
		if len(discoveredFeatures.ToolsNotAvailable) > 0 {
			stop, err := selectToolProvider(ctx)
			if err != nil {
				return err
			}
			if stop {
				displayTools(discoveredFeatures.Tools)
				break
			}
		} else {
			displayTools(discoveredFeatures.Tools)
			break
		}
	}
	return nil
}

func ClearAll(ctx context.Context) error {
	discoveredFeatures := features.Discover(ctx)
	for _, inference := range append(discoveredFeatures.InferencesNotAvailable, discoveredFeatures.Inferences...) {
		if inference.Attributes().SupportsSetup() {
			done, err := inference.Clear(ctx)
			if err != nil {
				return err
			}
			if done {
				fmt.Printf("✅ The setup for %q inference provider has been cleared\n", inference.Attributes().Name())
			}
		}
	}
	for _, tool := range append(discoveredFeatures.ToolsNotAvailable, discoveredFeatures.Tools...) {
		if tool.Attributes().SupportsSetup() {
			done, err := tool.Clear(ctx)
			if err != nil {
				return err
			}
			if done {
				fmt.Printf("✅ The setup for %q tools provider has been cleared\n", tool.Attributes().Name())
			}
		}
	}
	return nil
}

func selectInferenceProvider(ctx context.Context) error {
	discoveredFeatures := features.Discover(ctx)
	inferenceNames := []list.Item{}
	for _, inference := range discoveredFeatures.InferencesNotAvailable {
		if inference.Attributes().SupportsSetup() {
			inferenceNames = append(inferenceNames, selector.Item(inference.Attributes().Name()))
		}
	}

	inference, err := selector.Select("No inference detected, please select below the inference you may want to use:", inferenceNames)
	if err != nil {
		return err
	}
	log.Infof("inference selected by user: %v", inference)

	for _, notAvailableInference := range discoveredFeatures.InferencesNotAvailable {
		if notAvailableInference.Attributes().Name() == inference {
			err = notAvailableInference.InstallHelp()
			if err != nil {
				return err
			}
			break
		}
	}

	return nil
}

func selectToolProvider(ctx context.Context) (stop bool, err error) {
	discoveredFeatures := features.Discover(ctx)
	toolNames := []list.Item{}
	for _, tool := range discoveredFeatures.ToolsNotAvailable {
		if tool.Attributes().SupportsSetup() {
			toolNames = append(toolNames, selector.Item(tool.Attributes().Name()))
		}
	}
	if len(toolNames) == 0 {
		return true, nil
	}
	toolNames = append(toolNames, selector.Item("Terminate setup"))
	tool, err := selector.Select("Some tools can be setup, please select below the tool you may want to use:", toolNames)
	if err != nil {
		return false, err
	}
	if tool == "Terminate setup" {
		return true, nil
	}
	log.Infof("tool selected by user: %v", tool)

	for _, notAvailableTool := range discoveredFeatures.ToolsNotAvailable {
		if notAvailableTool.Attributes().Name() == tool {
			err = notAvailableTool.InstallHelp()
			if err != nil {
				return false, err
			}
			break
		}
	}

	return false, nil
}

func displayTools(tools []api.ToolsProvider) {
	fmt.Printf("✅ The following tools providers will be used:\n")
	for _, tool := range tools {
		supportsSetup := tool.Attributes().SupportsSetup()
		setup := "(no setup needed)"
		if supportsSetup {
			setup = "(setup done)"
		}
		fmt.Printf("  - %s %s\n", tool.Attributes().Name(), setup)
	}
}
