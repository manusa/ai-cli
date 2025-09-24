package ai

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// DynamicToolCallingChatModel is a wrapper around model.ToolCallingChatModel that allows dynamic tool reloading.
// Allows LLM model mutation to support dynamic tool reloading.
type DynamicToolCallingChatModel struct {
	delegate model.ToolCallingChatModel
}

var _ model.ToolCallingChatModel = (*DynamicToolCallingChatModel)(nil)

func NewDynamicToolCallingChatModel(base model.ToolCallingChatModel, err error) (*DynamicToolCallingChatModel, error) {
	if err != nil {
		return nil, err
	}
	return &DynamicToolCallingChatModel{delegate: base}, nil
}

func (m *DynamicToolCallingChatModel) ReloadTools(ctx context.Context, tools []tool.BaseTool) (err error) {
	infos := make([]*schema.ToolInfo, 0, len(tools))
	for _, t := range tools {
		info, e := t.Info(ctx)
		if e != nil {
			continue
		}
		infos = append(infos, info)
	}
	m.delegate, err = m.delegate.WithTools(infos)
	return err
}

func (m *DynamicToolCallingChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return m.delegate.Generate(ctx, input, opts...)
}

func (m *DynamicToolCallingChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return m.delegate.Stream(ctx, input, opts...)
}

func (m *DynamicToolCallingChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	m.delegate, _ = m.delegate.WithTools(tools)
	return m, nil
}
