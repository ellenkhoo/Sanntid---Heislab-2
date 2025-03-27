package network

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/ellenkhoo/ElevatorProject/elevator"
	hra "github.com/ellenkhoo/ElevatorProject/hallRequestAssigner"
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
		time.Sleep(100 * time.Millisecond)
	}
}

func (ac *ActiveConnections) ListenAndAcceptConnections(masterData *MasterData, client *ClientConnectionInfo, port string, networkChannels *sharedConsts.NetworkChannels) {

	ln, _ := net.Listen("tcp", ":"+port)

	for {
		hostConn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error acepting connection:", err)
			continue
		}

		// TEST
		tcpConn, err := ConfigureTCPConn(hostConn)
		if err != nil {
			fmt.Println("Failed to configure TCP settings")
			hostConn.Close()
			continue
		}

		go ReceiveMessage(masterData, client, ac, client.Channels, tcpConn)
		//go ac.AddHostConnection(tcpConn, networkChannels.SendChan)
	}
}

// Adds the host's connection with a client to ActiveConnections
func (ac *ActiveConnections) AddHostConnection(masterData *MasterData, clientID string, conn net.Conn, sendChan chan sharedConsts.Message) {

	//remoteIP, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	newConn := MasterConnectionInfo{
		ClientIP: clientID,
		HostConn: conn,
	}

	fmt.Printf("NewConn. ClientIP: %s", newConn.ClientIP)

	ac.mutex.Lock()
	ac.Conns = append(ac.Conns, newConn)
	ac.mutex.Unlock()

	if ExistsPriorCabRequests(masterData.AllElevStates, clientID) {

		priorCabRequests := masterData.AllElevStates[clientID].CabRequests

		cabRequestsWithID := CabRequestsWithID {
			ID: clientID, 
			CabRequests: priorCabRequests,
		}

		cabRequestsWithIDJSON, err := json.Marshal(cabRequestsWithID)
		if err != nil {
			fmt.Println("Error marshalling final priorCabRequests: ", err)
			return
		}

		priorCabRequestmsg := sharedConsts.Message{
			Type:    sharedConsts.PriorCabRequestsMessage,
			Target:  sharedConsts.TargetClient,
			Payload: cabRequestsWithIDJSON,
		}

		fmt.Println("sending prior cabrequest", priorCabRequests)
		sendChan <- priorCabRequestmsg
		fmt.Println("sent prior cabrequest")
	}


	fmt.Println("Sending acitveConnections", ac.Conns)
	ac.SendActiveConnections(sendChan)
}

func ExistsPriorCabRequests(AllElevStates map[string]elevator.ElevStates, targetIP string) bool {

	elevState, exists := AllElevStates[targetIP]
	if !exists {
		fmt.Println("No elevator state found for IP:", targetIP)
		return false
	}

	for _, cabRequest := range elevState.CabRequests {
		if cabRequest {
			// If any cabRequest is true, return false (not empty)
			return true
		}
	}
	// Otherwise, the ElevStates is not empty
	return false
}

func (ac *ActiveConnections) SendActiveConnections(sendChan chan sharedConsts.Message) {

	var IPs []string
	for _, conn := range ac.Conns {
		IPs = append(IPs, conn.ClientIP)
	}

	activeConnectionsDataJSON, err := json.Marshal(IPs)
	if err != nil {
		fmt.Println("Error marshalling activeConnections: ", err)
		return
	}

	activeConnectionsMessage := sharedConsts.Message{
		Type:    sharedConsts.ActiveConnectionsMessage,
		Target:  sharedConsts.TargetClient,
		Payload: activeConnectionsDataJSON,
	}

	fmt.Println("Master sending activeConnections")
	sendChan <- activeConnectionsMessage
	fmt.Println("Master sent activeConnections")
}

func (ac *ActiveConnections) MasterSendMessages(client *ClientConnectionInfo) {

	fmt.Println("Arrived at masterSend")
	var targetConn net.Conn
	for msg := range client.Channels.SendChan {
		switch msg.Target {

		case sharedConsts.TargetMaster:
			client.Channels.MasterChan <- msg

		case sharedConsts.TargetClient:
			for clients := range ac.Conns {
				targetConn = ac.Conns[clients].HostConn
				SendMessage(client, ac, msg, targetConn)
			}
		}
	}
}

func (masterData *MasterData) HandleReceivedMessagesToMaster(ac *ActiveConnections, msg sharedConsts.Message, client *ClientConnectionInfo) {

	fmt.Println("At handleMessagesToMaster")
	switch msg.Type {
	case sharedConsts.LocalRequestMessage:

		var request elevator.ButtonEvent
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

		var elevMessage elevator.MessageToMaster
		err := json.Unmarshal(msg.Payload, &elevMessage)
		if err != nil {
			fmt.Println("Error unmarshalling payload: ", err)
			return
		}

		fmt.Printf("Received current state from elevator: %#v\n", elevMessage.ElevStates)

		// Check if the current state is valid
		if elevMessage.ElevStates.Behaviour != "" {
			ID := elevMessage.ElevStates.IP
			masterData.mutex.Lock()
			masterData.AllElevStates[ID] = elevMessage.ElevStates
			masterData.mutex.Unlock()
			requestsToDo := elevMessage.RequestsToDo
			ClearHallRequestAtCurrentFloor(requestsToDo, masterData, ID)
		}

		assignedOrders := hra.SendStateToHRA(masterData.AllElevStates, masterData.GlobalHallRequests)
		masterData.mutex.Lock()

		for ID, orders := range *assignedOrders {
			masterData.AllAssignedRequests[ID] = orders
			fmt.Println("Assigned orders for ID: ", ID, " are: ", orders)
		}
		masterData.mutex.Unlock()

		backupData := Worldview{
			GlobalHallRequests:  masterData.GlobalHallRequests,
			AllAssignedRequests: masterData.AllAssignedRequests,
		}

		// Update local backupData to keep track of what's been sent
		masterData.mutex.Lock()
		masterData.BackupData = backupData
		masterData.mutex.Unlock()

		// Marshal clientData
		fmt.Println("Data to be marshaled:", backupData)
		clientDataJSON, err := json.Marshal(backupData)
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

		if client.ID == client.HostIP {
			client.Channels.ElevatorChan <- orderMsg
		}

		// Send message to clients
		if len(ac.Conns) >= 1 {
			client.Channels.SendChan <- orderMsg
		}
	}
}

// Compares the elevator's requestsToDo with the master's assigned requests
func ClearHallRequestAtCurrentFloor(RequestsToDo [elevator.N_FLOORS][elevator.N_BUTTONS]bool, masterData *MasterData, ID string) {

	for f := 0; f < elevator.N_FLOORS; f++ {
		for btn := 0; btn < elevator.N_BUTTONS-1; btn++ {
			if RequestsToDo[f][btn] != masterData.AllAssignedRequests[ID][f][btn] {
				masterData.mutex.Lock()
				masterData.GlobalHallRequests[f][btn] = false
				masterData.mutex.Unlock()
				fmt.Println("GlobalHallRequest cleared at floor", f, "Btn:", btn)
			}
		}
	}
}
