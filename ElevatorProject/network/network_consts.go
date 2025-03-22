package network

import (
	"net"
	"sync"
	"time"

	"github.com/ellenkhoo/ElevatorProject/elevator"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

// // NETWORK CONSTS
// type HelloMsg struct {
// 	Message string
// 	Iter    int
// }

// type MessageType int

// const (
// 	MasterOrdersMessage MessageType = iota
// 	BackupAcknowledgeMessage
// 	LocalRequestMessage
// 	CurrentStateMessage
// 	HelloMessage
// 	RankMessage
// 	ElevClearedOrderMessage
// )

// type MessageTarget int

// const (
// 	TargetMaster MessageTarget = iota
// 	TargetClient
// 	TargetBackup
// 	TargetElevator
// )

// type Message struct {
// 	Type    MessageType
// 	Target  MessageTarget
// 	Payload interface{}
// }

// Keeping track of connections
type MasterConnectionInfo struct {
	ClientIP string
	Rank     int
	HostConn net.Conn
	HeartbeatTimer *time.Timer
}

type ClientConnectionInfo struct {
	ID          string
	HostIP      string
	Rank        int
	ClientConn  net.Conn
	Channels sharedConsts.NetworkChannels
	HeartbeatTimer *time.Timer
}

// type NetworkChannels struct {
// 	SendChan 	chan Message
// 	ReceiveChan chan Message
// 	MasterChan   chan Message
// 	BackupChan   chan Message
// 	ElevatorChan chan Message
// }

type ActiveConnections struct {
	mutex sync.Mutex
	Conns []MasterConnectionInfo
	//ClientTimers map[string]*time.Timer
}

// type MasterData struct {
// 	GlobalHallRequests [][2]bool `json:"globalHallRequests"`
// 	AllAssignedRequests   map[string][][2]bool	`json:"allAssignedRequests"`
// 	AllElevStates map[string]elevator.ElevStates	`json:"allElevStates"`
// 	mutex sync.Mutex
// }

type MasterData struct {
	GlobalHallRequests [elevator.N_FLOORS][2]bool `json:"globalHallRequests"`
	AllAssignedRequests   map[string][elevator.N_FLOORS][2]bool	`json:"allAssignedRequests"`
	AllElevStates map[string]elevator.ElevStates	`json:"allElevStates"`
	mutex sync.Mutex
	//HeartbeatTimer *time.Timer
}

type BackupData struct {
	GlobalHallRequests [elevator.N_FLOORS][2]bool `json:"globalHallRequests"`
	AllAssignedRequests   map[string][elevator.N_FLOORS][2]bool	`json:"allAssignedRequests"`
}

type ElevatorRequest struct {
	GlobalHallRequests [elevator.N_FLOORS][2]bool `json:"globalHallRequests"`
	AssignedRequests   [elevator.N_FLOORS][2]bool `json:"assignedRequests"`
}

