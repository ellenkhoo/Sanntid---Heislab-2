package network

import (
	"fmt"
	"time"
	"encoding/json"

	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	// "github.com/ellenkhoo/ElevatorProject/heartbeat"
	// "github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

// func SendHeartbeatToMaster(ac *ActiveConnections, sendChan chan Message, receiveChan chan Message) bool {
// 	sendChan <- Message{
// 		Type:   HelloMessage,
// 		Target: TargetMaster,
// 		Payload: HelloMsg{
// 			Message: "heartbeat",
// 			Iter:    0,
// 		},
// 	}

// 	go ClientSendMessages(sendChan, ac.Conns[0].HostConn)

// 	timeout := time.After(5 * time.Second)
// 	select {
// 	case reply := <-receiveChan:
// 		if reply.Type == HelloMessage {
// 			return true
// 		}
// 	case <-timeout:
// 		fmt.Println("timeout: no response from master")
// 		return false
// 	}
// 	return false
// }

// func findMasterConnection(ac *ActiveConnections) net.Conn {
// 	// ac.mutex.Lock()
// 	// defer ac.mutex.Unlock()

// 	for _, conn := range ac.Conns {
// 		if conn.Rank == 1 {
// 			return conn.HostConn
// 		}
// 	}
// 	return nil
// }

// func findBackupConnection(ac *ActiveConnections) net.Conn {
// 	// ac.mutex.Lock()
// 	// defer ac.mutex.Unlock()

// 	for _, conn := range ac.Conns {
// 		if conn.Rank == 2 {
// 			return conn.HostConn
// 		}
// 	}
// 	return nil
// }

// func SendHeartbeatToClient(ac *ActiveConnections, sendChan chan Message, receiveChan chan Message) bool {
// 	sendChan <- Message{
// 		Type:   HelloMessage,
// 		Target: TargetBackup,
// 		Payload: HelloMsg{
// 			Message: "heartbeat",
// 			Iter:    0,
// 		},
// 	}

// 	go ac.MasterSendMessages(sendChan)

// 	timeout := time.After(5 * time.Second)
// 	select {
// 	case reply := <-receiveChan:
// 		if reply.Type == HelloMessage {
// 			return true
// 		}
// 	case <-timeout:
// 		fmt.Println("timeout: no response from backup")
// 		return false
// 	}
// 	return false
// }

// func StartHeartbeat(ac *ActiveConnections, sendChan chan Message, receiveChan chan Message, bcastPortInt int, bcastPortString string, peersPort int, TCPPort string, networkChannels NetworkChannels) {
// 	ticker := time.NewTicker(5 * time.Second)
// 	for range ticker.C {
// 		masterConn := findMasterConnection(ac)
// 		if masterConn != nil {
// 			if !SendHeartbeatToMaster(ac, sendChan, receiveChan) {
// 				fmt.Println("master disconnected, starting failover...")
// 				bcastPortInt_dis := bcastPortInt
// 				bcastPortString_dis := bcastPortString
// 				peersPort_dis := peersPort
// 				TCPPort_dis := TCPPort
// 				handleMasterDisconnection(ac, sendChan, receiveChan, bcastPortInt_dis, bcastPortString_dis, peersPort_dis, TCPPort_dis, networkChannels)
// 			}
// 		}

// 		backupConn := findBackupConnection(ac)
// 		if backupConn != nil {
// 			if !SendHeartbeatToClient(ac, sendChan, receiveChan) {
// 				fmt.Println("Backup disconnected, starting failover...")

// 				for _, conn := range ac.Conns {
// 					if conn.HostConn == backupConn {
// 						ac.HandleClientDisconnection(conn.ClientIP)
// 						break
// 					}
// 				}
// 			}
// 		}
// 	}
// }

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
	
	ticker := time.NewTimer(5*time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Println("sending heartbeat to clients")
		sendChan <- msg
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
	for {
		time.Sleep(5 * time.Second)
		fmt.Println("sending heartbeat from clients", clientConn.ID)
		sendChan <- msg
	}
}
// func (client *ClientConnectionInfo) ListenForHeartbeats(networkChannels sharedConsts.NetworkChannels) {
// 	buffer := make([]byte, 1024)
// 	timeout := time.NewTimer(5 * time.Second)
// 	readChan := make(chan error, 1)


// 	go func () {
// 		for {
// 			client.ClientConn.SetReadDeadline(time.Now().Add(5 * time.Second))
// 			_, err := io.ReadFull(client.ClientConn, buffer)
// 			select {
// 			case readChan <- err:
// 			default:
// 			}
// 		}
// 	}()

// 	go func ()  {
// 		for{
// 			select {
// 			case err := <- readChan:
// 				if err != nil{
// 					fmt.Println("Lost connection with matser. Waiting for timeout to confirm")
// 				} else {
// 					fmt.Print("Recevied heartbeat from master")
// 					timeout.Reset(5 * time.Second)
// 				}
// 			case <- timeout.C:
// 				fmt.Println("Master has not sent heartbeat in 5 seconds, starting failover...")
// 				if client.Rank == 2 {
// 					fmt.Println("backup failover")
// 				} else if client.Rank > 2 {
// 					fmt.Println("client failover")
// 				}
// 				return
// 			}
// 		}
// 	}()
	
// }
