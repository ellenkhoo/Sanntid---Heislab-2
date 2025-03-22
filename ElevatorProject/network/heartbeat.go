package network

import (
	"fmt"
	"time"
	"encoding/json"

	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	// "github.com/ellenkhoo/ElevatorProject/heartbeat"
	// "github.com/ellenkhoo/ElevatorProject/sharedConsts"
)


func (ac *ActiveConnections) MasterSendHeartbeats(sendChan chan sharedConsts.Message) {
	heartbeatPayload, err := json.Marshal("HB")
	if err != nil {
		fmt.Println("Error marshalling heartbeat: ", err)
		return
	}

	msg := sharedConsts.Message{
		Type:    sharedConsts.Heartbeat,
		Target:  sharedConsts.TargetClient,
		Payload: heartbeatPayload,
	}
	
	ticker := time.NewTicker(2*time.Second)
	defer ticker.Stop()

	for {
		<- ticker.C
		fmt.Println("sending heartbeat to clients")
		sendChan <- msg
	}
}


func (client *ClientConnectionInfo) ClientHandleHeartbeatTimeout() {
	for {
		select {
		case <- client.HeartbeatTimer.C:
			fmt.Println("no heartbeat received from master.Master may be down, staring failover!")
			return
		}
	}
}


func (clientConn *ClientConnectionInfo) ClientSendHeartbeats(sendChan chan sharedConsts.Message) {
	heartbeatPayload, err := json.Marshal(clientConn.ID)
	if err != nil {
		fmt.Println("Error marshalling heartbeat: ", err)
		return
	}

	msg := sharedConsts.Message{
		Type:    sharedConsts.Heartbeat,
		Target:  sharedConsts.TargetMaster,
		Payload: heartbeatPayload,
	}
	ticker := time.NewTicker(2*time.Second)
	defer ticker.Stop()

	for {
		<- ticker.C
		fmt.Println("sending heartbeat from client:", clientConn.ID)
		sendChan <- msg
	}
}
