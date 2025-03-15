package elevator_logic

import (
	elevio "ElevatorProject/Driver"
	"ElevatorProject/comm"
	"ElevatorProject/elevator"
	"ElevatorProject/fsm"
	"ElevatorProject/request"
	"ElevatorProject/timers"
	"fmt"
	"time"
	"net"
)

func ElevLogic_runElevator(fsm fsm.FSM, maxDuration time.Duration, conn net.Conn)  {

	fmt.Println("Arrived at runElevator")

	// Initialize channels
	buttons_chan := make(chan elevio.ButtonEvent)
	floors_chan := make(chan int)
	obstruction_chan := make(chan bool)
	stop_chan := make(chan bool)
	start_timer := make(chan time.Duration)

	// Initialize timer, stop it until needed
	timer := time.NewTimer(time.Duration(fsm.El.Config.DoorOpenDuration))
	timer.Stop()

	// Start goroutines
	go elevio.PollButtons(buttons_chan)
	go elevio.PollFloorSensor(floors_chan)
	go elevio.PollObstructionSwitch(obstruction_chan)
	go elevio.PollStopButton(stop_chan)
	go timers.Timer_start(timer, start_timer)

	request.Clear_all_requests(fsm.El)
	fsm.SetAllLights()

	if elevio.GetFloor() == -1 {
		fsm.Fsm_onInitBetweenFloors()
		fmt.Printf("Init between floor")
	}

	for {
		select {
		case order := <-buttons_chan:
			fmt.Printf("Button pushed. Order at floor: %d", order.Floor)
			// If cab call
			if order.Button == elevator.B_Cab {
				fsm.El.ElevStates.CabRequests[order.Floor] = true
			}

			// Send request + current state to master
			if !(fsm.El.GlobalHallRequests[order.Floor][order.Button]) {
				comm.Comm_sendReceivedOrder(order, conn)
			}
			comm.Comm_sendCurrentState(fsm.El.ElevStates, conn)

		case floor_input := <-floors_chan:
			fmt.Printf("Floor sensor: %d\n", floor_input)

			if floor_input != -1 && floor_input != fsm.El.ElevStates.Floor {
				fsm.Fsm_onFloorArrival(floor_input, start_timer)
			}
			comm.Comm_sendCurrentState(fsm.El.ElevStates, conn)


		case obstruction := <-obstruction_chan:
			if obstruction {
				if fsm.El.Behaviour == elevator.EB_DoorOpen {
					start_timer <- maxDuration
				}
			} else {
				start_timer <- fsm.El.Config.DoorOpenDuration
			}
			comm.Comm_sendCurrentState(fsm.El.ElevStates, conn)

		case <-timer.C:
			fsm.Fsm_onDoorTimeout(start_timer)
			comm.Comm_sendCurrentState(fsm.El.ElevStates, conn)
		}
		
	}
}
