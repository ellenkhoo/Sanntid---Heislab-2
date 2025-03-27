package network

import (
	"net"
	"sync"

	"github.com/ellenkhoo/ElevatorProject/elevator"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

var BcastPort = "9999"
var TCPPort = "8088"

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
	ID         string
	HostIP     string
	ClientConn net.Conn
	Channels   sharedConsts.NetworkChannels
	BackupData BackupData
	ClientMtx  sync.Mutex
}

type MasterData struct {
	GlobalHallRequests  [elevator.N_FLOORS][2]bool            `json:"globalHallRequests"`
	AllAssignedRequests map[string][elevator.N_FLOORS][2]bool `json:"allAssignedRequests"`
	AllElevStates       map[string]elevator.ElevStates        `json:"allElevStates"`
	BackupData          Worldview
	mutex               sync.Mutex
}

type Worldview struct {
	GlobalHallRequests  [elevator.N_FLOORS][2]bool            `json:"globalHallRequests"`
	AllAssignedRequests map[string][elevator.N_FLOORS][2]bool `json:"allAssignedRequests"`
}

type BackupData struct {
	Worldview                   Worldview
	MastersActiveConnectionsIPs []string
}

type ElevatorRequest struct {
	GlobalHallRequests [elevator.N_FLOORS][2]bool `json:"globalHallRequests"`
	AssignedRequests   [elevator.N_FLOORS][2]bool `json:"assignedRequests"`
}

type CabRequestsWithID struct {
	ID          string                  `json:"id"`
	CabRequests [elevator.N_FLOORS]bool `json:"cabRequests"`
}
