package ai

// TODO: Might need to be moved to a separate package
type Session struct {
	Messages []string // TODO
}

func (s *Session) InProgress() bool {
	return false
}
