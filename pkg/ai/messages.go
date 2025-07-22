package ai

type MessageType string

const (
	MessageTypeSystem    MessageType = "system"
	MessageTypeUser      MessageType = "user"
	MessageTypeAssistant MessageType = "assistant"
	MessageTypeTool      MessageType = "tool"
)

type Message struct {
	Type MessageType
	Text string
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

func (m *Message) Role() string {
	return string(m.Type)
}
