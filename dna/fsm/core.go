package fsm

import (
	"fmt"
)

// StateType 定义状态类型
type StateType int

const (
	VirtualStateType StateType = iota // 虚拟状态：管理子节点
	WorkStateType                     // 工作状态：执行具体行为
	TaskStateType                     // 任务状态：高优先级，不可中断
)

// State 状态接口
type State interface {
	ID() string
	Name() string
	Type() StateType
	Children() []State
	// CheckEntryCondition 检查是否满足进入条件
	CheckEntryCondition(ctx *Context) bool

	// OnEnter 进入状态时触发
	OnEnter(ctx *Context)

	// OnUpdate 状态更新逻辑，返回下一个要跳转的状态ID，若为""则保持当前状态，若为"__DONE__"则表示状态结束
	OnUpdate(ctx *Context) string

	// OnExit 退出状态时触发
	OnExit(ctx *Context)

	// OnResume 从中断中恢复时触发
	OnResume(ctx *Context)

	// IsInterruptible 是否允许被高优先级任务打断
	IsInterruptible() bool
}

// Context 运行上下文，用于传递数据
type Context struct {
	Data    map[string]interface{}
	Current string
}

func NewContext() *Context {
	return &Context{
		Data:    make(map[string]interface{}),
		Current: "",
	}
}

// FSM 有限状态机核心
type FSM struct {
	states         map[string]State
	currentState   State
	initialStateID string
	// 任务队列 (高优先级)
	taskQueue []State

	// 中断栈 (用于保存被中断的状态)
	interruptStack []State

	ctx *Context

	isRunning bool
}

func NewFSM() *FSM {
	return &FSM{
		states:         make(map[string]State),
		taskQueue:      make([]State, 0),
		interruptStack: make([]State, 0),
		ctx:            NewContext(),
	}
}

// RegisterState 注册状态
// RegisterState 是FSM结构体的一个方法，用于注册一个新的状态
// 参数:
//   - s: 要注册的状态，实现了State接口
//
// 功能:
//
//	将传入的状态s以其ID为键存储到f.states映射中，使FSM能够识别和管理该状态
func (f *FSM) RegisterState(s State) {
	f.states[s.ID()] = s // 将状态s存储到states映射中，键为状态的ID
	children := s.Children()
	if children != nil {
		for _, child := range children {
			f.RegisterState(child)
		}
	}
}

// AddTask 添加高优先级任务
func (f *FSM) AddTask(s State) {
	if s.Type() != TaskStateType {
		fmt.Printf("[FSM] Warning: Adding non-task state %s to task queue\n", s.Name())
	}
	f.taskQueue = append(f.taskQueue, s)
	fmt.Printf("[FSM] Added Task: %s\n", s.Name())
}

// Start 启动状态机
func (f *FSM) Start(initialStateID string) {
	startState, ok := f.states[initialStateID]
	if !ok {
		panic(fmt.Sprintf("Initial state %s not found", initialStateID))
	}
	f.initialStateID = initialStateID
	if !startState.CheckEntryCondition(f.ctx) {
		panic(fmt.Sprintf("Initial state %s entry condition failed", initialStateID))
	}

	f.currentState = startState
	f.isRunning = true
	fmt.Printf("[FSM] Start. Initial State: %s\n", f.currentState.Name())
	f.currentState.OnEnter(f.ctx)
}

// Update 驱动状态机运行
func (f *FSM) Update() {
	if !f.isRunning || f.currentState == nil {
		return
	}

	// 1. 检查是否有高优先级任务需要执行中断
	if len(f.taskQueue) > 0 {
		nextTask := f.taskQueue[0]

		// 如果当前已经在执行任务状态，通常不中断任务，除非有更高级别的逻辑（这里简化为任务不可中断其他任务）
		// 如果当前是工作状态，且允许中断
		if f.currentState.Type() == WorkStateType && f.currentState.IsInterruptible() {
			// 执行中断逻辑
			f.taskQueue = f.taskQueue[1:] // 移除任务
			f.interrupt(nextTask)
			return // 本次Update周期已用于切换状态
		} else if f.currentState.Type() == VirtualStateType {
			// 虚拟状态通常只是容器，是否打断取决于具体设计。
			// 假设虚拟状态如果不执行具体Work逻辑，可以被打断，或者它只是个壳。
			// 这里为了简单，假设虚拟状态如果正在管理一个子节点，应该由子节点的Update逻辑决定。
			// 但如果虚拟状态本身就是当前State，我们允许打断。
			f.taskQueue = f.taskQueue[1:]
			f.interrupt(nextTask)
			return
		} else if f.currentState.Type() == TaskStateType {
			// 当前也是任务，排队等待当前任务完成
		}
	}

	// 2. 正常执行当前状态逻辑
	nextStateID := f.currentState.OnUpdate(f.ctx)

	// 3. 处理状态转换
	if nextStateID == Done {
		f.handleStateDone()
	} else if nextStateID != "" && nextStateID != f.currentState.ID() {
		f.changeState(nextStateID)
	}
}

// interrupt 执行中断操作
func (f *FSM) interrupt(nextTask State) {
	fmt.Printf("[FSM] Interrupting %s for Task %s\n", f.currentState.Name(), nextTask.Name())

	// 保存当前状态到栈
	// 注意：这里没有调用OnExit，而是挂起。
	// 也可以设计为调用OnSuspend，这里为了简化，我们假设中断不完全退出，或者我们在恢复时有特殊处理
	// 通常做法：Exit -> Push -> Enter Task. Resume -> Pop -> Enter(ResumeMode).

	// 这里我们选择：Exit当前状态，但压栈
	f.currentState.OnExit(f.ctx)
	f.interruptStack = append(f.interruptStack, f.currentState)

	// 切换到任务状态
	f.currentState = nextTask
	// 任务状态不需要注册在map里也能运行，但最好还是注册。这里假设Task也是独立的实例。
	f.currentState.OnEnter(f.ctx)
}

// handleStateDone 处理状态完成的情况
func (f *FSM) handleStateDone() {
	fmt.Printf("[FSM] State %s Finished.\n", f.currentState.Name())
	f.currentState.OnExit(f.ctx)

	// 1. 优先检查任务队列 (Task Queue) - 连续执行任务
	if len(f.taskQueue) > 0 {
		nextTask := f.taskQueue[0]
		f.taskQueue = f.taskQueue[1:]
		fmt.Printf("[FSM] Starting next task from queue: %s\n", nextTask.Name())
		f.currentState = nextTask
		f.currentState.OnEnter(f.ctx)
		return
	}

	// 2. 检查中断栈 (Resume)
	if len(f.interruptStack) > 0 {
		// Pop
		lastIdx := len(f.interruptStack) - 1
		prevState := f.interruptStack[lastIdx]
		f.interruptStack = f.interruptStack[:lastIdx]

		fmt.Printf("[FSM] Resuming state: %s\n", prevState.Name())
		f.currentState = prevState
		// 恢复状态，调用 OnResume 而不是 OnEnter
		f.currentState.OnResume(f.ctx)
		return
	}

	// 3. 既没有任务也没有要恢复的状态，进入Root
	f.returnToInitial()
}

// changeState 普通状态切换
func (f *FSM) changeState(nextStateID string) {
	f.ctx.Current = nextStateID
	nextState, ok := f.states[nextStateID]
	if !ok {
		fmt.Printf("[FSM] Error: State %s not found\n", nextStateID)
		return
	}

	//if !nextState.CheckEntryCondition(f.ctx) {
	//	fmt.Printf("[FSM] Condition failed for %s. Staying in %s\n", nextState.Name(), f.currentState.Name())
	//	return
	//}

	fmt.Printf("[FSM] Transition: %s -> %s\n", f.currentState.Name(), nextState.Name())
	f.currentState.OnExit(f.ctx)
	f.currentState = nextState
	f.currentState.OnEnter(f.ctx)
}

func (f *FSM) returnToInitial() {
	if f.initialStateID == "" {
		fmt.Println("[FSM] No initial state configured. Stopping.")
		f.currentState = nil
		f.isRunning = false
		return
	}
	nextState, ok := f.states[f.initialStateID]
	if !ok {
		fmt.Printf("[FSM] Initial state %s not found. Stopping.\n", f.initialStateID)
		f.currentState = nil
		f.isRunning = false
		return
	}
	if !nextState.CheckEntryCondition(f.ctx) {
		fmt.Printf("[FSM] Initial state %s entry condition failed. Stopping.\n", nextState.Name())
		f.currentState = nil
		f.isRunning = false
		return
	}
	fmt.Printf("[FSM] Returning to initial state: %s\n", nextState.Name())
	f.currentState = nextState
	f.isRunning = true
	f.currentState.OnEnter(f.ctx)
}
