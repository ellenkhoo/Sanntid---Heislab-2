package main

import (
	"ElevatorProject/elevator"
	"ElevatorProject/hra"
	"fmt"
)

func main() {

	var elevst1 = elevator.ElevStates{Behaviour: "moving", Floor: 0, Direction: "up", CabRequests: []bool{false, false, false, false}, ID: 0}
	var elevst2 = elevator.ElevStates{Behaviour: "idle", Floor: 0, Direction: "down", CabRequests: []bool{true, false, false, false}, ID: 1}

	var allElevStates = make(map[int]elevator.ElevStates)

	allElevStates[10] = elevst1
	allElevStates[2] = elevst2

	output := hra.SendStateToHRA(allElevStates, [][2]bool{{false, false}, {false, true}, {false, false}, {false, false}})

	for k, v := range *output {
		fmt.Printf("%6v :  %+v\n", k, v)
	}

	// 	var conn net.Conn

	// 	if len(os.Args) > 1 && os.Args[1] == "slave" {
	// 		conn = comm.Comm_slaveConnectToMaster()
	// 	} else {
	// 		conn = comm.Comm_masterConnectToSlave()
	// 	}

	// 	fmt.Println("Started!")

	// 	// burde vel ikke måtte definere denne på nytt, er jo definert i elevio?
	// 	numFloors := 4
	// 	const maxDuration time.Duration = 1<<63 - 1

	// 	elevio.Init("localhost:15657", numFloors)

	// 	fsm := fsm.FSM{El: elevator.Elevator_uninitialized(), Od: elevator_io_device.Elevio_getOutputDevice()}

	// 	elevator_logic.ElevLogic_runElevator(fsm, maxDuration, conn)
	//
}
