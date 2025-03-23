package elevator

import (
	elevio "github.com/ellenkhoo/ElevatorProject/elevator/Driver"
	//"github.com/ellenkhoo/ElevatorProject/comm"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	"github.com/ellenkhoo/ElevatorProject/timers"
	//"net"
)

func SendCurrentState(networkChannels *sharedConsts.NetworkChannels, fsm *FSM) {
	// Marshal elevStates
	elevStatesJSON, err := json.Marshal(fsm.El.ElevStates)
	if err != nil {
		fmt.Println("Error marshalling elevStates: ", err)
		return
	}
	// Create message
	stateMsg := sharedConsts.Message{
		Type:    sharedConsts.CurrentStateMessage,
		Target:  sharedConsts.TargetMaster,
		Payload: elevStatesJSON,
	}
	//Send message
	networkChannels.SendChan <- stateMsg
}

func sendLocalOrder(order elevio.ButtonEvent, networkChannels *sharedConsts.NetworkChannels) {
	// Marshal order
	orderJSON, err := json.Marshal(order)
	if err != nil {
		fmt.Println("Error marshalling order: ", err)
		return
	}
	// Create message
	reqMsg := sharedConsts.Message{
		Type:    sharedConsts.LocalRequestMessage,
		Target:  sharedConsts.TargetMaster,
		Payload: orderJSON,
	}
	// Send message
	networkChannels.SendChan <- reqMsg
}
func ElevLogic_runElevator(networkChannels *sharedConsts.NetworkChannels, fsm *FSM, maxDuration time.Duration) {

	fmt.Println("Arrived at runElevator")

	// Initialize channels
	buttons_chan := make(chan elevio.ButtonEvent)
	floors_chan := make(chan int)
	obstruction_chan := make(chan bool)
	stop_chan := make(chan bool)
	start_timer := make(chan time.Duration)

	// Initialize timer, stop it until needed
	timer := time.NewTimer(time.Duration(timers.DoorOpenDuration))
	timer.Stop()

	// Start goroutines
	go elevio.PollButtons(buttons_chan)
	go elevio.PollFloorSensor(floors_chan)
	go elevio.PollObstructionSwitch(obstruction_chan)
	go elevio.PollStopButton(stop_chan)
	go timers.Timer_start(timer, start_timer)

	Clear_all_requests(*fsm.El)
	fsm.SetAllLights()

	if elevio.GetFloor() == -1 {
		fsm.Fsm_onInitBetweenFloors()
		fmt.Printf("Init between floor")
	}

	for {
		select {

		case order := <-buttons_chan:
			fmt.Printf("Button pushed. Order at floor: %d\n", order.Floor)
	
			fsm.Fsm_mtx.Lock()

			// If cab call
			if order.Button == B_Cab {
				fsm.El.ElevStates.CabRequests[order.Floor] = true
				fmt.Println("Cab request in elevLogic: ", fsm.El.ElevStates.CabRequests)
			} else {
				sendLocalOrder(order, networkChannels)
			}

			fsm.Fsm_mtx.Unlock()

			SendCurrentState(networkChannels, fsm)

		case floor_input := <-floors_chan:
			fmt.Printf("Floor sensor: %d\n", floor_input)

			// fsm.Fsm_mtx.Lock()

			if floor_input != -1 && floor_input != fsm.El.ElevStates.Floor {
				//Master informeres i funksjonskallet nedenfor
				fsm.Fsm_onFloorArrival(networkChannels, floor_input, start_timer)
			}

			// fsm.Fsm_mtx.Unlock()

		case obstruction := <-obstruction_chan:
			if obstruction {
				if fsm.El.Behaviour == EB_DoorOpen {
					start_timer <- maxDuration
				}
			} else {
				start_timer <- timers.DoorOpenDuration
			}
			
			SendCurrentState(networkChannels, fsm)

		case <-timer.C:
			fsm.Fsm_onDoorTimeout(start_timer)
			// comm.Comm_sendCurrentState(fsm.El.ElevStates, conn)

		case <- networkChannels.UpdateChan:
			fmt.Println("Received update")
			fsm.Fsm_onRequestsToDo(networkChannels,start_timer)
		}
	}
}
