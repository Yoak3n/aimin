package fsm

import "fmt"

// BaseState 基础状态实现，减少样板代码
type BaseState struct {
	id            string
	name          string
	stateType     StateType
	interruptible bool
	checkEntry    func(*Context) bool
}

const (
	Done      = "__DONE__"
	Interrupt = "__INTERRUPT__"
)

func (b *BaseState) ID() string { return b.id }
func (b *BaseState) Name() string {
	return b.name
}
func (b *BaseState) Type() StateType { return b.stateType }
func (b *BaseState) Children() []State {
	return nil
}
func (b *BaseState) IsInterruptible() bool { return b.interruptible }
func (b *BaseState) CheckEntryCondition(ctx *Context) bool {
	if b.checkEntry != nil {
		return b.checkEntry(ctx)
	}
	return true
}
func (b *BaseState) OnEnter(ctx *Context)         { fmt.Printf("  -> Enter %s\n", b.name) }
func (b *BaseState) OnExit(ctx *Context)          { fmt.Printf("  <- Exit %s\n", b.name) }
func (b *BaseState) OnResume(ctx *Context)        { fmt.Printf("  -> Resume %s\n", b.name) }
func (b *BaseState) OnUpdate(ctx *Context) string { return "" } // 默认保持

// WorkAction 定义工作状态的具体行为逻辑
// 返回值: nextStateID (""表示继续当前状态, "__DONE__"表示结束, 其他值为跳转目标)
type WorkAction func(ctx *Context) string

// WorkState 普通工作状态
type WorkState struct {
	BaseState
	action WorkAction
}

func (w *WorkState) Children() []State {
	return nil
}

func NewWorkState(id, name string, action WorkAction, check func(ctx *Context) bool) *WorkState {
	return &WorkState{
		BaseState: BaseState{
			id:            id,
			name:          name,
			stateType:     WorkStateType,
			interruptible: true, // 允许打断
			checkEntry:    check,
		},
		action: action,
	}
}

func (w *WorkState) OnUpdate(ctx *Context) string {
	if w.action != nil {
		return w.action(ctx)
	}
	return Done
}

// TaskState 高优先级任务状态
type TaskState struct {
	BaseState
	action func()
}

func (t *TaskState) Children() []State {
	return nil
}

func NewTaskState(id, name string, action func()) *TaskState {
	return &TaskState{
		BaseState: BaseState{
			id:            id,
			name:          name,
			stateType:     TaskStateType,
			interruptible: false, // 任务不可被打断
		},
		action: action,
	}
}

func (t *TaskState) OnUpdate(ctx *Context) string {
	fmt.Printf("    !!! Executing Task %s !!!\n", t.name)
	if t.action != nil {
		t.action()
	}
	return Done // 任务通常是一次性的，执行完即结束
}

// VirtualState 虚拟状态 (管理子节点)
// 这里实现为一个简单的序列执行器：按顺序执行子节点
type VirtualState struct {
	BaseState
	childIDs        []string
	currentChildIdx int
	fsmRef          *FSM // 需要引用FSM来驱动子状态切换，或者简化逻辑
	// 简化逻辑：VirtualState 本身不直接跑子FSM，而是通过返回子状态ID让主FSM去切换
	// 但这样主FSM的changeState是扁平的。
	// 如果要体现"管理"，VirtualState 应该是"Container"。
	// 让我们采用稍微复杂一点的逻辑：VirtualState Update 时负责告诉主FSM去哪个子状态。
	// 但主FSM的 Update 逻辑是 flat 的。
	// 方案B：VirtualState 作为一个 Group Manager。
	// 当进入 VirtualState 时，它自动把 Context 指向第一个 Child。
	// 它的 OnUpdate 逻辑检查 Child 是否完成。
	// 这需要 VirtualState 维护自己的 "Sub-FSM" 或者仅仅是一个简单的状态序列。
}

func (v *VirtualState) Children() []State {
	return nil
}

func NewVirtualState(id, name string, children []string) *VirtualState {
	return &VirtualState{
		BaseState: BaseState{
			id:            id,
			name:          name,
			stateType:     VirtualStateType,
			interruptible: true,
		},
		childIDs: children,
	}
}

// VirtualState 的逻辑比较特殊，它需要知道子状态是否结束。
// 由于我们在 Core 里定义的 Update 返回的是 nextStateID。
// 我们可以让 VirtualState 充当一个 Router。
// 这里的实现方式：
// 1. Enter VirtualState -> 立即返回第一个 Child 的 ID。
// 2. 但这样 VirtualState 就退出了。
//
// 用户的需求是 "虚拟状态，主要用于管理子节点"。
// 我们可以把 VirtualState 当作一个"父节点"，它本身不结束，直到所有子节点结束。
// 但 Go 的这个简单实现里，State 是互斥的。
// 我们可以这样做：
// 可以在 Context 里记录 parent。
// 或者，我们可以简单地让 VirtualState 作为一个生成器。
//
// 让我们换个思路：VirtualState 是一个复合状态。
// 但为了保持 FSM Core 简单，我们不搞递归 FSM。
// 我们让 VirtualState 仅仅作为一个"宏"状态。
// 当 FSM 处于 VirtualState 时，它实际上是在执行 VirtualState 的逻辑，
// 而 VirtualState 的逻辑就是手动调用子状态的逻辑。
// 这种方式比较重。
//
// 另一种常见方式：
// 扁平化注册。VirtualState 只是逻辑上的分组。
// 但为了满足"管理"的语义，我们实现一个 CompositeState。
//
// 我们用一个简单的 Stack 或者是 Index 来管理。
// 当 VirtualState OnEnter -> index = 0
// VirtualState OnUpdate ->
//    Delegate to children[index].OnUpdate()
//    If child returns DONE -> index++, Delegate to next.
//    If all children DONE -> return DONE.
// 这要求 VirtualState 持有子状态的实例，而不仅仅是 ID。
// 让我们修改一下构造函数，传入实例。

type CompositeState struct {
	BaseState
	children  []State
	current   int
	RouterKey string
	selector  func(*Context, []State) int
}

func (c *CompositeState) Children() []State {
	return c.children
}

func NewCompositeState(id, name string, children []State, check func(ctx *Context) bool) *CompositeState {
	cs := &CompositeState{
		BaseState: BaseState{
			id:            id,
			name:          name,
			stateType:     VirtualStateType,
			interruptible: true,
			checkEntry:    check,
		},
		children:  children,
		RouterKey: name,
	}
	return cs
}
func (c *CompositeState) SetRouterKey(key string) {
	c.RouterKey = key
}

func (c *CompositeState) SetSelect(s func(*Context, []State) int) {
	c.selector = s
}

func (c *CompositeState) OnEnter(ctx *Context) {
	c.BaseState.OnEnter(ctx)
	start := 0
	// 外部选择children,返回索引，实际是操作routerKey
	if c.selector != nil {
		idx := c.selector(ctx, c.children)
		if idx >= 0 && idx < len(c.children) {
			start = idx
		}
	}
	c.current = start
	// 让selector来做
	//if c.routerKey != "" && len(c.children) > 0 {
	//	ctx.Data[c.routerKey] = c.children[c.current].ID()
	//}
	//if len(c.children) > 0 {
	//	if c.children[c.current].CheckEntryCondition(ctx) {
	//		c.children[c.current].OnEnter(ctx)
	//	} else {
	//		fmt.Printf("  [Composite] Selected child %s blocked by entry condition\n", c.children[c.current].Name())
	//	}
	//}
}

func (c *CompositeState) OnUpdate(ctx *Context) string {
	if c.current >= len(c.children) {
		return Done
	}
	child := c.children[c.current]
	//for idx, _ := range c.children {
	//	if nextChild.CheckEntryCondition(ctx) {
	//		child = nextChild
	//		c.current = idx
	//		break
	//	}
	//}
	for idx := c.current; idx < len(c.children); idx++ {
		nextChild := c.children[idx]
		if nextChild.CheckEntryCondition(ctx) {
			child = nextChild
			c.current = idx
			break
		}
	}
	child.OnEnter(ctx)
	// 代理执行子状态
	result := child.OnUpdate(ctx)
	if result == Done {
		child.OnExit(ctx)
		if c.RouterKey != "" {
			delete(ctx.Data, c.RouterKey)
		}
		return Done
	} else if result != "" {
		// 子状态请求跳转？
		// 在复合状态内部跳转比较复杂，这里假设子状态只能返回 DONE 或 空
		// 如果子状态返回了具体的 ID，说明它想跳出复合状态，或者跳到复合状态内的其他节点
		// 这里简化：如果子状态返回非空ID，且不是DONE，我们认为它要跳出整个复合状态
		return result
	}

	return ""
}

func (c *CompositeState) OnExit(ctx *Context) {
	// 退出时，如果当前子状态还在运行，也要退出
	//if c.current < len(c.children) {
	//	c.children[c.current].OnExit(ctx)
	//}
	c.BaseState.OnExit(ctx)
}

func (c *CompositeState) OnResume(ctx *Context) {
	c.BaseState.OnResume(ctx)
	// 恢复时，恢复当前子状态
	if c.current < len(c.children) {
		// 注意：这里子状态可能也需要 Resume，或者简单地 OnEnter/Resume
		// 简单起见，如果子状态支持 Resume，调用 Resume，否则 Enter
		// 但接口里只有 OnResume。
		c.children[c.current].OnResume(ctx)
	}
}
