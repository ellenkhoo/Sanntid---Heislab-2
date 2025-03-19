package network

import (
	"encoding/json"
	"fmt"
	"net"

	//"github.com/ellenkhoo/ElevatorProject/heartbeat"
	//"github.com/ellenkhoo/ElevatorProject/roles"
)

// When a new connection is established on the client side, this function adds it to the list of active connections
func (client *ClientConnectionInfo) AddClientConnection(id string, clientConn net.Conn, sendChan chan Message, receiveChan chan Message) {
	//defer conn.Close()
	remoteIP, _, _ := net.SplitHostPort(clientConn.RemoteAddr().String())

	fmt.Println("Adding client connection")

	*client = ClientConnectionInfo{
		ID: 	id,
		HostIP:          remoteIP,
		ClientConn:  clientConn,
		SendChan:    sendChan,
		ReceiveChan: receiveChan,
	}

	fmt.Println("Going to handle connection")
	go HandleConnection(*client)
}

// Maybe not the most describing name
func HandleConnection(client ClientConnectionInfo) {
	// Read from TCP connection and send to the receive channel
	fmt.Println("Reacing from TCP")
	go func() {
		decoder := json.NewDecoder(client.ClientConn)
		for {
			var msg Message
			err := decoder.Decode(&msg)
			if err != nil {
				fmt.Println("Error decoding message: ", err)
				return
			}
			client.ReceiveChan <- msg
		}
	}()

	// Read from the send channel and write to the TCP connection
	fmt.Println("Sending to TCP")
	go func() {
		encoder := json.NewEncoder(client.ClientConn)
		for msg := range client.SendChan {
			err := encoder.Encode(msg)
			if err != nil {
				fmt.Println("Error encoding message: ", err)
				return
			}
		}
	}()
}

func ClientSendMessages(sendChan chan Message, conn net.Conn) {

	fmt.Println("Ready to send msg to master")

	encoder := json.NewEncoder(conn)
	for msg := range sendChan {
		fmt.Println("Sending message:", msg)
		err := encoder.Encode(msg)
		if err != nil {
			fmt.Println("Error encoding message: ", err)
			return
		}
	}
}

// Messages sent to a client means that the data is meant both for an elevator thread and the potential backup
func (clientConn *ClientConnectionInfo)HandleReceivedMessageToClient(msg Message) {

	clientID := clientConn.ID

	switch msg.Type {
	case rankMessage:
		data := msg.Payload
		if rank, ok := data.(int); ok {
			fmt.Println("Setting my rank to", rank)
			clientConn.Rank = rank
			if rank == 2 {
				fmt.Println("My rank is 2 and I will become backup")
				// start backup
			}
		}
	case masterOrdersMessage:
		data := msg.Payload
		if masterData, ok := data.(MasterData); ok {
			backupData:= CreateBackupData(masterData)
			elevatorData := CreateElevatorData(masterData, clientID)

			backupMsg := Message{
				Type: masterOrdersMessage,
				Target: TargetBackup,
				Payload: backupData,
			}


			elevatorMsg := Message{
				Type: masterOrdersMessage,
				Target: TargetElevator,
				Payload: elevatorData,
			}

			fmt.Println("Sending messages to backup and elevator")
			clientConn.ReceiveChan <- backupMsg
			clientConn.ReceiveChan <- elevatorMsg
		}
	// case heartbeat: //
	// 	// start timer
	// case timeout:
	// 	// start master
	}
}

// This function returns only the assigned requests relevant to a particular elevator + globalHallRequests
func CreateElevatorData(masterData MasterData, elevatorID string) ElevatorRequest{
 

    localAssignedRequests := masterData.AllAssignedRequests[elevatorID]
	globalHallRequests := masterData.GlobalHallRequests

	elevatorData := ElevatorRequest{
		GlobalHallRequests: globalHallRequests,
		AssignedRequests: localAssignedRequests,
	}

	return elevatorData
}

func CreateBackupData(masterData MasterData) BackupData {
 

    AllAssignedRequests := masterData.AllAssignedRequests
	globalHallRequests := masterData.GlobalHallRequests

	backupData := BackupData{
		GlobalHallRequests: globalHallRequests,
		AllAssignedRequests: AllAssignedRequests,
	}

	return backupData
}

