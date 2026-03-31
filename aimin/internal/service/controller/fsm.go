package controller

import (
	"github.com/Yoak3n/aimin/aimin/app/componet"
)

func GetFSMStatus() string {
	fsm := componet.GetGlobalComponent().FSM()
	return fsm.CurrentStatus()
}
