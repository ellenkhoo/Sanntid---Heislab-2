package network

import (
	"fmt"
	"time"

	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

type AcknowledgeTracker struct {
	PendingAcks  map[string]bool
	Timeout      time.Duration
	RetryChannel chan sharedConsts.Message
}

func NewAcknowledgeTracker(SendChan chan sharedConsts.Message, timeout time.Duration) *AcknowledgeTracker {
	return &AcknowledgeTracker{
		PendingAcks:  make(map[string]bool),
		Timeout:      timeout,
		RetryChannel: SendChan,
	}
}

func (ackTracker *AcknowledgeTracker) AwaitAcknowledge(clientID string, worldviewMsg sharedConsts.Message) {
	ackTracker.PendingAcks[clientID] = false

	retryLimit := 3
	retryCount := 0

	go func() {
		for retryCount < retryLimit {
			time.Sleep(ackTracker.Timeout)
			if !ackTracker.PendingAcks[clientID] {
				fmt.Println("Acknowledgement not received from:", clientID)
				ackTracker.RetryChannel <- worldviewMsg
				retryCount++
				fmt.Println("retryCount:", retryCount)
			} else {
				break
			}
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
