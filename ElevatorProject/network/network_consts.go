package network

import (
	"net"
	"sync"
)

type HelloMsg struct {
	Message string
	Iter    int
}

type MessageType int

const (
	masterOrdersMessage MessageType = iota
	backupAcknowledgeMessage
	localRequestMessage
	currentStateMessage
	HelloMessage
	rankMessage
)

type MessageTarget int

const (
	TargetMaster MessageTarget = iota
	TargetClient
	TargetBackup
	TargetElevator
)

type Message struct {
	Type    MessageType
	Target  MessageTarget
	Payload interface{}
}

// Keeping track of connections
type MasterConnectionInfo struct {
	ClientIP string
	Rank     int
	HostConn net.Conn
}

type ClientConnectionInfo struct {
	ID          string
	HostIP      string
	Rank        int
	ClientConn  net.Conn
	SendChan    chan Message
	ReceiveChan chan Message
}

type NetworkChannels struct {
	sendChan 	chan Message
	receiveChan chan Message
	MasterChan   chan Message
	BackupChan   chan Message
	ElevatorChan chan Message
}

type ActiveConnections struct {
	mutex sync.Mutex
	Conns []MasterConnectionInfo
}

type MasterToClientData struct {
	GlobalHallRequests [][2]bool `json:"globalHallRequests"`
	AssignedRequests   map[string][][2]bool
}

type ElevatorRequest struct {
	GlobalHallRequests [][2]bool `json:"globalHallRequests"`
	AssignedRequests   [][2]bool `json:"assignedRequests"`
}
