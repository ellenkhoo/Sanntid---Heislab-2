package elevator

import (
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	"github.com/ellenkhoo/ElevatorProject/timers"
)

func InitElevator(localID string, networkChannels *sharedConsts.NetworkChannels) *FSM {
	InitializeElevatorDriver("localhost:15658", N_FLOORS)

	fsm := FSM{Elevator: InitializeElevator(), OutputDevice: GetOutputDevice()}
	fsm.FSM_mutex.Lock()
	fsm.Elevator.ElevStates.ID = localID
	fsm.FSM_mutex.Unlock()

	go RunElevator(networkChannels, &fsm, timers.MaxDuration)

	return &fsm
}
