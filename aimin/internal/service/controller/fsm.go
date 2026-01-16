package controller

import (
	"github.com/Yoak3n/aimin/aimin/cmd/app/componet"
)

func GetFSMStatus() string {
	fsm := componet.GetGlobalComponent().FSM()
	return fsm.CurrentStatus()
}
