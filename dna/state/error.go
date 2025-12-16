package state

var ErrInterrupted = InterruptError{msg: "interrupted"}

type InterruptError struct {
	msg string
}

func (ie InterruptError) Error() string {
	return ie.msg
}
