package elevator

import (
	"fmt"

	elevio "github.com/ellenkhoo/ElevatorProject/elevator/Driver"
)

type Button elevio.ButtonType

type Dirn int

// Formatting
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
	fmt.Println("ElevStates after formatting:", ElevStates)
}

// Output device
type ElevOutputDevice struct {
	FloorIndicator     func(floor int)
	RequestButtonLight func(floor int, button elevio.ButtonType, value bool)
	DoorLight          func(value int)
	StopButtonLight    func(value int)
	MotorDirection     func(direction elevio.MotorDirection)
}

func HardwareSetFloorIndicator(floor int) {
	println("Floor indicator set to:", floor)
}

func WrapRequestButtonLight(f int, b elevio.ButtonType, v bool) {
	HardwareSetButtonLamp(b, f, v)
}

func HardwareSetButtonLamp(b elevio.ButtonType, f int, v bool) {
	elevio.SetButtonLamp(b, f, v)
}

func HardwareSetDoorOpenLamp(value int) {
	print("Door light set to:", value)
}

func HardwareSetStopLamp(value int) {
	println("Stop button light set to:", value)
}

func WrapMotorDirection(d elevio.MotorDirection) {
	HardwareSetMotorDirection(d)
}

func HardwareSetMotorDirection(d elevio.MotorDirection) {
	fmt.Printf("Setting motor direction to %d\n", d)
	elevio.SetMotorDirection(d)
}

func GetOutputDevice() *ElevOutputDevice {
	return &ElevOutputDevice{
		FloorIndicator:     HardwareSetFloorIndicator,
		RequestButtonLight: WrapRequestButtonLight,
		DoorLight:          HardwareSetDoorOpenLamp,
		StopButtonLight:    HardwareSetStopLamp,
		MotorDirection:     WrapMotorDirection,
	}
}
