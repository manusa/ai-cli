package api

type Ai interface {
	InferenceAttributes() InferenceAttributes
	ToolEnabledCount() int
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
