package ai

type Session struct {
	systemPrompt      Message
	messages          []Message
	messageInProgress Message
	running           bool
}

func (s *Session) HasMessages() bool {
	return len(s.messages) > 0
}

func (s *Session) Messages() []Message {
	ret := make([]Message, len(s.messages))
	copy(ret, s.messages)
	if s.IsRunning() && s.messageInProgress.Text != "" {
		ret = append(ret, s.messageInProgress)
	}
	return ret
}

func (s *Session) SystemPrompt() string {
	return s.systemPrompt.Text
}

func (s *Session) IsRunning() bool {
	return s.running
}
