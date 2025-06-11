package message

type ErrMsg struct{ Err error }

// For messages that contain errors it's often handy to also implement the
// error interface on the message.
func (e ErrMsg) Error() string { return e.Err.Error() }

// Menubar Message
type BackToMenubarMsg struct{}

// Rank Page Message
// type BackToRankMsg struct{}

// Detail Page Message
type DetailMsg struct {
	ID int
}

func (d *DetailMsg) AnimeID() int {
	return d.ID
}
