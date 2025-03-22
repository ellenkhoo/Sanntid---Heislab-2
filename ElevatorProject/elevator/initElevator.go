package elevator

import (
	"github.com/ellenkhoo/ElevatorProject/elevator/Driver"
	"github.com/ellenkhoo/ElevatorProject/timers"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

func InitElevator(localIP string, networkChannels *sharedConsts.NetworkChannels) FSM {
	elevio.Init("localhost:15657", N_FLOORS)

	fsm := FSM{El: Elevator_uninitialized(), Od: Elevio_getOutputDevice()}
	fsm.El.ElevStates.IP = localIP

	go ElevLogic_runElevator(networkChannels, fsm, timers.MaxDuration)

	return fsm
}