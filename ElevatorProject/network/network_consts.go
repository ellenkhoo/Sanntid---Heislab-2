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
	globalHallRequestMessage MessageType = iota
	assignedHallRequestsMessage
	backupAcknowledgeMessage
	localRequestMessage
	currentStateMessage
	helloMessage
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
	ClientIP    string
	Rank        int
	HostConn    net.Conn
}

type ClientConnectionInfo struct {
	HostIP 		string
	Rank 		int
	ClientConn  net.Conn
	SendChan    chan Message
	ReceiveChan chan Message
}

type NetworkChannels struct {
	MasterChan   chan Message
	BackupChan   chan Message
	ElevatorChan chan Message
}

type ActiveConnections struct {
	mutex    sync.Mutex
	conns []MasterConnectionInfo
}