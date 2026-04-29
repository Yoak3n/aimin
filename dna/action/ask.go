package action

import (
	"context"
	"time"
)

var RemoteAsk func(context.Context, string) []string

func ProactiveAsk(question string) []string {
	return ProactiveAskWithContext(context.Background(), question)
}

func ProactiveAskWithContext(base context.Context, question string) []string {
	if base == nil {
		base = context.Background()
	}
	ctx, cancel := context.WithTimeout(base, 5*time.Minute)
	defer cancel()
	if RemoteAsk != nil {
		out := RemoteAsk(ctx, question)
		if ctx.Err() == context.DeadlineExceeded {
			return []string{"[AskUser][超时] 用户未在超时时间内回复。"}
		}
		return out
	}
	return []string{}
}
