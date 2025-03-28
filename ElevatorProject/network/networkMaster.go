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

func (ac *ActiveConnections) ListenAndAcceptConnections(masterData *MasterData, client *ClientInfo, port string, networkChannels *sharedConsts.NetworkChannels) {

	ln, _ := net.Listen("tcp", ":"+port)

	for {
		hostConn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error acepting connection:", err)
			continue
		}

		tcpConn, err := ConfigureTCPConn(hostConn)
		if err != nil {
			fmt.Println("Failed to configure TCP settings")
			hostConn.Close()
			continue
		}

		go ReceiveTCPMessage(masterData, client, ac, client.Channels, tcpConn)
	}
}

// Adds the host's connection with a client to ActiveConnections
func (ac *ActiveConnections) MasterAddConnection(masterData *MasterData, clientID string, conn net.Conn, sendChan chan sharedConsts.Message) {

	newConn := MasterConnectionInfo{
		ClientID: clientID,
		HostConn: conn,
	}

	ac.AC_mutex.Lock()
	ac.Conns = append(ac.Conns, newConn)
	ac.AC_mutex.Unlock()

	if ExistsPriorCabRequests(masterData.AllElevStates, clientID) {
		SendPriorCabRequests(masterData, clientID, sendChan)
	}

	ac.SendActiveConnectionsToClient(sendChan)
}

func ExistsPriorCabRequests(AllElevStates map[string]elevator.ElevStates, targetID string) bool {

	elevState, exists := AllElevStates[targetID]
	if !exists {
		fmt.Println("No elevator state found for IP:", targetID)
		return false
	}

	for _, cabRequest := range elevState.CabRequests {
		if cabRequest {
			return true
		}
	}

	return false
}

func SendPriorCabRequests(masterData *MasterData, clientID string, sendChan chan sharedConsts.Message) {
	priorCabRequests := masterData.AllElevStates[clientID].CabRequests

	cabRequestsWithID := CabRequestsWithID{
		ID:          clientID,
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

	sendChan <- priorCabRequestmsg
}

func (ac *ActiveConnections) SendActiveConnectionsToClient(sendChan chan sharedConsts.Message) {

	var IPs []string
	for _, conn := range ac.Conns {
		IPs = append(IPs, conn.ClientID)
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

	sendChan <- activeConnectionsMessage
}

func (ac *ActiveConnections) MasterSendMessages(client *ClientInfo) {

	var targetConn net.Conn
	for msg := range client.Channels.SendChan {
		switch msg.Target {

		case sharedConsts.TargetMaster:
			client.Channels.MasterChan <- msg

		case sharedConsts.TargetClient:
			for clients := range ac.Conns {
				targetConn = ac.Conns[clients].HostConn
				SendTCPMessage(client, ac, msg, targetConn)
			}
		}
	}
}

func (masterData *MasterData) HandleReceivedMessageToMaster(ac *ActiveConnections, msg sharedConsts.Message, client *ClientInfo) {

	switch msg.Type {
	case sharedConsts.LocalHallRequestMessage:

		var request elevator.ButtonEvent
		err := json.Unmarshal(msg.Payload, &request)
		if err != nil {
			fmt.Println("Error unmarshalling payload: ", err)
			return
		}

		floor := request.Floor
		button := request.Button
		masterData.MasterData_mutex.Lock()
		masterData.GlobalHallRequests[floor][button] = true
		masterData.MasterData_mutex.Unlock()

	case sharedConsts.CurrentStateMessage:

		var elevMessage elevator.MessageToMaster
		err := json.Unmarshal(msg.Payload, &elevMessage)
		if err != nil {
			fmt.Println("Error unmarshalling payload: ", err)
			return
		}

		// Check if the current state is valid
		if elevMessage.ElevStates.Behaviour != "" {
			ID := elevMessage.ElevStates.ID
			masterData.MasterData_mutex.Lock()
			masterData.AllElevStates[ID] = elevMessage.ElevStates
			masterData.MasterData_mutex.Unlock()
			requestsToDo := elevMessage.RequestsToDo
			ClearHallRequestAtCurrentFloor(requestsToDo, masterData, ID)
		}

		activeElevStates := make(map[string]elevator.ElevStates)
		for _, conn := range ac.Conns {
			if conn.ClientID != "" {
				activeElevStates[conn.ClientID] = masterData.AllElevStates[conn.ClientID]
			}
		}
		activeElevStates[client.ID] = masterData.AllElevStates[client.ID]

		assignedOrders := hra.HallRequestAssigner(activeElevStates, masterData.GlobalHallRequests)

		masterData.MasterData_mutex.Lock()
		for ID, orders := range *assignedOrders {
			masterData.AllAssignedRequests[ID] = orders
		}
		masterData.MasterData_mutex.Unlock()

		backupData := GlobalRequestsWorldview{
			GlobalHallRequests:  masterData.GlobalHallRequests,
			AllAssignedRequests: masterData.AllAssignedRequests,
		}

		// Update local backupData to keep track of what's been sent
		masterData.MasterData_mutex.Lock()
		masterData.BackupData = backupData
		masterData.MasterData_mutex.Unlock()

		clientDataJSON, err := json.Marshal(backupData)
		if err != nil {
			fmt.Println("Error marshalling clientData: ", err)
			return
		}

		orderMsg := sharedConsts.Message{
			Type:    sharedConsts.MasterWorldviewMessage,
			Target:  sharedConsts.TargetClient,
			Payload: clientDataJSON,
		}

		// Send message to master's elevator
		if client.ID == client.HostID {
			client.Channels.ElevatorChan <- orderMsg
		}

		// Send message to clients
		if len(ac.Conns) >= 1 {
			client.Channels.SendChan <- orderMsg
		}
	}
}

// Compares the elevator's requestsToDo with the master's assigned requests for the relevant elevator
func ClearHallRequestAtCurrentFloor(RequestsToDo [elevator.N_FLOORS][elevator.N_BUTTONS]bool, masterData *MasterData, ID string) {

	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		for button := 0; button < elevator.N_BUTTONS-1; button++ {
			if RequestsToDo[floor][button] != masterData.AllAssignedRequests[ID][floor][button] {
				masterData.MasterData_mutex.Lock()
				masterData.GlobalHallRequests[floor][button] = false
				masterData.MasterData_mutex.Unlock()
			}
		}
	}
}
