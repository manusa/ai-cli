package setup

import (
	"context"
	"errors"
	"testing"

	"github.com/charmbracelet/bubbles/v2/list"
	"github.com/manusa/ai-cli/internal/test"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/manusa/ai-cli/pkg/ui/components/selector"
	"github.com/stretchr/testify/suite"
)

type SetupTestSuite struct {
	suite.Suite
}

func (s *SetupTestSuite) SetupTest() {
	inference.Clear()
	tools.Clear()
}

func (s *SetupTestSuite) TestSetupSelectInferenceWithSupportsSetup() {
	s.Run("Only inference providers with SupportsSetup are displayed to the user", func() {
		cfg := config.New()
		ctx := config.WithConfig(context.Background(), cfg)
		inference.Register(test.NewInferenceProvider(
			"inference1",
		))
		inference.Register(test.NewInferenceProvider(
			"inference2",
			test.WithSupportsSetup(),
		))
		selectCall := 0
		itemsCall := []list.Item{}
		selector.MockInit(func(title string, items []list.Item) (string, error) {
			selectCall++
			switch selectCall {
			case 1:
				itemsCall = items
				return "inference2", nil
			case 2:
				return "", errors.New("no inference selected")
			default:
				s.T().Fatalf("expected 2 select calls, got %d", selectCall)
				return "", errors.New("expected 2 select calls")
			}
		})
		_ = Run(ctx)
		if len(itemsCall) != 1 {
			s.T().Fatalf("expected 1 item to be called, got %d", len(itemsCall))
		}
		str, ok := itemsCall[0].(selector.Item)
		if !ok {
			s.T().Fatalf("expected item to be string")
		}
		if str != "inference2" {
			s.T().Fatalf("expected inference2 to be displayed, got %s", str)
		}
	})
}

func (s *SetupTestSuite) TestSetupStopsWhenInferenceIsAvailable() {
	s.Run("When one inference provider becomes available, the setup of inference providers stops", func() {
		cfg := config.New()
		ctx := config.WithConfig(context.Background(), cfg)
		var inferenceProvider *test.InferenceProvider
		inferenceProvider = test.NewInferenceProvider(
			"inference1",
			test.WithSupportsSetup(),
			test.WithInstallHelp(func() error {
				// the install help makes the inference provider available
				inferenceProvider.Available = true
				return nil
			}),
		)
		otherInferenceProvider := test.NewInferenceProvider(
			"inference2",
			test.WithSupportsSetup(),
		)
		selectCall := 0
		selector.MockInit(func(title string, items []list.Item) (string, error) {
			selectCall++
			if selectCall == 1 {
				return "inference1", nil
			}
			// protect from infinite loop in case of test failure
			s.T().Fatalf("expected 1 select call, got %d", selectCall)
			return "", errors.New("expected 1 select call")
		})
		inference.Register(inferenceProvider)
		inference.Register(otherInferenceProvider)

		_ = Run(ctx)
		if !inferenceProvider.Available {
			s.T().Fatalf("expected inference provider inference1 to be available")
		}
		if selectCall != 1 {
			s.T().Fatalf("expected 1 select call, got %d", selectCall)
		}
	})
}

func (s *SetupTestSuite) TestSetupContinueWithToolSetup() {
	s.Run("When one inference provider is available with a model, the setup continues with tool providers setup", func() {
		cfg := config.New()
		ctx := config.WithConfig(context.Background(), cfg)
		var inferenceProvider *test.InferenceProvider
		inferenceProvider = test.NewInferenceProvider(
			"inference1",
			test.WithSupportsSetup(),
			test.WithGetModel(func() (string, error) {
				return "model1", nil
			}),
			test.WithInstallHelp(func() error {
				// the install help makes the inference provider available
				inferenceProvider.Available = true
				return nil
			}),
		)
		selectCall := 0
		toolsSelectCall := []list.Item{}
		selector.MockInit(func(title string, items []list.Item) (string, error) {
			selectCall++
			switch selectCall {
			case 1:
				return "inference1", nil
			case 2:
				toolsSelectCall = items
				return "tools1", nil
			case 3:
				return "", errors.New("cancel tool setup")
			default:
				s.T().Fatalf("expected 3 select calls, got %d", selectCall)
				return "", errors.New("expected 3 select calls")
			}
		})
		inference.Register(inferenceProvider)

		tools.Register(test.NewToolsProvider(
			"tools1",
			test.WithToolsSupportsSetup(),
		))
		tools.Register(test.NewToolsProvider(
			"tools2",
		))

		_ = Run(ctx)

		if selectCall != 3 {
			s.T().Fatalf("expected 3 select calls, got %d", selectCall)
		}

		if len(toolsSelectCall) != 2 {
			s.T().Fatalf("expected 1 tool to be displayed plus terminate setup, got %d", len(toolsSelectCall))
		}
		str, ok := toolsSelectCall[0].(selector.Item)
		if !ok {
			s.T().Fatalf("expected item to be string")
		}
		if str != "tools1" {
			s.T().Fatalf("expected tools1 to be displayed, got %s", str)
		}
	})
}

func (s *SetupTestSuite) TestSetupStopsWhenToolSetupIsTerminated() {
	s.Run("When one inference provider is available with a model, the setup continues with tool providers setup", func() {
		cfg := config.New()
		ctx := config.WithConfig(context.Background(), cfg)
		var inferenceProvider *test.InferenceProvider
		inferenceProvider = test.NewInferenceProvider(
			"inference1",
			test.WithSupportsSetup(),
			test.WithGetModel(func() (string, error) {
				return "model1", nil
			}),
			test.WithInstallHelp(func() error {
				// the install help makes the inference provider available
				inferenceProvider.Available = true
				return nil
			}),
		)
		selectCall := 0
		selector.MockInit(func(title string, items []list.Item) (string, error) {
			selectCall++
			switch selectCall {
			case 1:
				return "inference1", nil
			case 2:
				return "Terminate setup", nil
			default:
				s.T().Fatalf("expected 2 select calls, got %d", selectCall)
				return "", errors.New("expected 2 select calls")
			}
		})
		inference.Register(inferenceProvider)

		tools.Register(test.NewToolsProvider(
			"tools1",
			test.WithToolsSupportsSetup(),
		))
		tools.Register(test.NewToolsProvider(
			"tools2",
		))

		_ = Run(ctx)

		if selectCall != 2 {
			s.T().Fatalf("expected 2 select calls, got %d", selectCall)
		}
	})
}

func (s *SetupTestSuite) TestSetupStopsWhenAllToolsAreAvailable() {
	s.Run("The tools setup terminates when all tools are available", func() {
		cfg := config.New()
		ctx := config.WithConfig(context.Background(), cfg)
		var inferenceProvider *test.InferenceProvider
		inferenceProvider = test.NewInferenceProvider(
			"inference1",
			test.WithSupportsSetup(),
			test.WithGetModel(func() (string, error) {
				return "model1", nil
			}),
			test.WithInstallHelp(func() error {
				// the install help makes the inference provider available
				inferenceProvider.Available = true
				return nil
			}),
		)
		selectCall := 0
		selector.MockInit(func(title string, items []list.Item) (string, error) {
			selectCall++
			switch selectCall {
			case 1:
				return "inference1", nil
			case 2:
				return "tools1", nil
			case 3:
				return "tools2", nil
			default:
				s.T().Fatalf("expected 3 select calls, got %d", selectCall)
				return "", errors.New("expected 3 select calls")
			}
		})
		inference.Register(inferenceProvider)

		var tools1 *test.ToolsProvider
		tools1 = test.NewToolsProvider(
			"tools1",
			test.WithToolsSupportsSetup(),
			test.WithToolsInstallHelp(func() error {
				// the install help makes the tools provider available
				tools1.Available = true
				return nil
			}),
		)
		var tools2 *test.ToolsProvider
		tools2 = test.NewToolsProvider(
			"tools2",
			test.WithToolsSupportsSetup(),
			test.WithToolsInstallHelp(func() error {
				// the install help makes the tools provider available
				tools2.Available = true
				return nil
			}),
		)

		tools.Register(tools1)
		tools.Register(tools2)

		_ = Run(ctx)

		if selectCall != 3 {
			s.T().Fatalf("expected 3 select calls, got %d", selectCall)
		}
	})
}

func TestSetup(t *testing.T) {
	suite.Run(t, new(SetupTestSuite))
}
