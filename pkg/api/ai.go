package api

type Ai interface {
	InferenceAttributes() InferenceAttributes
	ToolCount() int
	Reset()
	Session() Session
	Input() chan Message
}

type Session interface {
	HasMessages() bool
	Messages() []Message
	SystemPrompt() Message
	IsRunning() bool
}
