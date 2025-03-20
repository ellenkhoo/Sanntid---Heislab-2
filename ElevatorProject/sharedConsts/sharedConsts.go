package sharedConsts

import "encoding/json"

// NETWORK CONSTS
type HelloMsg struct {
	Message string
	Iter    int
}

type MessageType int

const (
	MasterOrdersMessage MessageType = iota
	BackupAcknowledgeMessage
	LocalRequestMessage
	CurrentStateMessage
	HelloMessage
	RankMessage
	ElevClearedOrderMessage
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
	SendChan 	chan Message
	ReceiveChan chan Message
	MasterChan   chan Message
	BackupChan   chan Message
	ElevatorChan chan Message
}
