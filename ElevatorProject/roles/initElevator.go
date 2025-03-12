package roles

import (
	"ElevatorProject/Driver"
	"ElevatorProject/fsm"
	"ElevatorProject/elevator"
	"ElevatorProject/elevator_logic"
	"ElevatorProject/elevator_io_device"
	"ElevatorProject/timers"
	
)

func InitElevator() {
	//husk å endre port om du er på samme PC
	elevio.Init("localhost:15659", elevator.N_FLOORS)

	fsm := fsm.FSM{El: elevator.Elevator_uninitialized(), Od: elevator_io_device.Elevio_getOutputDevice()}

	elevator_logic.ElevLogic_runElevator(fsm, timers.MaxDuration) //add conn!
}