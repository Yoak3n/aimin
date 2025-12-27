package state

type ResultStatus int

const (
	Default ResultStatus = iota
	Running
	Interrupted
	ToReturn
	Returned
)

type CtxStatus int

const (
	TaskState CtxStatus = iota
	IdleState
)
