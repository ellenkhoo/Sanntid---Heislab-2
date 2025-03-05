package elevator_logic

import (
	"Driver-go/elevio"
	"elevator"
	"fmt"
	"fsm"
	"request"
	"time"
	"timer"
	"net"
)

func ElevLogic_runElevator (fsm fsmpkg.FSM, maxDuration time.Duration, conn net.Conn) {

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
	go timerpkg.Timer_start(timer, start_timer)

	requestpkg.Clear_all_requests(fsm.El)
	fsm.SetAllLights()

	if elevio.GetFloor() == -1 {
		fsm.Fsm_onInitBetweenFloors()
		fmt.Printf("Init between floor")
	}

	// fmt.Printf("Current floor: %d \n", fsm.El.Floor)
	// fmt.Printf("Current Dirn: %d \n", fsm.El.Dirn)

	for {
		select {
		case order := <-buttons_chan:
			fmt.Printf("Button pushed. Order at floor: %d", order.Floor)
			// If cab call
			if order.Button == elevator.B_Cab {
				fsm.El.CabRequests[order.Floor] = true
			}
			
			// Send beskjed til master: ordre + state

				//hente alle state fra heisene og deretter kjøre FunctionHRA, får da en output om hvor heisen skal gå til alle heiser
				// output := hraAssigner.FunctionHRA(states, hallrequests)
				// for id, hallrequests := range *output {
				// 	elevatorID, err := strconv.Atoi(id)
				// 	if err != nil {
				// 		fmt.Println("Error: ", err)
				// 	}
				// 	hallrequests -> denne må sendes til heis med samme elevator id som variabel elevatorID

				// }

			// if !(fsm.El.Requests[order.Floor][order.Button]) {
			// 	fsm.Fsm_onRequestButtonPress(order.Floor, order.Button, start_timer)
			// 	communicationpkg.Comm_sendReceivedOrder(order, fsm.El.IP, conn)
			// }

		case floor_input := <-floors_chan:
			fmt.Printf("Floor sensor: %d\n", floor_input)

			if floor_input != -1 && floor_input != fsm.El.Floor {
				fsm.Fsm_onFloorArrival(floor_input, start_timer)
			}

		case obstruction := <-obstruction_chan:
			if obstruction {
				if fsm.El.Behaviour == elevator.EB_DoorOpen {
					start_timer <- maxDuration
				}
			} else {
				start_timer <- fsm.El.Config.DoorOpenDuration
			}
			
		case <-timer.C:
			fsm.Fsm_onDoorTimeout(start_timer)
		}
	}
}