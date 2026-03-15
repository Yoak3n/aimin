package ws

import (
	"log"
)

func (wh *WebSocketHub) BroadcastState(state string) {
	select {
	case wh.State <- state:
	default:
		log.Println("State channel full or no receiver, skipping state update:", state)
	}
}
