package node

import (
	"dna/decision/state"
	"errors"
)

type AskData struct {
	Question string
	Answer   string
}

func AskNode() *state.ReturnState {
	return state.NewReturnState(
		Ask,
		AskAction,
	)
}

func AskAction(ctx *state.Context) (any, error) {
	if d, ok := ctx.Data[Ask].(AskData); ok {
		// 等待用户回答（在实际应用中，这里可能涉及异步等待或回调机制）
		a := "我已知晓"
		askData := AskData{
			Question: d.Question,
			Answer:   a,
		}
		ctx.Data[Ask] = askData
		return askData, nil
	}
	return nil, errors.New("no ask data for answer")
}
