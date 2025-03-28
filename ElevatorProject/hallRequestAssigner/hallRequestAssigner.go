package hra

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/ellenkhoo/ElevatorProject/elevator"
)

type HRAElevState struct {
	Behaviour   string                  `json:"behaviour"`
	Floor       int                     `json:"floor"`
	Direction   string                  `json:"direction"`
	CabRequests [elevator.N_FLOORS]bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [elevator.N_FLOORS][2]bool `json:"hallRequests"`
	States       map[string]HRAElevState    `json:"states"`
}

func SendStateToHRA(allElevStates map[string]elevator.ElevStates, globalHallRequest [elevator.N_FLOORS][2]bool) *map[string][elevator.N_FLOORS][2]bool {
	inputFormatHRA := make(map[string]HRAElevState)
	for id, state := range allElevStates {
		inputFormatHRA[fmt.Sprintf("%s", id)] = HRAElevState{
			Behaviour:   state.Behaviour,
			Floor:       state.CurrentFloor,
			Direction:   state.Direction,
			CabRequests: state.CabRequests,
		}
	}

	input := HRAInput{
		HallRequests: globalHallRequest,
		States:       inputFormatHRA,
	}

	// Marshal input
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return nil
	}

	// Checks what executable should be used
	var cmd string
	os := runtime.GOOS
	if os == "windows" {
		cmd = "./hallRequestAssigner/hall_request_assigner.exe"
	} else {
		cmd = "./hallRequestAssigner/hall_request_assigner"
	}

	ret, err := exec.Command(cmd, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return nil
	}

	output := new(map[string][elevator.N_FLOORS][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return nil
	}

	return output
}
