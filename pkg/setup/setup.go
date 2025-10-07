package setup

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/v2/list"
	"github.com/charmbracelet/log"

	"github.com/manusa/ai-cli/pkg/features"
	"github.com/manusa/ai-cli/pkg/ui/components/selector"
)

func Run(ctx context.Context) error {
	for {
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
