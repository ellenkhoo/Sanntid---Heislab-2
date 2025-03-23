package elevator

import (
	"fmt"
	"time"

	"github.com/ellenkhoo/ElevatorProject/elevator/Driver"
	"github.com/ellenkhoo/ElevatorProject/timers"
)

//tar en elevatorBehaviour-verdi som argument og returnerer
//en peker til en streng som representerer navnet på veriden

//eb_toString-funksjonen tar en ElevatorBehaviour-verdi som input
//den bruker operatoren "? :" til å sammenligne verdien mot kjente
//enum-verdier (EB_Idle, EB_DoorOpen, EB_Moving)

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota 
	EB_DoorOpen
	EB_Moving
)


func ElevatorBehaviourToString(behaviour ElevatorBehaviour) string {
	switch behaviour {
	case EB_Idle:
		return "idle"
	case EB_DoorOpen:
		return "doorOpen"
	case EB_Moving:
		return "moving"
	default:
		return "unknown"
	}
}

func MotorDirectionToString(direction elevio.MotorDirection) string {
	switch direction {
	case elevio.MD_Up:
		return "up"
	case elevio.MD_Down:
		return "down"
	case elevio.MD_Stop:
		return "stop"
	default:
		return "unknown"
	}	
}	

func FormatElevStates(elevator *Elevator, ElevStates *ElevStates) {
	behaviourStr := ElevatorBehaviourToString(elevator.Behaviour)
	directionStr := MotorDirectionToString(elevator.Dirn)

	ElevStates.Behaviour = behaviourStr
	ElevStates.Direction = directionStr
}

type ElevStates struct {
	Behaviour   string `json:"behaviour"`
	Floor       int		`json:"floor"`
	Direction   string	`json:"direction"`
	CabRequests [N_FLOORS]bool	`json:"cabRequests"`
	IP          string	`json:"ip"`
}

type MessageToMaster struct {
	ElevStates ElevStates
	RequestsToDo [N_FLOORS][N_BUTTONS]bool
}

// elevator struct representerer statene til heisen
type Elevator struct {
	ElevStates 		 *ElevStates
	PrevFloor        int
	Dirn             elevio.MotorDirection
	Behaviour        ElevatorBehaviour
	GlobalHallRequests     [N_FLOORS][N_BUTTONS - 1]bool
	AssignedRequests [N_FLOORS][N_BUTTONS - 1]bool
	RequestsToDo     [N_FLOORS][N_BUTTONS]bool //CabRequests + AssignedRequests
	Config           ElevatorConfig
	//State HRAElevState
}

const (
	N_FLOORS   = 4
	N_BUTTONS  = 3
	B_HallUp   = 0
	B_HallDown = 1
	B_Cab      = 2
)

// Nytt, usikkert
type ElevatorOrder struct {
	Order      elevio.ButtonEvent
	ElevatorIP int
}

//

func Elevator_print(e Elevator) {
	fmt.Println(" +-----------------+")
	fmt.Printf("|floor = %-2d          |\n", e.ElevStates.Floor)
	fmt.Printf("  |dirn  = %-12.12s|\n", e.Dirn)
	fmt.Printf("  |behav = %-12.12s|\n", e.Behaviour)
	fmt.Println(" +-----------------+")
	fmt.Println("  |  | up  | dn  | cab |")

	for f := N_FLOORS - 1; f >= 0; f-- {
		fmt.Printf("| %d", f)
		for btn := 0; btn < N_BUTTONS; btn++ {
			if (f == N_FLOORS-1 && btn == B_HallUp) || (f == 0 && btn == B_HallDown) {
				fmt.Print("|       ")
			} else {
				if e.RequestsToDo[f][btn] {
					fmt.Print("|   #   ")
				} else {
					fmt.Print("|   -   ")
				}
			}
		}
		fmt.Println("|")
	}
	fmt.Println(" +-----------------------+")
}

// definierer konfigurasjonsstruktur
type ElevatorConfig struct {
	ClearRequestVariant string
	DoorOpenDuration    time.Duration
}

// funksjonen for å returnere en uinitialisert heis
func Elevator_uninitialized() *Elevator {
	return &Elevator{
		ElevStates: &ElevStates{
			Behaviour: "idle",
			Floor: -1,
			Direction: "stop",       
			CabRequests: [N_FLOORS]bool{false, false, false, false},
			IP: "0.0.0.0",  
			
		},           
		Dirn:             elevio.MD_Stop, 
		Behaviour:        EB_Idle,      
		GlobalHallRequests:     [N_FLOORS][N_BUTTONS - 1]bool{
			{false, false}, 
			{false, false}, 
			{false, false}, 
			{false, false}},
		RequestsToDo:     [N_FLOORS][N_BUTTONS]bool{
			{false, false, false},  
			{false, false, false},
			{false, false, false},
			{false, false, false},
		},
		Config: ElevatorConfig{
			ClearRequestVariant: "CV_InDirn",    
			DoorOpenDuration: timers.DoorOpenDuration,
		},
	}
}

