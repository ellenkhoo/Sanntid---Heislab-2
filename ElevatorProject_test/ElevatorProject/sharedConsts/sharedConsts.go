package sharedConsts

import "encoding/json"

// This module defines shared constants, message types, and communication channels
// used by both the network and elevator modules. It ensures consistency in inter-module
// communication by standardizing message structures and types.

type HelloMsg struct {
	Message string
	Iter    int
}

type MessageType int

const (
	MasterWorldviewMessage MessageType = iota
	UpdateOrdersMessage
	LocalRequestMessage
	CurrentStateMessage
	ActiveConnectionsMessage
	ClientIDMessage
	PriorCabRequestsMessage
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
	Payload json.RawMessage
}

type NetworkChannels struct {
	SendChan     chan Message
	ReceiveChan  chan Message
	MasterChan   chan Message
	BackupChan   chan Message
	ElevatorChan chan Message
	RestartChan  chan string
	UpdateChan   chan string
}
