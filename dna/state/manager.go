package state

import (
	"context"
	"dna/pkg/logger"
	"dna/pkg/util"
	"fmt"
	"sync"
	"time"
)

type StackStateManager struct {
	start       State
	current     State
	currentID   string
	storage     StateStorage
	taskQueue   chan Task
	interruptCh chan Task // 中断通道
	interrupted State
	resumeCh    chan bool // 恢复通道
	cancel      context.CancelFunc
	ctx         *Context
	mu          sync.RWMutex
}

func NewStackStateManager(ctx context.Context, initial State) *StackStateManager {
	sm := &StackStateManager{
		start:       initial,
		current:     initial,
		currentID:   "",
		taskQueue:   make(chan Task, 100),
		interruptCh: make(chan Task, 10),
		resumeCh:    make(chan bool, 10),
		storage:     NewInMemoryStateStorage(),
		mu:          sync.RWMutex{},
	}
	cwc, cancel := context.WithCancel(ctx)
	sm.cancel = cancel
	stateCtx := NewStateContext(cwc)
	sm.ctx = stateCtx
	return sm
}

func (sm *StackStateManager) Run() error {
	// 初始化上下文
	defer sm.cancel()
	ctx := sm.ctx.Context
	stateCtx := sm.ctx
	go sm.taskListener(ctx)

	// for sm.current != nil {
	// 	// 创建当前状态的上下文
	// 	stateCtx.Caller = sm.current
	// 	// 执行当前状态
	// 	result, err := sm.current.Execute(stateCtx)
	// 	if err != nil {
	// 		// 需要兼容错误处理，让状态树继续运行
	// 		return fmt.Errorf("state %d execution failed: %w", sm.current.Type(), err)
	// 	}

	// 	// TODO 保存结果

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case task := <-sm.interruptCh:
			// 处理中断请求
			stateCtx.Interrupt <- struct{}{}
			if err := sm.handleInterrupt(task); err != nil {
				logger.Logger.Errorf("Handle interrupt error: %v\n", err)
			}
		default:
			// 正常执行当前状态
			if sm.current != nil {
				// 创建执行上下文
				stateCtx.Caller = sm.current
				sm.currentID = util.GenEventID(sm.current.Label())
				// 执行当前状态
				result, err := sm.current.Execute(stateCtx)
				if err != nil {
					fmt.Printf("State %d execution error: %v\n",
						sm.current.Type(), err)
					// 错误处理：可以进入错误恢复状态
					continue
				}

				// 处理状态转移
				if result.IsReturn {
					if result.NextState != nil {
						stateCtx.ReturnTo = nil
						sm.current = result.NextState
					}
				} else if result.NextState != nil {
					// 如果需要保存当前状态
					//if result.ShouldSave {
					//	sm.saveCurrentState()
					//}
					// 更新当前状态
					sm.current = result.NextState
				} else if result.IsComplete {
					// TODO 中断了原来的状态，接下来进入高优先级任务队列
					// 等待恢复信号
					<-sm.resumeCh
					// 处理恢复请求
					if err := sm.handleResume(); err != nil {
						fmt.Printf("Handle resume error: %v\n", err)
					}
				} else {
					sm.current = sm.start
				}
			}

			// 短暂休眠避免CPU忙等待
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (sm *StackStateManager) SubmitTask(task Task) error {
	select {
	case sm.interruptCh <- task:
		return nil
	default:
		return fmt.Errorf("interrupt channel full")
	}
}

func (sm *StackStateManager) queueTask(task Task) {
	// 实现任务排队逻辑
	// 可以在适当的时机（如当前状态完成时）检查队列并执行
}

func (sm *StackStateManager) taskListener(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-sm.taskQueue:
			// 根据任务优先级决定是否立即中断
			if task.Priority > 5 { // 高优先级任务立即中断
				sm.SubmitTask(task)
			} else {
				// 低优先级任务排队等待
				sm.queueTask(task)
			}
		}
	}
}

// handleInterrupt 处理中断
func (sm *StackStateManager) handleInterrupt(task Task) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	fmt.Printf("Interrupting state %d for task %s\n",
		sm.current.Type(), task.ID)
	sm.ctx.Interrupt <- struct{}{}
	defer func() {
		sm.resumeCh <- true
	}()
	// 保存当前状态
	if err := sm.saveCurrentState(); err != nil {
		return fmt.Errorf("save current state error: %w", err)
	}

	// 切换到工作状态（处理任务）
	// workState := NewTaskWorkState(task)

	// // 将工作状态压栈
	// ism.stateStack.Push(&SavedState{
	// 	StateName: sm.currentState.Type(),
	// 	StateData: sm.serializeCurrentState(),
	// 	Timestamp: time.Now().Unix(),
	// 	Metadata: map[string]interface{}{
	// 		"interrupted_by": task.ID,
	// 	},
	// })

	// 切换到正常状态
	return nil
}

func (sm *StackStateManager) handleResume() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 从栈中恢复之前的状态
	// savedState, ok := sm.stateStack.Pop()
	// if !ok {
	// 	return fmt.Errorf("no saved state to resume")
	// }

	// 加载之前的状态
	// 这里需要根据 savedState.StateName 创建相应的状态实例
	// 然后调用 LoadState(savedState.StateData)
	//sm.storage.Load("some_key")
	// fmt.Printf("Resuming to state %s\n", savedState.StateName)
	sm.current = sm.interrupted
	return nil
}

// saveCurrentState 保存当前状态
func (sm *StackStateManager) saveCurrentState() error {
	if sm.current == nil {
		return nil
	}

	//stateData, err := sm.current.SaveState()
	//if err != nil {
	//	return err
	////}
	//
	//// 保存到存储
	//key := "interrupt"

	//return sm.storage.Save(key, sm.current.Type())
	sm.interrupted = sm.current
	return nil
}

func (sm *StackStateManager) restorePreviousState() error {
	// 发送恢复信号
	select {
	case sm.resumeCh <- true:
		return nil
	default:
		return fmt.Errorf("resume channel full")
	}
}
