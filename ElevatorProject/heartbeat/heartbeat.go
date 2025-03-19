package heartbeat

import (
	"fmt"
	"net"
	"time"

	"github.com/ellenkhoo/ElevatorProject/network"
)

func SendHeartbeatToMaster(ac *network.ActiveConnections, sendChan chan network.Message, receiveChan chan network.Message) bool {
	sendChan <- network.Message{
		Type:   network.HelloMessage, //hvorfor ikke network.helloMessage??
		Target: network.TargetMaster,
		Payload: network.HelloMsg{
			Message: "heartbeat",
			Iter:    0,
		},
	}

	go network.ClientSendMessages(sendChan, ac.Conns[0].HostConn)

	timeout := time.After(5 * time.Second)
	select {
	case reply := <-receiveChan:
		if reply.Type == network.HelloMessage {
			return true
		}
	case <-timeout:
		fmt.Println("timeout: no response from master")
		return false
	}
	return false
}

func findMasterConnection(ac *network.ActiveConnections) net.Conn {
	// ac.mutex.Lock()
	// defer ac.mutex.Unlock()

	for _, conn := range ac.Conns {
		if conn.Rank == 1 {
			return conn.HostConn
		}
	}
	return nil
}

func findBackupConnection(ac *network.ActiveConnections) net.Conn {
	// ac.mutex.Lock()
	// defer ac.mutex.Unlock()

	for _, conn := range ac.Conns {
		if conn.Rank == 2 {
			return conn.HostConn
		}
	}
	return nil
}

func SendHeartbeatToBackup(ac *network.ActiveConnections, sendChan chan network.Message, receiveChan chan network.Message) bool {
	sendChan <- network.Message{
		Type:   network.HelloMessage,
		Target: network.TargetBackup,
		Payload: network.HelloMsg{
			Message: "heartbeat",
			Iter:    0,
		},
	}

	go ac.MasterSendMessages(sendChan)

	timeout := time.After(5 * time.Second)
	select {
	case reply := <-receiveChan:
		if reply.Type == network.HelloMessage {
			return true
		}
	case <-timeout:
		fmt.Println("timeout: no response from backup")
		return false
	}
	return false
}

func StartHeartbeat(ac *network.ActiveConnections, sendChan chan network.Message, receiveChan chan network.Message) {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		masterConn := findMasterConnection(ac)
		if masterConn != nil {
			if !SendHeartbeatToMaster(ac, sendChan, receiveChan) {
				fmt.Println("master disconnected, starting failover...")
				//handleMasterDisconnection(ac)
			}
		}

		backupConn := findBackupConnection(ac)
		if backupConn != nil {
			if !SendHeartbeatToBackup(ac, sendChan, receiveChan) {
				fmt.Println("Backup disconnected, starting failover...")
				//handleBackupDisconnection(ac)
			}
		}
	}
}
