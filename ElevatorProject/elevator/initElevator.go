package elevator

import (
	elevio "github.com/ellenkhoo/ElevatorProject/elevator/Driver"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	"github.com/ellenkhoo/ElevatorProject/timers"
	"fmt"
)

func InitElevator(localIP string, networkChannels *sharedConsts.NetworkChannels) *FSM {
	elevio.Init("localhost:15657", N_FLOORS)

	fsm := FSM{El: InitializeElevator(), Od: GetOutputDevice()}
	fsm.Fsm_mtx.Lock()
	fsm.El.ElevStates.IP = localIP
	fsm.Fsm_mtx.Unlock()

	fmt.Println("Behavoiur: ", fsm.El.ElevStates.Behaviour)

	go RunElevator(networkChannels, &fsm, timers.MaxDuration)

	return &fsm
}
