package state

import (
	"context"
	"time"
)

type Manager struct {
	Save        SaveData
	Ctx         *Context
	current     State
	start       State
	interruptCh chan Task
	taskQueue   chan Task
}

func NewStateManager(root State) *Manager {
	return &Manager{
		Save:      nil,
		Ctx:       nil,
		taskQueue: make(chan Task, 100),
		start:     root,
	}
}

func (m *Manager) Run(ctx context.Context) error {
	cwc, cancel := context.WithCancel(ctx)
	m.current = m.start
	m.Ctx = NewStateContext(cwc)
	defer cancel()
	for {
		select {
		case <-cwc.Done():
			return nil
		default:
			// 正常执行状态机逻辑
			if m.current != nil {
				m.Ctx.Caller = m.current
				result, err := m.current.Execute(m.Ctx)
				if err != nil {
					return err
				}
				switch result.GetStatus() {
				case Interrupted:
					// 处理中断逻辑
					m.waitForResume(m.Ctx)
				case ToReturn:
					// 处理回归逻辑
					m.Ctx.ReturnTo = m.current
					m.current = result.GetNextState()
				case Returned:
					m.Ctx.Data[result.Caller().Type()] = result.GetData()
					m.current = result.GetNextState()
				default:
					// 正常完成，进入下一个状态
					m.current = result.GetNextState()
				}
			} else {
				m.current = m.start
			}
		}

		time.Sleep(time.Microsecond * 100)
	}
}

func (m *Manager) SetStartState(state State) {
	m.start = state
}

func (m *Manager) waitForResume(ctx *Context) {
	resumeState := <-ctx.resume
	// 根据 resumeFlag 处理恢复逻辑
	if m.Save.Type() == resumeState.Type() {

	}
	m.current = resumeState
}
