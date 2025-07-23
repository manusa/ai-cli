package test

import (
	"context"
	"github.com/GoogleCloudPlatform/kubectl-ai/gollm"
)

func NewTextResponse(text string) *ChatResponse {
	return &ChatResponse{
		candidates: []gollm.Candidate{
			&Candidate{
				parts: []gollm.Part{
					&Part{text: text},
				},
			},
		},
	}
}

type Client struct {
	StreamingResponse func(ctx context.Context, contents ...any) (gollm.ChatResponseIterator, error)
}

var _ gollm.Client = &Client{}

func (t *Client) Close() error {
	return nil
}

func (t *Client) StartChat(systemPrompt, model string) gollm.Chat {
	return &Chat{tc: t}
}

func (t *Client) GenerateCompletion(ctx context.Context, req *gollm.CompletionRequest) (gollm.CompletionResponse, error) {
	panic("not implemented")
}

func (t *Client) SetResponseSchema(schema *gollm.Schema) error {
	panic("not implemented")
}

func (t *Client) ListModels(ctx context.Context) ([]string, error) {
	panic("not implemented")
}

type Chat struct {
	tc *Client
}

var _ gollm.Chat = &Chat{}

func (t *Chat) Send(_ context.Context, _ ...any) (gollm.ChatResponse, error) {
	panic("not implemented")
}

func (t *Chat) SendStreaming(ctx context.Context, contents ...any) (gollm.ChatResponseIterator, error) {
	if t.tc.StreamingResponse != nil {
		return t.tc.StreamingResponse(ctx, contents)
	}
	return func(yield func(gollm.ChatResponse, error) bool) {
		if !yield(NewTextResponse("AI is not running, this is a test"), nil) {
			return
		}
	}, nil
}

func (t *Chat) SetFunctionDefinitions(_ []*gollm.FunctionDefinition) error {
	panic("not implemented")
}

func (t *Chat) IsRetryableError(_ error) bool {
	panic("not implemented")
}

type ChatResponse struct {
	usageMetadata any
	candidates    []gollm.Candidate
}

var _ gollm.ChatResponse = &ChatResponse{}

func (t *ChatResponse) UsageMetadata() any {
	return t.usageMetadata
}

func (t *ChatResponse) Candidates() []gollm.Candidate {
	return t.candidates
}

type Candidate struct {
	parts []gollm.Part
}

var _ gollm.Candidate = &Candidate{}

func (t *Candidate) String() string {
	panic("not implemented")
}

func (t *Candidate) Parts() []gollm.Part {
	return t.parts
}

type Part struct {
	text string
}

var _ gollm.Part = &Part{}

func (t *Part) AsText() (string, bool) {
	return t.text, true
}

func (t *Part) AsFunctionCalls() ([]gollm.FunctionCall, bool) {
	return nil, false
}
