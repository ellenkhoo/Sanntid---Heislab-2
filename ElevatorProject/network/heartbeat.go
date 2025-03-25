package network

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	// "github.com/ellenkhoo/ElevatorProject/heartbeat"
	// "github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

func (ac *ActiveConnections) MasterSendHeartbeats(heartbeatChan chan sharedConsts.Message) {

	heartbeatPayload, err := json.Marshal("HB")
	if err != nil {
		fmt.Println("Error marshalling heartbeat: ", err)
		return
	}

	heartbeatMsg := sharedConsts.Message{
		Type:    sharedConsts.Heartbeat,
		Target:  sharedConsts.TargetClient,
		Payload: heartbeatPayload,
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	fmt.Println("Ticker started for sending heartbeats.")

	for {
		fmt.Println("waiting for ticker")
		<-ticker.C
		fmt.Println("Before sending heartbeat")
		heartbeatChan <- heartbeatMsg
	}
		
	// 	// Loop through each client connection and send heartbeat to their respective HeartbeatChan
	// 	for _, clientConn := range ac.Conns {
	// 		select {
	// 		case clientConn.Channels.HeartbeatChan <- heartbeatMsg:
	// 			fmt.Println("Sent heartbeat to client: ", clientConn.ClientIP)
	// 		default:
	// 			fmt.Println("Heartbeat channel for client is full or blocked, could not send.")
	// 		}
	// 	}
	// }
}

func (clientConn *ClientConnectionInfo) StartListeningForHeartbeats() {
    go func() {
        for msg := range clientConn.Channels.HeartbeatChan {
            fmt.Println("Received heartbeat message in go-routine")  // Logging for å bekrefte at du er i go-rutinen
            clientConn.HandleReceivedMessageToClient(msg)
        }
    }()
}

// func (ac *ActiveConnections) MasterHandleHeartbeatTimeout() {
// 	ac.mutex.Lock()
// 	for clientID, timer := range ac.ClientTimers {
// 		select  {
// 		case <- timer.C:
// 			fmt.Println("no heartbeat received from client:", clientID, "starting failover...")
// 		default:
// 		}
// 	}
// 	ac.mutex.Unlock()
// }

// func (masterData *MasterData) MasterStartheartbeatTimer(){

// 	masterData.HeartbeatTimer = time.NewTimer(5 * time.Second)
// }

/*______________________________________________________________________________________________________________-
______________________________________________________________________________________________________________-*/

func (clientConn *ClientConnectionInfo) ClientSendHeartbeats(heartbeatChan chan sharedConsts.Message) {
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
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		fmt.Println("sending heartbeat from client:", clientConn.ID)
		heartbeatChan <- msg
	}
}

func (client *ClientConnectionInfo) ClientHandleHeartbeatTimeout() {
	for {
		select {
		case <-client.HeartbeatTimer.C:
			fmt.Println("no heartbeat received from master.Master may be down, staring failover!")
			return
		}
	}
}
