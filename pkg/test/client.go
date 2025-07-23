package test

import (
	"context"
	"github.com/GoogleCloudPlatform/kubectl-ai/gollm"
)

type TestClient struct{}

var _ gollm.Client = &TestClient{}

func (t TestClient) Close() error {
	return nil
}

func (t TestClient) StartChat(systemPrompt, model string) gollm.Chat {
	return &TestChat{}
}

func (t TestClient) GenerateCompletion(ctx context.Context, req *gollm.CompletionRequest) (gollm.CompletionResponse, error) {
	panic("not implemented")
}

func (t TestClient) SetResponseSchema(schema *gollm.Schema) error {
	panic("not implemented")
}

func (t TestClient) ListModels(ctx context.Context) ([]string, error) {
	panic("not implemented")
}

type TestChat struct {
}

var _ gollm.Chat = &TestChat{}

func (t TestChat) Send(_ context.Context, _ ...any) (gollm.ChatResponse, error) {
	panic("not implemented")
}

func (t TestChat) SendStreaming(ctx context.Context, contents ...any) (gollm.ChatResponseIterator, error) {
	response := &TestChatResponse{
		candidates: []gollm.Candidate{
			TestCandidate{
				parts: []gollm.Part{
					TestPart{text: "AI is not running, this is a test"},
				},
			},
		},
	}
	return func(yield func(gollm.ChatResponse, error) bool) {
		if !yield(response, nil) {
			return
		}
	}, nil
}

func (t TestChat) SetFunctionDefinitions(_ []*gollm.FunctionDefinition) error {
	panic("not implemented")
}

func (t TestChat) IsRetryableError(_ error) bool {
	panic("not implemented")
}

type TestChatResponse struct {
	usageMetadata any
	candidates    []gollm.Candidate
}

var _ gollm.ChatResponse = &TestChatResponse{}

func (t *TestChatResponse) UsageMetadata() any {
	return t.usageMetadata
}

func (t *TestChatResponse) Candidates() []gollm.Candidate {
	return t.candidates
}

type TestCandidate struct {
	parts []gollm.Part
}

var _ gollm.Candidate = &TestCandidate{}

func (t TestCandidate) String() string {
	panic("not implemented")
}

func (t TestCandidate) Parts() []gollm.Part {
	return t.parts
}

type TestPart struct {
	text string
}

var _ gollm.Part = &TestPart{}

func (t TestPart) AsText() (string, bool) {
	return t.text, true
}

func (t TestPart) AsFunctionCalls() ([]gollm.FunctionCall, bool) {
	return nil, false
}
