package network

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/ellenkhoo/ElevatorProject/elevator"
	elevio "github.com/ellenkhoo/ElevatorProject/elevator/Driver"
	"github.com/ellenkhoo/ElevatorProject/hra"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

func AnnounceMaster(localIP string, port string) {
	fmt.Println("Announcing master")
	broadcastAddr := "255.255.255.255" + ":" + port
	conn, err := net.Dial("udp", broadcastAddr)
	if err != nil {
		fmt.Println("Error starting UDP listener:", err)
		return
	}
	defer conn.Close()

	for {
		msg := "I am Master"
		conn.Write([]byte(msg))
		time.Sleep(1 * time.Second) //announces every second, maybe it should happen more frequently?
	}
}

// Adds the host's connection with the relevant client in the list of active connections
func (ac *ActiveConnections) AddHostConnection(conn net.Conn, sendChan chan sharedConsts.Message) {

	remoteIP, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	newConn := MasterConnectionInfo{
		ClientIP: remoteIP,
		// Rank:     rank,
		HostConn: conn,
	}

	fmt.Printf("NewConn. ClientIP: %s", newConn.ClientIP)

	ac.mutex.Lock()
	ac.Conns = append(ac.Conns, newConn)
	ac.mutex.Unlock()

	// // Marshal rank
	// rankJSON, err := json.Marshal(rank)
	// if err != nil {
	// 	fmt.Println("Error marshalling rank: ", err)
	// 	return
	// }

	// msg := sharedConsts.Message{
	// 	Type:    sharedConsts.RankMessage,
	// 	Target:  sharedConsts.TargetClient,
	// 	Payload: rankJSON,
	// }

	// fmt.Println("Sending rank message on channel")
	// sendChan <- msg
}

// Master listenes and accepts connections
func (ac *ActiveConnections) ListenAndAcceptConnections(port string, networkChannels *sharedConsts.NetworkChannels) {

	ln, _ := net.Listen("tcp", ":"+port)

	for {
		hostConn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error acepting connection:", err)
			continue
		}
		// rank := len(ac.Conns) + 2

		go ReceiveMessage(networkChannels.ReceiveChan, hostConn)
		go ac.AddHostConnection(hostConn, networkChannels.SendChan)
	}
}

func (ac *ActiveConnections) MasterSendMessages(networkChannels *sharedConsts.NetworkChannels) {

	fmt.Println("Arrived at masterSend")
	var targetConn net.Conn
	for msg := range networkChannels.SendChan {
		switch msg.Target {
		// Må sende worldview til backup først, så til heis
		case sharedConsts.TargetBackup:
			fmt.Println("Backup is target")
			for clients := range ac.Conns {
				targetConn = ac.Conns[clients].HostConn
				SendMessage(msg, targetConn)
			}

			// if targetConn != nil {
			// 	encoder := json.NewEncoder(targetConn)
			// 	fmt.Println("Sending message:", msg)
			// 	err := encoder.Encode(msg)
			// 	if err != nil {
			// 		fmt.Println("Error encoding message: ", err)
			// 		return
			// } else {
			// 	// If targetConn is nil, log a message or handle the case
			// 	fmt.Println("No valid connection found for the message")
			// }

		// case sharedConsts.TargetMaster:
		// 	networkChannels.MasterChan <- msg

		// case sharedConsts.TargetElevator:
		// 	// do something
		case sharedConsts.TargetClient:
			// Send to remote clients
			for clients := range ac.Conns {
				targetConn = ac.Conns[clients].HostConn
				SendMessage(msg, targetConn)
			}
		}
	}
}

func (masterData *MasterData) HandleReceivedMessagesToMaster(ac *ActiveConnections, msg sharedConsts.Message, networkChannels *sharedConsts.NetworkChannels, ackTracker *AcknowledgeTracker) {

	fmt.Println("At handleMessagesToMaster")
	switch msg.Type {
	case sharedConsts.LocalRequestMessage:
		// Update GlobalHallRequests
		var request elevio.ButtonEvent
		err := json.Unmarshal(msg.Payload, &request)
		if err != nil {
			fmt.Println("Error unmarshalling payload: ", err)
			return
		}

		fmt.Println("Received request: ", request)
		floor := request.Floor
		button := request.Button
		masterData.mutex.Lock()
		fmt.Println("Updating globalHallRequests")
		masterData.GlobalHallRequests[floor][button] = true
		masterData.mutex.Unlock()

	case sharedConsts.CurrentStateMessage:
		// Update allElevStates
		var elevStates elevator.ElevStates
		err := json.Unmarshal(msg.Payload, &elevStates)
		if err != nil {
			fmt.Println("Error unmarshalling payload: ", err)
			return
		}

		fmt.Printf("Received current state from elevator: %#v\n", elevStates)
		ID := elevStates.IP
		masterData.mutex.Lock()
		fmt.Println("Updating allElevStates")
		masterData.AllElevStates[ID] = elevStates
		floor := elevStates.Floor
		dirn := elevStates.Direction
		if dirn == "D_Up" {
			masterData.GlobalHallRequests[floor][0] = false
		} else if dirn == "D_Down" {
			masterData.GlobalHallRequests[floor][1] = false
		}
		masterData.mutex.Unlock()
		assignedOrders := hra.SendStateToHRA(masterData.AllElevStates, masterData.GlobalHallRequests)
		masterData.mutex.Lock()
		for ID, orders := range *assignedOrders {
			masterData.AllAssignedRequests[ID] = orders
			fmt.Println("Assigned orders for ID: ", ID, " are: ", orders)
		}
		masterData.mutex.Unlock()

		clientData := BackupData{
			GlobalHallRequests:  masterData.GlobalHallRequests,
			AllAssignedRequests: masterData.AllAssignedRequests,
		}

		// Marshal clientData
		clientDataJSON, err := json.Marshal(clientData)
		if err != nil {
			fmt.Println("Error marshalling clientData: ", err)
			return
		}

		// Create message
		orderMsg := sharedConsts.Message{
			Type:    sharedConsts.MasterWorldviewMessage,
			Target:  sharedConsts.TargetClient,
			Payload: clientDataJSON,
		}
		// Send message
		networkChannels.SendChan <- orderMsg
		ackTracker.AwaitAcknowledge(ID, orderMsg)

	case sharedConsts.BackupAcknowledgeMessage:
		fmt.Println("At backup ack")
		var clientID string
		err := json.Unmarshal(msg.Payload, &clientID)
		if err != nil {
			fmt.Println("Error decoding BackupAcknoledgement:", err)
			return
		}
		ackTracker.Acknowledge(clientID)

		if ackTracker.AllAcknowledged() {
			fmt.Println("All acknowledgments received. Orders can be sent to elevators.")

			// Data to remote clients
			clientData := "Send requests to elevator"
			clientDataJSON, err := json.Marshal(clientData)
			if err != nil {
				fmt.Println("Error marshalling backup data: ", err)
				return
			}

			clientMsg := sharedConsts.Message{
				Type:    sharedConsts.UpdateOrdersMessage,
				Target:  sharedConsts.TargetClient,
				Payload: clientDataJSON,
			}

			var targetConn net.Conn
			for clients := range ac.Conns {
				targetConn = ac.Conns[clients].HostConn
				SendMessage(clientMsg, targetConn)
			}

			// Local elevator also needs update

		}
		// case sharedConsts.MasterWorldviewMessage:
		// 	var backupWorldview BackupData
		// 	err := json.Unmarshal(msg.Payload, &backupWorldview)
		// 	if err != nil {
		// 		fmt.Println("Error decoding message to master: ", err)
		// 		return
		// 	}
		// 	backupHallRequests := backupWorldview.GlobalHallRequests
		// 	backupAssignedRequests := backupWorldview.AllAssignedRequests
		// 	if masterData.GlobalHallRequests == backupHallRequests && mapsAreEqual(masterData.AllAssignedRequests, backupAssignedRequests) {
		// 		fmt.Println("Worldviews are the same")

		// 		data := "Send requests to elevator"
		// 		dataJSON, err := json.Marshal(data)
		// 		if err != nil {
		// 			fmt.Println("Error marshalling backup data: ", err)
		// 			return
		// 		}

		// 		backupMsg := sharedConsts.Message{
		// 			Type:    sharedConsts.UpdateOrdersMessage,
		// 			Target:  sharedConsts.TargetMaster,
		// 			Payload: backupDataJSON,
		// 		}
		// 		// backup can send worldview to elevator
		// 	} else {
		// 		fmt.Println("Worldviews are different")
		// 		// send master's worldview again
		// 	}
	}
}

// flytt til et annet sted?
func mapsAreEqual(map1, map2 map[string][4][2]bool) bool {
	// First, check if the lengths are the same
	if len(map1) != len(map2) {
		return false
	}

	// Now, check if all the keys and their corresponding values are equal
	for key, val1 := range map1 {
		val2, exists := map2[key]
		if !exists || val1 != val2 {
			return false
		}
	}

	return true
}
