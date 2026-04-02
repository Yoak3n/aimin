package action

import (
	"context"
	"time"
)

var RemoteAsk func(context.Context, string) []string

func ProactiveAsk(question string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
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
