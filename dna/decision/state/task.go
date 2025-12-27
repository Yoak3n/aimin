package state

import (
	"time"
)

// Task 任务怎么从外部插入，从内部添加，是个问题
type Task struct {
	ID        string
	Type      uint
	Priority  int // 优先级，数值越高优先级越高
	Data      any
	CreatedAt time.Time
	Timeout   time.Duration
}
