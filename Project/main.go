package main

import (
	//"PR_translated_to_go/elevator_io_device"
	"Driver-go/elevio"
	"communicationpkg"
	elevatorpkg "elevator"
	elevator_io_devicepkg "elevator_io_device"
	"elevator_logicpkg"
	"fmt"
	fsmpkg "fsm"

	// requestpkg "request"
	"time"
	// timerpkg "timer"
	// "elevator_logicpkg"
	"os"
)

func main() {

	if len(os.Args) > 1 && os.Args[1] == "slave" {
		conn := communicationpkg.Comm_slaveConnectToMaster()
	} else {
		conn := communicationpkg.Comm_masterConnectToSlave()
	}

	fmt.Println("Started!")

	// burde vel ikke måtte definere denne på nytt, er jo definert i elevio
	numFloors := 4
	const maxDuration time.Duration = 1<<63 - 1

	elevio.Init("localhost:15657", numFloors)

	fsm := fsmpkg.FSM{El: elevatorpkg.Elevator_uninitialized(), Od: elevator_io_devicepkg.Elevio_getOutputDevice()}

	elevator_logicpkg.ElevLogic_runMaster(fsm, maxDuration)
}
