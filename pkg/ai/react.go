package ai

import (
	"context"
	"reflect"

	"github.com/charmbracelet/log"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	callbackutils "github.com/cloudwego/eino/utils/callbacks"
	"github.com/manusa/ai-cli/pkg/api"
)

const (
	DefaultMaxSteps = 50
)

// ReActAgent is an agent that uses the ReAct framework to interact with the user and tools.
// Currently, it's just wrapping around the standard react.Agent, but we might eventually want to fully implement
// our own agent to have more control over the process and graph chain.
type ReActAgent struct {
	*react.Agent
	ai *Ai
}

func NewReActAgent(ctx context.Context, ai *Ai) (agent *ReActAgent, err error) {
	agent = &ReActAgent{ai: ai}
	agent.Agent, err = react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: ai.llm,
		MaxStep:          DefaultMaxSteps,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools:               ai.tools.EnabledTools(),
			ExecuteSequentially: true,
			UnknownToolsHandler: agent.unknownToolHandler,
		},
	})
	if err != nil {
		return
	}
	return
}

// unknownToolHandler is called when the model tries to call a tool that is not in the list of available tools.
// This is a workaround because graph tool declarations are immutable after the graph is created and compiled.
// The model might be actually calling a tool that was enabled after the graph was created (hence not unknown).
//
// The only issue is that standard callbacks are not called for unknown tools, so we manually call them here.
func (r *ReActAgent) unknownToolHandler(ctx context.Context, name, input string) (string, error) {
	r.OnToolCallStart(ctx, &callbacks.RunInfo{Name: name}, &tool.CallbackInput{ArgumentsInJSON: input})
	defer r.OnToolCallEnd(ctx, &callbacks.RunInfo{Name: name}, &tool.CallbackOutput{})
	return r.ai.tools.InvokeTool(ctx, name, input)
}

func (r *ReActAgent) OnChatModelStart(ctx context.Context, runInfo *callbacks.RunInfo, input *model.CallbackInput) context.Context {
	// Reload the tools if the model is a DynamicToolCallingChatModel
	// Enables dynamic tool reloading by replacing the delegate (immutable) model inside DynamicToolCallingChatModel
	if runInfo.Type != reflect.TypeOf(DynamicToolCallingChatModel{}).Name() {
		return ctx
	}
	_ = r.ai.llm.ReloadTools(ctx, r.ai.tools.EnabledTools())
	return ctx
}

func (r *ReActAgent) OnToolCallStart(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
	log.Debug("calling tool", "name", info.Name, "input", input.ArgumentsInJSON)
	return ctx
}

func (r *ReActAgent) OnToolCallEnd(ctx context.Context, info *callbacks.RunInfo, output *tool.CallbackOutput) context.Context {
	log.Debug("called tool", "name", info.Name, "response", output.Response)
	r.ai.appendMessage(api.NewToolMessage(output.Response, info.Name))
	return ctx
}

func (r *ReActAgent) Stream(ctx context.Context) (*schema.StreamReader[*schema.Message], error) {
	return r.Agent.Stream(
		ctx,
		r.ai.schemaMessages(),
		agent.WithComposeOptions(compose.WithCallbacks(
			callbackutils.NewHandlerHelper().ChatModel(&callbackutils.ModelCallbackHandler{
				OnStart: r.OnChatModelStart,
			}).Handler(),
			callbackutils.NewHandlerHelper().Tool(&callbackutils.ToolCallbackHandler{
				OnStart: r.OnToolCallStart,
				OnEnd:   r.OnToolCallEnd,
			}).Handler(),
		)))
}
