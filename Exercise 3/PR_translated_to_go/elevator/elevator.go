package elevatorpkg

import (
	"Driver-go/elevio"
	// "elevator_io_device"
	"fmt"
	"time"
)

//tar en elevatorBehaviour-verdi som argument og returnerer
//en peker til en streng som representerer navnet på veriden

//eb_toString-funksjonen tar en ElevatorBehaviour-verdi som input
//den bruker operatoren "? :" til å sammenligne verdien mot kjente
//enum-verdier (EB_Idle, EB_DoorOpen, EB_Moving)

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota //iota brukes for å forenkle
	//opprettelsen av sekvensielle verdier(starter på null
	//og øker automatisk med 1 for hver ny konstant i blokken)
	EB_DoorOpen
	EB_Moving
)

var ElevatorBehaviourToString = map[ElevatorBehaviour]string{
	EB_Idle:     "EB_Idle",
	EB_DoorOpen: "EB_DoorOpen",
	EB_Moving:   "EB_Moving",
}

func Eb_toString(eb ElevatorBehaviour) string {
	if str, exists := ElevatorBehaviourToString[eb]; exists {
		return str
	}
	return "EB_UNDEFINED"
}

//
//funksjonen elevator_print skriver ut en visuell representasjon av tilstanden til en
//heis. den viser informasjon som; heisens etasje, retning, (elevio_dirn_toString), behavior(eb_toString)
//en tabell som viser bestillinger for hver etasje og knappetype
//funksjonen skriver ut bestillinger for hver etasje i fallende rekkefølge

// elevator struct representerer statene til heisen
type Elevator struct {
	Floor     int
	Dirn      elevio.MotorDirection
	Behaviour ElevatorBehaviour
	Requests  [N_FLOORS][N_BUTTONS]bool
	Config    ElevatorConfig
}

const (
	N_FLOORS   = 4
	N_BUTTONS  = 3
	B_HallUp   = 0
	B_HallDown = 1
	B_Cab      = 2
)

func Elevator_print(e Elevator) {
	fmt.Println(" +-----------------+")
	fmt.Printf("|floor = %-2d          |\n", e.Floor)
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
				if e.Requests[f][btn] {
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

//
//funksjonen elevator_uninitialized oppretter og returnerer en uinitialisert heis
//(en instans av elevator-structen) med standard verdier

// definierer konfigurasjonsstruktur
type ElevatorConfig struct {
	ClearRequestVariant string
	DoorOpenDuration    time.Duration
}

// funksjonen for å returnere en uinitialisert heis
func Elevator_uninitialized() Elevator {
	return Elevator{
		Floor:     -1,      //ugyldug etasje
		Dirn:      elevio.MD_Stop,  //heisen er stoppet
		Behaviour: EB_Idle, //inaktiv tilstand
		Config: ElevatorConfig{
			ClearRequestVariant: "CV_All", //fjerner alle forespørsler
			DoorOpenDuration:    3.0,      //3 sekunder døråpning
		},
	}
}

/*
func main() {
	e := Elevator{
		Floor:     2,
		Dirn:      "UP",
		Behaviour: EB_Moving,
		Requests: [N_FLOORS][N_BUTTONS]bool{
			{false, false, true},
			{true, false, false},
			{false, true, false},
			{false, false, false},
		},
	}
	elevatorPrint(e)

	elevator := ElevatorUninitialized()
	fmt.Println(elevator)
}
*/
