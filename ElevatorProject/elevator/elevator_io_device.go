package elevator

import (
	"github.com/ellenkhoo/ElevatorProject/elevator/Driver"
	"fmt"
)

type Button elevio.ButtonType
type Dirn int

/*
const (
	B_HallUp   Button = 0
	B_HallDown Button = 1
	B_Cab      Button = 2

	D_Up   Dirn = 1
	D_Down Dirn = -1
	D_Stop Dirn = 0
)*/

var DirnToString = map[Dirn]string{
	Dirn(elevio.MD_Up):   "D_Up",
	Dirn(elevio.MD_Down): "D_Down",
	Dirn(elevio.MD_Stop): "D_Stop",
}

var StringToDirn = map[string]Dirn{
	"D_Up":   Dirn(elevio.MD_Up),
	"D_Down": Dirn(elevio.MD_Down),
	"D_Stop": Dirn(elevio.MD_Stop),
}


//Overflødig?
func Init() {
	Elevator_hardware_init()
}

func Elevator_hardware_init() {
	fmt.Println("Initialising elevator hardware...")
}

func Wrap_request_button(f int, b Button) int {
	return Elevator_hardware_get_button_signal(b, f)
}

func Elevator_hardware_get_button_signal(b Button, f int) int {
	fmt.Printf("Getting button signal for floor %d and button %d\n", f, b)
	return 1
}


func Wrap_request_button_light(f int, b elevio.ButtonType, v bool) {
	Elevator_hardware_set_button_lamp(b, f, v)
}

func Elevator_hardware_set_button_lamp(b elevio.ButtonType, f int, v bool) {
	elevio.SetButtonLamp(b, f, v)
}

func Wrap_motor_direction(d elevio.MotorDirection) {
	Elevator_hardware_set_motor_direction(d)
}

func Elevator_hardware_set_motor_direction(d elevio.MotorDirection) {
	fmt.Printf("Setting motor direction to %d\n", d)
	elevio.SetMotorDirection(d)
}

type ElevInputDevice struct {
	FloorSensor   func() int
	RequestButton func(floor int, button Button) int
	Obstruction   func() bool
}

func Elevator_hardware_get_floor_sensor_signal() int {
	return 1
}

func Elevator_hardware_get_obstruction_signal() bool {
	return false 
}

func Elevio_getInputDevice() ElevInputDevice {
	return ElevInputDevice{
		FloorSensor:   Elevator_hardware_get_floor_sensor_signal,
		RequestButton: Wrap_request_button,
		Obstruction:   Elevator_hardware_get_obstruction_signal,
	}
}

type ElevOutputDevice struct {
	FloorIndicator     func(floor int)
	RequestButtonLight func(floor int, button elevio.ButtonType, value bool)
	DoorLight          func(value int)
	StopButtonLight    func(value int)
	MotorDirection     func(direction elevio.MotorDirection)
}


func Elevator_hardware_set_floor_indicator(floor int) {
	println("Floor indicator set to:", floor)
}

func Elevator_hardware_set_door_open_lamp(value int) {
	print("Door light set to:", value)
}

func Elevator_hardware_set_stop_lamp(value int) {
	println("Stop button light set to:", value)
}

func Elevio_getOutputDevice() *ElevOutputDevice {
	return &ElevOutputDevice{
		FloorIndicator:     Elevator_hardware_set_floor_indicator,
		RequestButtonLight: Wrap_request_button_light,
		DoorLight:          Elevator_hardware_set_door_open_lamp,
		StopButtonLight:    Elevator_hardware_set_stop_lamp,
		MotorDirection:     Wrap_motor_direction,
	}
}

var ButtonToString = map[Button]string{
	Button(elevio.BT_HallUp):   "B_HallUp",
	Button(elevio.BT_HallDown): "B_HallDown",
	Button(elevio.BT_Cab):      "B_Cab",
}



func Elevio_button_toString(b Button) string {
	if str, exists := ButtonToString[b]; exists {
		return str
	}
	return "D_UNDEFINED"
}

func Elevio_drin_toString(d Dirn) string {
	if str, exists := DirnToString[d]; exists {
		return str
	}
	return "D_UNDEFINED"
}