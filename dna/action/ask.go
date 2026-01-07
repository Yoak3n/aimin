package action

import (
	"context"
	"fmt"
	"time"
)

func ProactiveAsk(question string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	return askFromCmd(ctx, question)
}

func askFromCmd(ctx context.Context, question string) []string {
	resultChan := make(chan string, 1)
	go func() {
		var answer string
		_, err := fmt.Scanln(&answer)
		if err != nil {
			resultChan <- ""
			return
		}
		// 将输入分割成切片返回
		resultChan <- answer
	}()

	select {
	case result := <-resultChan:
		return []string{result}
	case <-ctx.Done():
		fmt.Println("\n输入超时")
		return []string{}
	}
}
