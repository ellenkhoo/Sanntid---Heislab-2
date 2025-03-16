package roles

import (
	"ElevatorProject/Driver"
	"ElevatorProject/fsm"
	"ElevatorProject/elevator"
	"ElevatorProject/elevator_logic"
	"ElevatorProject/elevator_io_device"
	"ElevatorProject/timers"
	"net"
)

func InitElevator(rank int, conn net.Conn) {
	elevio.Init("localhost:15657", elevator.N_FLOORS)

	fsm := fsm.FSM{El: elevator.Elevator_uninitialized(), Od: elevator_io_device.Elevio_getOutputDevice()}
	fsm.El.Rank = rank

	elevator_logic.ElevLogic_runElevator(fsm, timers.MaxDuration, conn)
}