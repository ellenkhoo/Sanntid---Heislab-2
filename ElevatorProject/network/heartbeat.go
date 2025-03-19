package network

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

// sjekker om master lever
func SendHeartbeatToMaster(ac *ActiveConnections, sendChan chan sharedConsts.Message, receiveChan chan sharedConsts.Message, ctx context.Context) bool {
	sendChan <- sharedConsts.Message{
		Type:   sharedConsts.HelloMessage,
		Target: sharedConsts.TargetMaster,
		Payload: sharedConsts.HelloMsg{
			Message: "heartbeat",
			Iter:    0,
		},
	}

	go ClientSendMessages(sendChan, ac.Conns[0].HostConn, ctx)

	timeout := time.After(5 * time.Second)
	select {
	case reply := <-receiveChan:
		if reply.Type == sharedConsts.HelloMessage {
			return true
		}
	case <-timeout:
		fmt.Println("timeout: no response from master")
		return false
	}
	return false
}

func findMasterConnection(ac *ActiveConnections) net.Conn {
	// ac.mutex.Lock()
	// defer ac.mutex.Unlock()

	for _, conn := range ac.Conns {
		if conn.Rank == 1 {
			return conn.HostConn
		}
	}
	return nil
}

func findBackupConnection(ac *ActiveConnections) net.Conn {
	// ac.mutex.Lock()
	// defer ac.mutex.Unlock()

	for _, conn := range ac.Conns {
		if conn.Rank == 2 {
			return conn.HostConn
		}
	}
	return nil
}

func SendHeartbeatToClient(ac *ActiveConnections, sendChan chan sharedConsts.Message, receiveChan chan sharedConsts.Message, ctx context.Context) bool {
	sendChan <- sharedConsts.Message{
		Type:   sharedConsts.HelloMessage,
		Target: sharedConsts.TargetClient,
		Payload: sharedConsts.HelloMsg{
			Message: "heartbeat",
			Iter:    0,
		},
	}

	go ac.MasterSendMessages(sendChan, ctx)

	timeout := time.After(5 * time.Second)
	select {
	case reply := <-receiveChan:
		if reply.Type == sharedConsts.HelloMessage {
			return true
		}
	case <-timeout:
		fmt.Println("timeout: no response from backup")
		return false
	}
	return false
}

func StartHeartbeat(ac *ActiveConnections, sendChan chan sharedConsts.Message, receiveChan chan sharedConsts.Message, bcastPortInt int, bcastPortString string, peersPort int, TCPPort string, 
	networkChannels sharedConsts.NetworkChannels, ctx context.Context, cancel context.CancelFunc) {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		masterConn := findMasterConnection(ac)
		if masterConn != nil {
			if !SendHeartbeatToMaster(ac, sendChan, receiveChan, ctx) {
				fmt.Println("master disconnected, starting failover...")
				//bcastPortInt_dis := bcastPortInt
				bcastPortString_dis := bcastPortString
				peersPort_dis := peersPort
				TCPPort_dis := TCPPort
				handleMasterDisconnection(ac, sendChan, receiveChan, bcastPortString_dis, peersPort_dis, TCPPort_dis, networkChannels, ctx, cancel)
			}
		}

		backupConn := findBackupConnection(ac)
		if backupConn != nil {
			if !SendHeartbeatToClient(ac, sendChan, receiveChan, ctx) {
				fmt.Println("Backup disconnected, starting failover...")

				for _, conn := range ac.Conns {
					if conn.HostConn == backupConn {
						ac.HandleClientDisconnection(conn.ClientIP)
						break
					}
				}
			}
		}
	}
}