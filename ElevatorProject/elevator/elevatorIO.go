package elevator

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const _pollRate = 20 * time.Millisecond

var _initialized bool = false
var _numFloors int
var _elevioMtx sync.Mutex
var _conn net.Conn

type Button ButtonType

type Dirn int

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ButtonType int

type ButtonEvent struct {
	Floor  int        `json:"floor"`
	Button ButtonType `json:"button"`
}

func InitializeElevatorDriver(addr string, numFloors int) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_numFloors = numFloors
	_elevioMtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}

func SetMotorDirection(dir MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button ButtonType, floor int, value bool) {
	write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}

func PollButtons(receiver chan<- ButtonEvent) {
	prev := make([][3]bool, _numFloors)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < _numFloors; f++ {
			for b := ButtonType(0); b < 3; b++ {
				v := GetButton(b, f)
				if v != prev[f][b] && v != false {
					receiver <- ButtonEvent{f, ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := GetFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func GetButton(button ButtonType, floor int) bool {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

func GetFloor() int {
	a := read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

func GetStop() bool {
	a := read([4]byte{8, 0, 0, 0})
	return toBool(a[1])
}

func GetObstruction() bool {
	a := read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}

func read(in [4]byte) [4]byte {
	_elevioMtx.Lock()
	defer _elevioMtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

func write(in [4]byte) {
	_elevioMtx.Lock()
	defer _elevioMtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}

// Output device
type ElevOutputDevice struct {
	FloorIndicator     func(floor int)
	RequestButtonLight func(floor int, button ButtonType, value bool)
	DoorLight          func(value int)
	StopButtonLight    func(value int)
	MotorDirection     func(direction MotorDirection)
}

func HardwareSetFloorIndicator(floor int) {
	println("Floor indicator set to:", floor)
}

func WrapRequestButtonLight(f int, b ButtonType, v bool) {
	HardwareSetButtonLamp(b, f, v)
}

func HardwareSetButtonLamp(b ButtonType, f int, v bool) {
	SetButtonLamp(b, f, v)
}

func HardwareSetDoorOpenLamp(value int) {
	print("Door light set to:", value)
}

func HardwareSetStopLamp(value int) {
	println("Stop button light set to:", value)
}

func WrapMotorDirection(d MotorDirection) {
	HardwareSetMotorDirection(d)
}

func HardwareSetMotorDirection(d MotorDirection) {
	fmt.Printf("Setting motor direction to %d\n", d)
	SetMotorDirection(d)
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

func MotorDirectionToString(direction MotorDirection) string {
	switch direction {
	case MD_Up:
		return "up"
	case MD_Down:
		return "down"
	case MD_Stop:
		return "stop"
	default:
		return "unknown"
	}
}

func FormatElevStates(elevator Elevator) MessageToMaster {
	behaviourStr := ElevatorBehaviourToString(elevator.Behaviour)
	directionStr := MotorDirectionToString(elevator.Dirn)
	elevator.ElevStates.Behaviour = behaviourStr
	elevator.ElevStates.Direction = directionStr
	fmt.Println("ElevStates after formatting:", elevator.ElevStates)

	UpdatedElevStates := ElevStates{
		Behaviour:   behaviourStr,
		CurrentFloor:       elevator.ElevStates.CurrentFloor,
		Direction:   directionStr,
		CabRequests: elevator.ElevStates.CabRequests,
		ID:          elevator.ElevStates.ID,
	}

	return MessageToMaster{
		ElevStates:   UpdatedElevStates,
		RequestsToDo: elevator.RequestsToDo,
	}
}
