package roles

import (
	"github.com/ellenkhoo/ElevatorProject/elevator"
	"github.com/ellenkhoo/ElevatorProject/elevator/Driver"
	"github.com/ellenkhoo/ElevatorProject/timers"
	"net"
)

func InitElevator(rank int, conn net.Conn) {
	elevio.Init("localhost:15657", elevator.N_FLOORS)

	fsm := elevator.FSM{El: elevator.Elevator_uninitialized(), Od: elevator.Elevio_getOutputDevice()}
	fsm.El.Rank = rank

	elevator.ElevLogic_runElevator(fsm, timers.MaxDuration, conn)
}