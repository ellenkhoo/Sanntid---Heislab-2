package network

import (
	"fmt"
	"net"
	"time"
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

func (ac *ActiveConnections) SendHeartbeats() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Println("Master: Sending heartbeat")

		ac.mutex.Lock()
		for _, conn := range ac.Conns {
			_, err := conn.HostConn.Write([]byte("HB"))
			if err != nil {
				fmt.Println("Error sending heartbeat to", conn.ClientIP, ":", err)
			}
		}
		ac.mutex.Unlock()
	}
}

func ListenForHeartbeats(conn net.Conn, role string, onMasterFail func()) {
	buffer := make([]byte, 2)
	timeout := time.NewTimer(5 * time.Second)

	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, err := conn.Read(buffer)

		if err != nil || string(buffer) != "HB" {
			fmt.Printf("%s: lost connection with master! starter failover...", role)
			onMasterFail()
			return
		}

		fmt.Printf("%s: Received heartbeat from master\n", role)
		timeout.Reset(5 * time.Second)
	}
}
