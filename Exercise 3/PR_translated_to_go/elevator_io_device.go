package main

import "fmt"

//definierer enum for Button og Dirn
type Button int
type Dirn int

const (
	//knappetyper, overflødig?
	// B_HallUp Button = iota
	// B_HallDown
	// B_Cab

	//retningstyper starter på ny iota-sekvens
	D_Up Dirn = iota
	D_Down
	D_Stop
)

//funksjonen som stimulerer initialisering av heismaskinvare
func init() {
	elevator_hardware_init()
}

//simulerer initialisering av heisens maskinvare
func elevator_hardware_init() {
	fmt.Println("Initialising elevator hardware...")
	//her kan du vi legge til kode for maskinvareinitialisering
}

//funksjonen som simulerer å hente signalet fra en knapp på en spesifikk etasje
func wrap_request_button(f int, b Button) int {
	return elevator_hardware_get_button_signal(b, f)
}

//simulerer å hente knappesignal for en spesifikk etasje og knapp
func elevator_hardware_get_button_signal(b Button, f int) int {
	fmt.Printf("Getting button signal for floor %d and button %d\n", f, b)
	return 1 //simulerer at signalet er aktivert
}

//funksjonen for å sette en lampe på en knapp (tillater å sette den på eller av)
func wrap_request_button_light(f int, b Button, v int) {
	elevator_hardware_set_button_lamp(b, f, v)
}

//simulerer å sette lampen til en spesifikk verdi for en knapp på en etasje
func elevator_hardware_set_button_lamp(b Button, f int, v int) {
	fmt.Printf("Setting button light for florr %d, buttin %d, value %d\n", f, b, v)
	//her kan vi legge til kode som setter lampeverdien på en knapp
}

//funksjonen som simulerer å sette motorretning
func wrap_motor_direction(d Dirn) {
	elevator_hardware_set_motor_direction(d)
}

//simulerer å sette motorretningen til heisen
func elevator_hardware_set_motor_direction(d Dirn) {
	fmt.Printf("Setting motor direction to %d\n", d)
	//her kan vi legge til kode som setter motorens retning
}

//definerer en sturktur for å holde funksjoner relatert til heisinndata
type ElevInputDevice struct {
	FloorSensor   func() int
	RequestButton func(florr int, button Button) int
	Obstruction   func() bool
}

//simulerte funksjoner for å etterligne maskinvarefunksjonene
func elevator_hardware_get_floor_sensor_signal() int {
	//simulerer deteksjon av etasjesensor
	return 1 //eks. heisen er i etasje 1
}

// func wrap_request_button(floor int, button Button) int {
// 	//simulerer knappetrykk
// 	return 0 //eks. ingen knappetrykk registrert
// }

func elevator_hardware_get_obstruction_signal() bool {
	//simulerer hindringssensor
	return false //eks. ingen hindring oppdaget
}

//funksjon for å returnere en instans av ElevInputDevice
func elevio_getInputDevice() ElevInputDevice {
	return ElevInputDevice{
		FloorSensor:   elevator_hardware_get_floor_sensor_signal,
		RequestButton: wrap_request_button,
		Obstruction:   elevator_hardware_get_obstruction_signal,
	}
}

//elevOutputDevice stuct holds function pointers for controlling elevator output devices
type ElevOutputDevice struct {
	FloorIndicator     func(floor int)
	RequestButtonLight func(floor int, button Button, value int)
	DoorLight          func(value int)
	StopButtonLight    func(value int)
	MotorDirection     func(direvtion Dirn)
}

//simulated hardware functios to mimic the actual hardware control
func elevator_hardware_set_floor_indicator(floor int) {
	//simulate setting the floor indicator lamp
	println("Floor indicator set to:", floor)
}

// func wrap_request_button_light(floor int, button Button, value int){
// 	//simulate turning on/off the request button light
// 	println("Request button light on floor", floor, "button", button, "set to:", value)
// }

func elevator_hardware_set_door_open_lamp(value int) {
	//simulate setting the door open light
	print("Door light set to:", value)
}

func elevator_hardware_set_stop_lamp(value int) {
	//simulate setting the stop lamp
	println("Stop button light set to:", value)
}

// func wrap_motor_direction(direction Dirn){
// 	//simulate setting motor direction
// 	println("Motor direction set to:", direction)
// }

//function to return an instance of elevOutputDevice with function assignments
func elevio_getOutputDevice() ElevOutputDevice {
	return ElevOutputDevice{
		FloorIndicator:     elevator_hardware_set_floor_indicator,
		RequestButtonLight: wrap_request_button_light,
		DoorLight:          elevator_hardware_set_door_open_lamp,
		StopButtonLight:    elevator_hardware_set_stop_lamp,
		MotorDirection:     wrap_motor_direction,
	}
}

//mapper knapper og retninger til strenger
var ButtonToString = map[Button]string{
	B_HallUp:   "B_HallUp",
	B_HallDown: "B_HallDown",
	B_Cab:      "B_Cab",
}

var dirnToString = map[Dirn]string{
	D_Up:   "D_Up",
	D_Down: "D_Down",
	D_Stop: "D_Stop",
}

var stringToDirn = map[string]Dirn{
	"D_Up":   D_Up,
	"D_Down": D_Down,
	"D_Stop": D_Stop,
}

//funksjoner for å hente tilsvarende string
func elevio_button_toString(b Button) string {
	if str, exists := ButtonToString[b]; exists {
		return str
	}
	return "D_UNDEFINED"
}

func elevio_drin_toString(d Dirn) string {
	if str, exists := dirnToString[d]; exists {
		return str
	}
	return "D_UNDEFINED"
}

//eks. i main
/*
func main() {
    inputDevice := GetInputDevice()

    fmt.Println("Floor Sensor:", inputDevice.FloorSensor())
    fmt.Println("Request Button:", inputDevice.RequestButton(2, B_HallUp))
    fmt.Println("Obstruction:", inputDevice.Obstruction())

	outputDevice := GetOutputDevice()

    outputDevice.FloorIndicator(3)
    outputDevice.RequestButtonLight(2, B_HallUp, 1)
    outputDevice.DoorLight(1)
    outputDevice.StopButtonLight(0)
    outputDevice.MotorDirection(DirnUp)
}
*/
