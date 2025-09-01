package api

type MessageType string

const (
	MessageTypeSystem    MessageType = "system"
	MessageTypeUser      MessageType = "user"
	MessageTypeAssistant MessageType = "assistant"
	MessageTypeTool      MessageType = "tool"
	MessageTypeError     MessageType = "error"
)

type Message struct {
	Type MessageType
	Text string

	// ToolMessage specific fields
	ToolName string
}

func NewSystemMessage(text string) Message {
	return Message{Type: MessageTypeSystem, Text: text}
}

func NewAssistantMessage(text string) Message {
	return Message{Type: MessageTypeAssistant, Text: text}
}

func NewUserMessage(text string) Message {
	return Message{Type: MessageTypeUser, Text: text}
}

func NewErrorMessage(text string) Message {
	return Message{Type: MessageTypeError, Text: text}
}

func NewToolMessage(text, toolName string) Message {
	return Message{
		Type:     MessageTypeTool,
		Text:     text,
		ToolName: toolName,
	}
}

func (m *Message) Role() string {
	return string(m.Type)
}
