package network

import (
	"encoding/json"
	"fmt"
	"net"
)

// Adds the host's connection with the relevant client in the list of active connections
func (ac *ActiveConnections) AddHostConnection(rank int, conn net.Conn, sendChan chan Message) {

	ac.mutex.Lock()
	defer ac.mutex.Unlock()

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

	ac.conns = append(ac.conns, newConn)

	msg := Message{
		Type: helloMessage,
		Target: TargetBackup,
		Payload: "Hello from master",
	}

	fmt.Println("Sending hello message on channel")
	sendMessageOnChannel(sendChan, msg)
}


// Master listenes and accepts connections
func (ac *ActiveConnections)ListenAndAcceptConnections(port string, sendChan chan Message) {

	ac.mutex.Lock()
	defer ac.mutex.Unlock()

	ln, _ := net.Listen("tcp", ":"+port)

	for {
		hostConn, _ := ln.Accept()
		// send rank
		rank := len(ac.conns) + 2
		msg := Message{
			Type: rankMessage,
			Target: TargetClient,
			Payload: rank,
		}
		sendMessageOnChannel(sendChan, msg)
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
			// need to find the conn object connected to backup
			ac.mutex.Lock()
			fmt.Println("Backup is target")
			for i := range ac.conns {
				if ac.conns[i].Rank == 2 {
					targetConn = ac.conns[i].HostConn
					fmt.Println("Found backup conn")
					break
				}
			}
			ac.mutex.Unlock()
		case TargetElevator:
			// do something
		case TargetClient: 
			// do something
		}
		
		if targetConn != nil {
			encoder := json.NewEncoder(targetConn)
			for msg := range sendChan {
				fmt.Println("Sending message:", msg)
				err := encoder.Encode(msg)
				if err != nil {
					fmt.Println("Error encoding message: ", err)
					return
				}
			}
		} else {
			// If targetConn is nil, log a message or handle the case
			fmt.Println("No valid connection found for the message")
		}
	}
}
