package network

import (
	"net"
	"sync"

	"github.com/ellenkhoo/ElevatorProject/elevator"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

var BcastPort = "9999"
var TCPPort = "8085"

type MasterConnectionInfo struct {
	ClientID string
	HostConn net.Conn
}

type ActiveConnections struct {
	AC_mutex sync.Mutex
	Conns    []MasterConnectionInfo
}

type ClientInfo struct {
	ID               string
	HostID           string
	ClientConn       net.Conn
	Channels         sharedConsts.NetworkChannels
	BackupData       BackupData
	ClientInfo_mutex sync.Mutex
}

type MasterData struct {
	GlobalHallRequests  [elevator.N_FLOORS][2]bool            `json:"globalHallRequests"`
	AllAssignedRequests map[string][elevator.N_FLOORS][2]bool `json:"allAssignedRequests"`
	AllElevStates       map[string]elevator.ElevStates        `json:"allElevStates"`
	BackupData          GlobalRequestsWorldview
	MasterData_mutex    sync.Mutex
}

type GlobalRequestsWorldview struct {
	GlobalHallRequests  [elevator.N_FLOORS][2]bool            `json:"globalHallRequests"`
	AllAssignedRequests map[string][elevator.N_FLOORS][2]bool `json:"allAssignedRequests"`
}

type BackupData struct {
	GlobalRequestsWorldview     GlobalRequestsWorldview
	MastersActiveConnectionsIDs []string
}

type LocalRequestsWorldview struct {
	GlobalHallRequests [elevator.N_FLOORS][2]bool `json:"globalHallRequests"`
	AssignedRequests   [elevator.N_FLOORS][2]bool `json:"assignedRequests"`
}

type CabRequestsWithID struct {
	ID          string                  `json:"id"`
	CabRequests [elevator.N_FLOORS]bool `json:"cabRequests"`
}
