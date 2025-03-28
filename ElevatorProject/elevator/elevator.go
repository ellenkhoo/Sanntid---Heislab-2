package elevator

import (
	"time"

	"github.com/ellenkhoo/ElevatorProject/timers"
)

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

const (
	N_FLOORS   = 4
	N_BUTTONS  = 3
	B_HallUp   = 0
	B_HallDown = 1
	B_Cab      = 2
)

type ElevStates struct {
	Behaviour    string         `json:"behaviour"`
	CurrentFloor int            `json:"floor"`
	Direction    string         `json:"direction"`
	CabRequests  [N_FLOORS]bool `json:"cabRequests"`
	ID           string         `json:"ip"`
}

type MessageToMaster struct {
	ElevStates   ElevStates
	RequestsToDo [N_FLOORS][N_BUTTONS]bool
}

type Elevator struct {
	ElevStates         *ElevStates
	PrevFloor          int
	Dirn               MotorDirection
	Behaviour          ElevatorBehaviour
	GlobalHallRequests [N_FLOORS][N_BUTTONS - 1]bool
	AssignedRequests   [N_FLOORS][N_BUTTONS - 1]bool
	RequestsToDo       [N_FLOORS][N_BUTTONS]bool //CabRequests + AssignedRequests
	DoorOpenDuration   time.Duration
}

func InitializeElevator() *Elevator {
	return &Elevator{
		ElevStates: &ElevStates{
			Behaviour:    "idle",
			CurrentFloor: -1,
			Direction:    "stop",
			CabRequests:  [N_FLOORS]bool{false, false, false, false},
			ID:           "0.0.0.0",
		},
		Dirn:      MD_Stop,
		Behaviour: EB_Idle,
		GlobalHallRequests: [N_FLOORS][N_BUTTONS - 1]bool{
			{false, false},
			{false, false},
			{false, false},
			{false, false}},
		RequestsToDo: [N_FLOORS][N_BUTTONS]bool{
			{false, false, false},
			{false, false, false},
			{false, false, false},
			{false, false, false},
		},
		DoorOpenDuration: timers.DoorOpenDuration,
	}
}
