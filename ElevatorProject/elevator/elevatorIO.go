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

type Button ButtonType

type ButtonType int

type ButtonEvent struct {
	Floor  int        `json:"floor"`
	Button ButtonType `json:"button"`
}

func SetMotorDirection(dir MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(floor int, button ButtonType, value bool) {
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
		for floor := 0; floor < _numFloors; floor++ {
			for button := ButtonType(0); button < 3; button++ {
				value := GetButton(button, floor)
				if value != prev[floor][button] && value != false {
					receiver <- ButtonEvent{floor, ButtonType(button)}
				}
				prev[floor][button] = value
			}
		}
	}
}

func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		value := GetFloor()
		if value != prev && value != -1 {
			receiver <- value
		}
		prev = value
	}
}

func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		value := GetStop()
		if value != prev {
			receiver <- value
		}
		prev = value
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		value := GetObstruction()
		if value != prev {
			receiver <- value
		}
		prev = value
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
	RequestButtonLight func(floor int, button ButtonType, value bool)
	DoorLight          func(value bool)
	StopButtonLight    func(value bool)
	MotorDirection     func(direction MotorDirection)
}

func GetOutputDevice() *ElevOutputDevice {
	return &ElevOutputDevice{
		RequestButtonLight: SetButtonLamp,
		DoorLight:          SetDoorOpenLamp,
		StopButtonLight:    SetStopLamp,
		MotorDirection:     SetMotorDirection,
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

	UpdatedElevStates := ElevStates{
		Behaviour:    behaviourStr,
		CurrentFloor: elevator.ElevStates.CurrentFloor,
		Direction:    directionStr,
		CabRequests:  elevator.ElevStates.CabRequests,
		ID:           elevator.ElevStates.ID,
	}

	return MessageToMaster{
		ElevStates:   UpdatedElevStates,
		RequestsToDo: elevator.RequestsToDo,
	}
}
