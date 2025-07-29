package model

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/stretchr/testify/assert"
)

type MyAvailableProvider struct{}

var myAvailableProvider = MyAvailableProvider{}

func (myDistantProvider MyAvailableProvider) Attributes() ModelAttributes {
	return ModelAttributes{
		Name:    "myAvailable",
		Distant: true,
	}
}

func (myProvider MyAvailableProvider) IsAvailable(cfg *config.Config) bool {
	return true
}

func (myProvider MyAvailableProvider) GetModel(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error) {
	return nil, nil
}

type MyNonAvailableProvider struct{}

var myNonAvailableProvider = MyNonAvailableProvider{}

func (myNonAvailableProvider MyNonAvailableProvider) Attributes() ModelAttributes {
	return ModelAttributes{
		Name:    "myNonAvailable",
		Distant: true,
	}
}

func (myProvider MyNonAvailableProvider) IsAvailable(cfg *config.Config) bool {
	return false
}

func (myProvider MyNonAvailableProvider) GetModel(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error) {
	return nil, nil
}

func TestGetAvailableModels(t *testing.T) {
	cleanup()
	Register(myNonAvailableProvider)
	Register(myAvailableProvider)
	cfg := config.New()
	discovered, err := getAvailableModels(cfg)
	if err != nil {
		t.Fatalf("failed to discover model: %v", err)
	}
	if len(discovered) != 1 {
		t.Fatalf("one model must be discovered")
	}
	modelAttributes := discovered[0]
	if modelAttributes.Name != "myAvailable" {
		t.Fatalf("model is not myAvailable: %v", modelAttributes.Name)
	}
}

func TestRegisterTwice(t *testing.T) {
	cleanup()
	Register(myNonAvailableProvider)
	assert.Panics(t, func() {
		Register(myNonAvailableProvider)
	}, "Registering a model twice should panic")
}

func TestDiscover(t *testing.T) {
	cleanup()
	Register(myAvailableProvider)
	cfg := config.New()
	_, err := Discover(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to discover model: %v", err)
	}
}
