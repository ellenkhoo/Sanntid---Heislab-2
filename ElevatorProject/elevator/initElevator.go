package elevator

import (
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	"github.com/ellenkhoo/ElevatorProject/timers"
)

func InitElevator(localIP string, networkChannels *sharedConsts.NetworkChannels) *FSM {
	InitializeElevatorDriver("localhost:15657", N_FLOORS)

	fsm := FSM{El: InitializeElevator(), Od: GetOutputDevice()}
	fsm.Fsm_mtx.Lock()
	fsm.El.ElevStates.IP = localIP
	fsm.Fsm_mtx.Unlock()

	go RunElevator(networkChannels, &fsm, timers.MaxDuration)

	return &fsm
}
