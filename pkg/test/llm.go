package test

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type ChatModel struct {
	StreamReader  func(input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error)
	WithToolsFunc func(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error)
}

var _ model.ToolCallingChatModel = &ChatModel{}

func (c *ChatModel) Generate(_ context.Context, _ []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	panic("not implemented")
}

func (c *ChatModel) Stream(_ context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	if c.StreamReader != nil {
		return c.StreamReader(input, opts...)
	}
	return schema.StreamReaderFromArray([]*schema.Message{schema.AssistantMessage("AI is not running, this is a test", nil)}), nil
}

func (c *ChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	if c.WithToolsFunc != nil {
		return c.WithToolsFunc(tools)
	}
	return c, nil
}
