package network

import ( 
	"fmt"
	"net"
	"encoding/json"
	"github.com/ellenkhoo/ElevatorProject/roles"
)
// When a new connection is established on the client side, this function adds it to the list of active connections
func (ac *ActiveConnections) AddClientConnection(id string, conn net.Conn, sendChan chan Message, receiveChan chan Message) {
	//defer conn.Close()
	remoteIP, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	fmt.Println("Adding client connection")

	// ac.mu.Lock()
	// defer ac.mu.Unlock()

	//Check if IP is already added
	// for _, connections := range ac.conns {
	// 	if connections.IP == remoteIP {
	// 		return
	// 	}
	// }

	newConn := ClientConnectionInfo{
		ID: 	id,
		HostIP:          remoteIP,
		// need to do something about ranking
		ClientConn:  conn,
		SendChan:    sendChan,
		ReceiveChan: receiveChan,
	}

	// ac.conns = append(ac.conns, newConn)

	fmt.Println("Going to handle connection")
	go HandleConnection(newConn)
}

// Maybe not the most describing name
func HandleConnection(conn ClientConnectionInfo) {
	// Read from TCP connection and send to the receive channel
	fmt.Println("Reacing from TCP")
	go func() {
		decoder := json.NewDecoder(conn.ClientConn)
		for {
			var msg Message
			err := decoder.Decode(&msg)
			if err != nil {
				fmt.Println("Error decoding message: ", err)
				return
			}
			conn.ReceiveChan <- msg
		}
	}()

	// Read from the send channel and write to the TCP connection
	fmt.Println("Sending to TCP")
	go func() {
		encoder := json.NewEncoder(conn.ClientConn)
		for msg := range conn.SendChan {
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
	// hele meldingen skal sendes til backup
	// heis skal også motta globalhallrequests, men bare sin del av assigned requests

	clientID := clientConn.ID

	switch msg.Type {
	case rankMessage:
		data := msg.Payload
		if rank, ok := data.(int); ok {
			clientConn.Rank = rank
		}
	case masterRequestMessage:
		var localAssignedRequest [][2]bool
		data := msg.Payload
		if masterData, ok := data.(MasterToClientData); ok {
			// Extract the assigned requests for this specific elevator
			masterData.AssignedRequests[clientID] = localAssignedRequest

			backupMsg := masterData
			elevatorMsg := createElevatorMessage(masterData, clientID)

			// må lage en mld av type Message!
			//SendMessageOnChannel(, backupMsg)

		}
		
	}
}

// This function returns only the assigned requests relevant to a particular elevator + globalHallRequests
func createElevatorMessage(masterData MasterToClientData, elevatorID string) ElevatorRequest{
 

    localAssignedRequests := masterData.AssignedRequests[elevatorID]
	globalHallRequests := masterData.GlobalHallRequests

	elevatorMsg := ElevatorRequest{
		GlobalHallRequests: globalHallRequests,
		AssignedRequests: localAssignedRequests,
	}

	return elevatorMsg
}

