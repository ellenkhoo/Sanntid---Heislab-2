package roles

import (
	"github.com/ellenkhoo/ElevatorProject/elevator"
	"github.com/ellenkhoo/ElevatorProject/elevator/Driver"
	"github.com/ellenkhoo/ElevatorProject/timers"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

func InitElevator(localIp string, networkChannels sharedConsts.NetworkChannels) elevator.FSM {
	elevio.Init("localhost:15657", elevator.N_FLOORS)

	fsm := elevator.FSM{El: elevator.Elevator_uninitialized(), Od: elevator.Elevio_getOutputDevice()}
	fsm.El.ElevStates.IP = localIp

	go elevator.ElevLogic_runElevator(networkChannels, fsm, timers.MaxDuration)

	return fsm
}