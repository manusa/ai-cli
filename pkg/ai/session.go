package ai

// TODO: Might need to be moved to a separate package
type Session struct {
	Messages []Message
}

func (s *Session) InProgress() bool {
	return false
}
