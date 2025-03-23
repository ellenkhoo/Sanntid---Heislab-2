package elevator

import (
	elevio "github.com/ellenkhoo/ElevatorProject/elevator/Driver"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	"github.com/ellenkhoo/ElevatorProject/timers"
	"encoding/json"
	"fmt"
	"time"
)

func SendCurrentState(networkChannels *sharedConsts.NetworkChannels, fsm *FSM) {

	fsm.Fsm_mtx.Lock()
	FormatElevStates(fsm.El, fsm.El.ElevStates)
	fsm.Fsm_mtx.Unlock()

	msgToMaster := MessageToMaster{
		ElevStates:   *fsm.El.ElevStates,
		RequestsToDo: fsm.El.RequestsToDo,
	}

	// Marshal message
	elevStatesJSON, err := json.Marshal(msgToMaster)
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
func RunElevator(networkChannels *sharedConsts.NetworkChannels, fsm *FSM, maxDuration time.Duration) {

	fmt.Println("Arrived at runElevator")

	// Initialize channels
	buttonsChan := make(chan elevio.ButtonEvent)
	floorsChan := make(chan int)
	obstructionChan := make(chan bool)
	stopChan := make(chan bool)
	timerChan := make(chan time.Duration)

	// Initialize timer, stop it until needed
	timer := time.NewTimer(time.Duration(timers.DoorOpenDuration))
	timer.Stop()

	// Start Goroutines
	go elevio.PollButtons(buttonsChan)
	go elevio.PollFloorSensor(floorsChan)
	go elevio.PollObstructionSwitch(obstructionChan)
	go elevio.PollStopButton(stopChan)
	go timers.Timer_start(timer, timerChan)

	ClearAllRequests(*fsm.El)
	fsm.SetAllLights()

	if elevio.GetFloor() == -1 {
		fsm.InitBetweenFloors()
		fmt.Printf("Elevator initialized between floors")
	}

	for {
		select {

		case order := <-buttonsChan:
			fmt.Printf("Button pushed. Order at floor: %d\n", order.Floor)

			fsm.Fsm_mtx.Lock()

			// If cab request
			if order.Button == B_Cab {
				fsm.El.ElevStates.CabRequests[order.Floor] = true
			} else {
				sendLocalOrder(order, networkChannels)
			}

			fsm.Fsm_mtx.Unlock()

			SendCurrentState(networkChannels, fsm)

		case floorInput := <-floorsChan:
			fmt.Printf("Floor sensor: %d\n", floorInput)

			if floorInput != -1 && floorInput != fsm.El.ElevStates.Floor {
				fsm.OnFloorArrival(networkChannels, floorInput, timerChan)
			}

		case obstruction := <-obstructionChan:
			if obstruction {
				if fsm.El.Behaviour == EB_DoorOpen {
					timerChan <- maxDuration
				}
			} else {
				timerChan <- timers.DoorOpenDuration
			}

			SendCurrentState(networkChannels, fsm)

		case <-timer.C:
			fsm.OnDoorTimeout(timerChan)
			SendCurrentState(networkChannels, fsm)

		case <-networkChannels.UpdateChan:
			fmt.Println("Received update")
			fsm.HandleRequestsToDo(networkChannels, timerChan)
		}
	}
}
