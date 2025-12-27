package state

type Result interface {
	GetData() any
	GetNextState() State
	GetStatus() ResultStatus
	SetCaller(caller State)
	Caller() State
}

// ResultData 执行结果
type ResultData struct {
	Data      any
	From      State
	NextState State
	ToReturn  State
	Status    ResultStatus
}

func (r ResultData) GetData() any {
	return r.Data
}

func (r ResultData) GetNextState() State {
	return r.NextState
}

func (r ResultData) SetCaller(caller State) {
	r.From = caller
}

func (r ResultData) Caller() State {
	return r.From
}

func (r ResultData) GetStatus() ResultStatus {
	return r.Status
}

//func NewResultData(nextState State, status ResultStatus, caller State,data ...any ) *ResultData {
//	return &ResultData{
//		Data:      data,
//		NextState: nextState,
//		Status:    status,
//		From:      caller,
//	}
//}
