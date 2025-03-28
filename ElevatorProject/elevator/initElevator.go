package elevator

import (
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	"github.com/ellenkhoo/ElevatorProject/timers"
)

func InitElevator(localID string, networkChannels *sharedConsts.NetworkChannels) *FSM {
	InitializeElevatorDriver("localhost:15658", N_FLOORS)

	fsm := FSM{El: InitializeElevator(), Od: GetOutputDevice()}
	fsm.Fsm_mtx.Lock()
	fsm.El.ElevStates.ID = localID
	fsm.Fsm_mtx.Unlock()

	go RunElevator(networkChannels, &fsm, timers.MaxDuration)

	return &fsm
}
