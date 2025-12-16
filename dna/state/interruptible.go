package state

import (
	"time"
)

type Task struct {
	ID        string
	Type      uint
	Priority  int // 优先级，数值越高优先级越高
	Data      any
	CreatedAt time.Time
	Timeout   time.Duration
}
