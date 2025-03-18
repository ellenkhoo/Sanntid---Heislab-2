package network

import (
	"fmt"
	"net"
	"encoding/json"
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

	ac.conns = append(ac.conns, newConn)

	msg := Message{
		Type: helloMessage,
		Target: TargetBackup,
		Payload: "Hello from master",
	}

	fmt.Println("Sending hello message on channel")
	sendChan <- msg
}


// Master listenes and accepts connections
func ListenAndAcceptConnections(ac ActiveConnections, port string, sendChan chan Message) {
	ln, _ := net.Listen("tcp", ":"+port)

	for {
		hostConn, _ := ln.Accept()
		// send rank
		rank := len(ac.conns) + 1
		msg := Message{
			Type: rankMessage,
			Target: TargetClient,
			Payload: rank,
		}
		sendMessageOnChannel(sendChan, msg)
		go ac.AddHostConnection(rank, hostConn, sendChan)
	}
}

func MasterSendMessages(sendChan chan Message, ac ActiveConnections) {

	var targetConn net.Conn
	for msg := range sendChan {
		switch msg.Target {
		case TargetBackup: 
			// need to find the conn object connected to backup
			for i := range ac.conns {
				if ac.conns[i].Rank == 2 {
					targetConn = ac.conns[i].HostConn
				} else {
					targetConn = nil
				}
			}
			encoder := json.NewEncoder(targetConn)
			for msg := range sendChan {
				fmt.Println("Sending message:", msg)
				err := encoder.Encode(msg)
				if err != nil {
					fmt.Println("Error encoding message: ", err)
					return
				}
			}
		case TargetElevator:
			// do something
		case TargetClient: 
			// do something
		}
		
	}
}
