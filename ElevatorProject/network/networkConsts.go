package network

import (
	"net"
	"sync"

	"github.com/ellenkhoo/ElevatorProject/elevator"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

var BcastPort = "9999"
var TCPPort = "8081"

// Keeping track of connections
type MasterConnectionInfo struct {
	ClientIP string
	HostConn net.Conn
}
type ActiveConnections struct {
	mutex sync.Mutex
	Conns []MasterConnectionInfo
}
type ClientConnectionInfo struct {
	ID     string
	HostIP string
	ClientConn net.Conn
	Channels   sharedConsts.NetworkChannels
	Worldview  BackupData
	ClientMtx  sync.Mutex
}

type MasterData struct {
	GlobalHallRequests  [elevator.N_FLOORS][2]bool            `json:"globalHallRequests"`
	AllAssignedRequests map[string][elevator.N_FLOORS][2]bool `json:"allAssignedRequests"`
	AllElevStates       map[string]elevator.ElevStates        `json:"allElevStates"`
	BackupData          BackupData
	mutex               sync.Mutex
}

type BackupData struct {
	GlobalHallRequests  [elevator.N_FLOORS][2]bool            `json:"globalHallRequests"`
	AllAssignedRequests map[string][elevator.N_FLOORS][2]bool `json:"allAssignedRequests"`
}

type ElevatorRequest struct {
	GlobalHallRequests [elevator.N_FLOORS][2]bool `json:"globalHallRequests"`
	AssignedRequests   [elevator.N_FLOORS][2]bool `json:"assignedRequests"`
}
