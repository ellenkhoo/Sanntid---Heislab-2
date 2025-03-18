package network

import (
	"encoding/json"
	"fmt"
	"net"
)

// Adds the host's connection with the relevant client in the list of active connections
func (ac *ActiveConnections) AddHostConnection(rank int, conn net.Conn, sendChan chan Message) {

	remoteIP, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	// Check if a connection already exists with this IP
	// for i, connection := range ac.conns {
	// 	if connection.ClientIP == remoteIP {
	// 		ac.conns[i].HostConn = conn
	// 		return
	// 	}
	// }

	// sett hostIp + clientIP
	newConn := MasterConnectionInfo{
		ClientIP: remoteIP,
		Rank: rank,
		HostConn: conn,
	}

	fmt.Printf("NewConn. ClientIP: %s, Rank: %d", newConn.ClientIP, newConn.Rank)

	ac.mutex.Lock()
	ac.conns = append(ac.conns, newConn)
	ac.mutex.Unlock()

	msg := Message{
		Type: rankMessage,
		Target: TargetBackup,
		Payload: rank,
	}

	fmt.Println("Sending rank message on channel")
	sendMessageOnChannel(sendChan, msg)
}


// Master listenes and accepts connections
func (ac *ActiveConnections)ListenAndAcceptConnections(port string, sendChan chan Message, receiveChan chan Message) {

	ln, _ := net.Listen("tcp", ":"+port)

	for {
		hostConn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error acepting connection:", err)
			continue
		}
		rank := len(ac.conns) + 2

		go ReceiveMessage(receiveChan, hostConn)
		go ac.AddHostConnection(rank, hostConn, sendChan)
	}
}

func (ac *ActiveConnections)MasterSendMessages(sendChan chan Message) {

	fmt.Println("Arrived at masterSend")
	
	var targetConn net.Conn
	for msg := range sendChan {
		fmt.Println("target: ", msg.Target)
		switch msg.Target {
		case TargetBackup: 
			// Need to find the conn object connected to backup
			fmt.Println("Backup is target")
			for i := range ac.conns {
				if ac.conns[i].Rank == 2 {
					targetConn = ac.conns[i].HostConn
					fmt.Println("Found backup conn")
					break
				}
			}

		case TargetElevator:
			// do something
		case TargetClient: 

			// do something
		}
		
		if targetConn != nil {
			encoder := json.NewEncoder(targetConn)
			fmt.Println("Sending message:", msg)
			err := encoder.Encode(msg)
			if err != nil {
				fmt.Println("Error encoding message: ", err)
				return
			}
		} else {
			// If targetConn is nil, log a message or handle the case
			fmt.Println("No valid connection found for the message")
		}
	}
}
