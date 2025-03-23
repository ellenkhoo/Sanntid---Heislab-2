package network

import (
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	"fmt"
	"time"
)

type AcknowledgeTracker struct {
	PendingAcks  map[string]bool
	Timeout      time.Duration
	RetryChannel chan sharedConsts.Message
}

func NewAcknowledgeTracker(timeout time.Duration) *AcknowledgeTracker {
	return &AcknowledgeTracker{
		PendingAcks:  make(map[string]bool),
		Timeout:      timeout,
		RetryChannel: make(chan sharedConsts.Message),
	}
}

func (ackTracker *AcknowledgeTracker) AwaitAcknowledge(clientID string, worldviewMsg sharedConsts.Message) {
	ackTracker.PendingAcks[clientID] = false
	go func() {
		time.Sleep(ackTracker.Timeout)
		if !ackTracker.PendingAcks[clientID] {
			fmt.Println("Acknowledgement not received from:", clientID)
			ackTracker.RetryChannel <- worldviewMsg
		}
	}()
}

func (ackTracker *AcknowledgeTracker) Acknowledge(clientID string) {
	ackTracker.PendingAcks[clientID] = true
	fmt.Println("Acknowledgement received from:", clientID)
}

func (ackTracker *AcknowledgeTracker) AllAcknowledged() bool {
	for _, acknowledged := range ackTracker.PendingAcks {
		if !acknowledged { // If any client has not acknowledged
			return false
		}
	}
	return true
}
