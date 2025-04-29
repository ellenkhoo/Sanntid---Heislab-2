package elevator

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	"github.com/ellenkhoo/ElevatorProject/timers"
)

func SendCurrentState(networkChannels *sharedConsts.NetworkChannels, elevator Elevator) {

	msgToMaster := FormatElevStates(elevator)

	if elevator.ElevStates == nil {
		fmt.Println("ElevStates is nil")
	}

	elevStatesJSON, err := json.Marshal(msgToMaster)
	if err != nil {
		fmt.Println("Error marshalling elevStates: ", err)
		return
	}

	stateMsg := sharedConsts.Message{
		Type:    sharedConsts.CurrentStateMessage,
		Target:  sharedConsts.TargetMaster,
		Payload: elevStatesJSON,
	}

	networkChannels.SendChan <- stateMsg
}

func SendLocalOrder(order ButtonEvent, networkChannels *sharedConsts.NetworkChannels) {
	orderJSON, err := json.Marshal(order)
	if err != nil {
		fmt.Println("Error marshalling order: ", err)
		return
	}

	reqMsg := sharedConsts.Message{
		Type:    sharedConsts.LocalHallRequestMessage,
		Target:  sharedConsts.TargetMaster,
		Payload: orderJSON,
	}

	networkChannels.SendChan <- reqMsg
}

func RunElevator(networkChannels *sharedConsts.NetworkChannels, fsm *FSM, maxDuration time.Duration) {

	// Initialize channels
	buttonsChan := make(chan ButtonEvent)
	floorsChan := make(chan int)
	obstructionChan := make(chan bool)
	stopChan := make(chan bool)
	timerChan := make(chan time.Duration)

	// Initialize timer, stop it until needed
	timer := time.NewTimer(time.Duration(fsm.Elevator.DoorOpenDuration))
	timer.Stop()

	// Start Goroutines
	go PollButtons(buttonsChan)
	go PollFloorSensor(floorsChan)
	go PollObstructionSwitch(obstructionChan)
	go PollStopButton(stopChan)
	go timers.StartTimer(timer, timerChan)

	ClearAllRequests(*fsm.Elevator)
	fsm.SetAllLights()

	if GetFloor() == -1 {
		fsm.InitBetweenFloors()
		fmt.Printf("Elevator initialized between floors")
	}

	for {
		select {

		case order := <-buttonsChan:
			fmt.Printf("Button pushed. Order at floor: %d\n", order.Floor)

			fsm.FSM_mutex.Lock()

			// If cab request
			if order.Button == B_Cab {
				fsm.Elevator.ElevStates.CabRequests[order.Floor] = true
			} else {
				SendLocalOrder(order, networkChannels)
			}

			fsm.FSM_mutex.Unlock()
			time.Sleep(1 * time.Second)
			SendCurrentState(networkChannels, *fsm.Elevator)

		case floorInput := <-floorsChan:
			fmt.Printf("Elevator is at floor: %d\n", floorInput)

			if floorInput != -1 && floorInput != fsm.Elevator.ElevStates.CurrentFloor {
				fsm.OnFloorArrival(networkChannels, floorInput, timerChan)
				SendCurrentState(networkChannels, *fsm.Elevator)
			}

		case obstruction := <-obstructionChan:
			if obstruction {
				if fsm.Elevator.Behaviour == EB_DoorOpen {
					timerChan <- maxDuration
				}
			} else {
				timerChan <- fsm.Elevator.DoorOpenDuration
			}

			SendCurrentState(networkChannels, *fsm.Elevator)

		case <-timer.C:
			fsm.OnDoorTimeout(timerChan)
			SendCurrentState(networkChannels, *fsm.Elevator)

		case <-networkChannels.UpdateChan:
			fsm.HandleRequestsToDo(networkChannels, timerChan)
		}
	}
}
