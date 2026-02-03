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
		return RemoteAsk(ctx, question)
	}
	return []string{}
}
