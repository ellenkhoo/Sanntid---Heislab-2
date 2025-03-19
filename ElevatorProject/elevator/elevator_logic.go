package elevator

import (
	"github.com/ellenkhoo/ElevatorProject/elevator/Driver"
	"github.com/ellenkhoo/ElevatorProject/comm"
	"github.com/ellenkhoo/ElevatorProject/timers"
	"github.com/ellenkhoo/ElevatorProject/network"
	"fmt"
	"time"
	"net"
)

func ElevLogic_runElevator(fsm FSM, maxDuration time.Duration)  {

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

	Clear_all_requests(fsm.El)
	fsm.SetAllLights()

	if elevio.GetFloor() == -1 {
		fsm.Fsm_onInitBetweenFloors()
		fmt.Printf("Init between floor")
	}

	for {
		select {

		//Lag melding, send den på sendChan
		case order := <-buttons_chan:
			fmt.Printf("Button pushed. Order at floor: %d\n", order.Floor)
			// If cab call
			if order.Button == B_Cab {
				fsm.El.ElevStates.CabRequests[order.Floor] = true
			} else { //Send hall calls til master
				reqMsg := network.Message{
					Type: localRequestMessage,
					Target: TargetMaster,
					Payload: order
				}
				network.networkChannels.sendChan <- reqmsg
			}
			//Alltid send state etter et knappetrykk
			stateMsg := network.Message{
				Type: currentStateMessage,
				Target: TargetMaster,
				Payload: fsm.El.ElevStates
			}
			network.NetworkChannels.sendChan <- stateMsg


		case floor_input := <-floors_chan:
			fmt.Printf("Floor sensor: %d\n", floor_input)

			if floor_input != -1 && floor_input != fsm.El.ElevStates.Floor {
				//Master informeres i funksjonskallet nedenfor
				fsm.Fsm_onFloorArrival(floor_input, start_timer)
			}


		case obstruction := <-obstruction_chan:
			if obstruction {
				if fsm.El.Behaviour == EB_DoorOpen {
					start_timer <- maxDuration
				}
			} else {
				start_timer <- fsm.El.Config.DoorOpenDuration
			}
			msg := network.Message{
				Type: currentStateMessage,
				Target: TargetMaster,
				Payload: fsm.El.ElevStates
			}

			network.networkChannels.sendChan <- msg

		case <-timer.C:
			fsm.Fsm_onDoorTimeout(start_timer)
			comm.Comm_sendCurrentState(fsm.El.ElevStates, conn)
		}
		
	}
}
