package ai

type Session struct {
	systemPrompt      Message
	messages          []Message
	messageInProgress Message
	error             error
	running           bool
}

func (s *Session) HasMessages() bool {
	return len(s.messages) > 0 || s.error != nil || (s.IsRunning() && s.messageInProgress.Text != "")
}

func (s *Session) Messages() []Message {
	ret := make([]Message, len(s.messages))
	copy(ret, s.messages)
	if s.IsRunning() && s.messageInProgress.Text != "" {
		ret = append(ret, s.messageInProgress)
	}
	if s.error != nil {
		ret = append(ret, NewErrorMessage(s.error.Error()))
	}
	return ret
}

func (s *Session) SystemPrompt() Message {
	return s.systemPrompt
}

func (s *Session) IsRunning() bool {
	return s.running
}
