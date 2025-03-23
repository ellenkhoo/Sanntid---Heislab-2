package elevator

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	elevio "github.com/ellenkhoo/ElevatorProject/elevator/Driver"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	"github.com/ellenkhoo/ElevatorProject/timers"
)

// Elevator FSM struct
type FSM struct {
	El *Elevator
	Od *ElevOutputDevice
	Fsm_mtx sync.Mutex
}

// Set all elevator lights
func (fsm *FSM) SetAllLights() {
	print("Setting all lights\n")
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := elevio.ButtonType(0); btn < N_BUTTONS-1; btn++ {
			fsm.Od.RequestButtonLight(floor, btn, fsm.El.GlobalHallRequests[floor][btn])
		}
		fsm.Od.RequestButtonLight(floor, elevio.BT_Cab, fsm.El.ElevStates.CabRequests[floor])
	}
}

// Handle initialization between floors
func (fsm *FSM) Fsm_onInitBetweenFloors() {
	fsm.Od.MotorDirection(elevio.MD_Up)
	fsm.El.Dirn = elevio.MD_Up
	fsm.El.Behaviour = EB_Moving
}

// Handle button press event
//Mulig at det trengs bedre navn
func (fsm *FSM) Fsm_onRequestsToDo(networkChannels *sharedConsts.NetworkChannels, start_timer chan time.Duration) {
	Elevator_print(*fsm.El)

	switch fsm.El.Behaviour {
	case EB_DoorOpen:
		if Requests_shouldClearImmediately(*fsm.El) {
			fmt.Println("Should clear order immediately")
			start_timer <- timers.DoorOpenDuration
			
			elevStatesJSON, err := json.Marshal(fsm.El.ElevStates)
			if err != nil {
				fmt.Println("Error marshalling elevStates: ", err)
				return
			}

			msg := sharedConsts.Message{
				Type: sharedConsts.CurrentStateMessage,
				Target: sharedConsts.TargetMaster,
				Payload: elevStatesJSON,
			}

			networkChannels.SendChan <- msg
		} else {
			fmt.Println("Shouldn't clear order immediately")
		}

	case EB_Idle:
		fsm.Fsm_mtx.Lock()
		pair := Requests_chooseDirection(*fsm.El)
		fmt.Println("Chose direction:", pair.Dirn)
		fsm.El.Dirn = pair.Dirn
		fsm.El.Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			start_timer <- timers.DoorOpenDuration
			//fsm.El = Requests_clearAtCurrentFloor(fsm.El)

		case EB_Moving:
			fsm.Od.MotorDirection(fsm.El.Dirn)

		case EB_Idle:
			// Do nothing
		}
		fsm.Fsm_mtx.Unlock()
	}

	fsm.SetAllLights()
	fmt.Println("\nNew state:")
	Elevator_print(*fsm.El)
}

// Handle floor arrival event
func (fsm *FSM) Fsm_onFloorArrival(networkChannels *sharedConsts.NetworkChannels, newFloor int, start_timer chan time.Duration) {
	fmt.Printf("\n\n(%d)\n", newFloor)
	Elevator_print(*fsm.El)

	fsm.El.ElevStates.Floor = newFloor

	elevio.SetFloorIndicator(newFloor)

	switch fsm.El.Behaviour {
	case EB_Moving:
		if Requests_shouldStop(*fsm.El) {
			fmt.Printf("Elevator stopping at floor %d \n", fsm.El.ElevStates.Floor)
			fsm.Od.MotorDirection(elevio.MD_Stop)

			fmt.Println("Going to lock fsm")
			fsm.Fsm_mtx.Lock()
			fsm.El.ElevStates.CabRequests[fsm.El.ElevStates.Floor] = false
			
			elevio.SetDoorOpenLamp(true)
			start_timer <- timers.DoorOpenDuration
			fmt.Print("Started doorOpen timer")
			fsm.El.Behaviour = EB_DoorOpen
			fmt.Println("Elevator behaviour: ", fsm.El.Behaviour)
			fsm.Fsm_mtx.Unlock()
			
			SendCurrentState(networkChannels, fsm)
		}
	}

	fmt.Println("\nNew state:")
	Elevator_print(*fsm.El)
}

// Handle door timeout event
func (fsm *FSM) Fsm_onDoorTimeout(start_timer chan time.Duration) {
	Elevator_print(*fsm.El)

	// fsm.SetAllLights()

	switch fsm.El.Behaviour {
	case EB_DoorOpen:
		pair := Requests_chooseDirection(*fsm.El)
		fsm.El.Dirn = pair.Dirn
		fsm.El.Behaviour = pair.Behaviour

		switch fsm.El.Behaviour {
		case EB_DoorOpen:
			start_timer <- timers.DoorOpenDuration
			// fsm.El = Requests_clearAtCurrentFloor(fsm.El)
			//fsm.SetAllLights()
		case EB_Moving, EB_Idle:
			elevio.SetDoorOpenLamp(false)
			fsm.Od.MotorDirection(fsm.El.Dirn)
		}

	}

	fmt.Println("\nNew state:")
	Elevator_print(*fsm.El)
}
