package setup

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/log"

	"github.com/manusa/ai-cli/pkg/features"
)

func Run(ctx context.Context) error {
	for {
		discoveredFeatures := features.Discover(ctx)
		if discoveredFeatures.Inference == nil {
			needRestart, err := selectInferenceProvider(ctx)
			if err != nil {
				return err
			}
			if needRestart {
				break
			}
			choices := []list.Item{
				item("Retry"),
				item("Quit"),
			}
			choice, err := Select("", choices)
			if err != nil {
				return err
			}
			if choice == "Quit" {
				return fmt.Errorf("user chose to quit")
			}
		} else {
			inference := *discoveredFeatures.Inference
			fmt.Printf("✅ The %q inference provider will be used\n", inference.Attributes().Name())
			model, err := inference.GetModel(ctx)
			if err != nil {
				fmt.Printf("\n%s\n", inference.InstallModelHelp())
				choices := []list.Item{
					item("Retry"),
					item("Quit"),
				}
				choice, err := Select("", choices)
				if err != nil {
					return err
				}
				if choice == "Quit" {
					return fmt.Errorf("user chose to quit")
				}
			} else {
				fmt.Printf("✅ The model %q will be used\n", model)
				break
			}
		}
	}
	return nil
}

func selectInferenceProvider(ctx context.Context) (needRestart bool, err error) {
	discoveredFeatures := features.Discover(ctx)
	inferenceNames := []list.Item{}
	for _, inference := range discoveredFeatures.InferencesNotAvailable {
		inferenceNames = append(inferenceNames, item(inference.Attributes().Name()))
	}

	inference, err := Select("No inference detected, please select below the inference you may want to use:", inferenceNames)
	if err != nil {
		return false, err
	}
	log.Infof("inference selected by user: %v", inference)

	for _, notAvailableInference := range discoveredFeatures.InferencesNotAvailable {
		if notAvailableInference.Attributes().Name() == inference {
			var helpMsg string
			helpMsg, needRestart = notAvailableInference.InstallHelp()
			fmt.Printf("\n%s\n", helpMsg)
			break
		}
	}

	return needRestart, nil
}
