package network

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net"
	"time"

	"github.com/ellenkhoo/ElevatorProject/elevator"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

func ListenForMaster(port string) (string, bool) {
	addr, _ := net.ResolveUDPAddr("udp", "0.0.0.0"+":"+port)
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting UDP listener:", err)
		return "", false //No existing master
	}

	defer conn.Close()

	buffer := make([]byte, 1024)

	// Each program listens for a random time, t, to ensure only one becomes master
	t := time.Duration(RandRange(800, 1500))
	conn.SetReadDeadline(time.Now().Add(t * time.Millisecond))
	_, remoteAddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("No master found, becoming master.")
		return "", false
	}

	fmt.Println("Master found at: ", remoteAddr.IP.String())
	return remoteAddr.IP.String(), true
}

func RandRange(min, max int) int {
	return rand.IntN(max-min) + min
}

func ConnectToMaster(masterIP string, listenPort string) (net.Conn, bool) {
	conn, err := net.Dial("tcp", masterIP+":"+listenPort)
	if err != nil {
		fmt.Println("Error connecting to master:", err)
		return nil, false
	}

	tcpConn, err := ConfigureTCPConn(conn)
	if err != nil {
		fmt.Println("Error reading from master:", err)
		conn.Close()
		return nil, false
	}

	fmt.Printf("Connected to master at %s\n: ", masterIP)
	return tcpConn, true
}

// When a new connection is established on the client side, this function updates clientConnctionInfo
func (client *ClientConnectionInfo) AddClientConnection(id string, clientConn net.Conn, networkChannels *sharedConsts.NetworkChannels) {

	fmt.Println("Adding client connection")

	*client = ClientConnectionInfo{
		ID:         id,
		ClientConn: clientConn,
		Channels:   *networkChannels,
	}
}

func ClientSendMessagesFromSendChan(ac *ActiveConnections, client *ClientConnectionInfo, sendChan chan sharedConsts.Message, conn net.Conn) {

	fmt.Println("Ready to send msg to master")
	for msg := range sendChan {
		SendMessage(client, ac, msg, conn)
	}

	select {
	case _, ok := <-client.Channels.RestartChan:
		if !ok {
			fmt.Println("Attempting to shutdown ReceiveMessage")
			return
		}
	default:
		//Do nothing
	}
}

func (clientConn *ClientConnectionInfo) HandleReceivedMessageToClient(msg sharedConsts.Message) {

	switch msg.Type {

	case sharedConsts.MasterWorldviewMessage:
		fmt.Println("Received master worldview message")
		data := msg.Payload
		var mastersWorldview Worldview
		err := json.Unmarshal(data, &mastersWorldview)
		if err != nil {
			fmt.Println("Error decoding message: ", err)
			return
		}

		clientConn.ClientMtx.Lock()
		clientConn.BackupData.Worldview = mastersWorldview
		clientConn.ClientMtx.Unlock()

		elevatorDataJSON, err := json.Marshal(mastersWorldview)
		if err != nil {
			fmt.Println("Error marshalling backup data: ", err)
			return
		}

		elevatorMsg := sharedConsts.Message{
			Payload: elevatorDataJSON,
		}

		clientConn.Channels.ElevatorChan <- elevatorMsg
	case sharedConsts.ActiveConnectionsMessage:
		fmt.Println("Received active connections message")

		data := msg.Payload
		var connectionData []string
		err := json.Unmarshal(data, &connectionData)
		if err != nil {
			fmt.Println("Error decoding message: ", err)
			return
		}

		clientConn.ClientMtx.Lock()
		clientConn.BackupData.MastersActiveConnectionsIPs = connectionData
		clientConn.ClientMtx.Unlock()
		fmt.Println("done updating activeconnections:", clientConn.BackupData.MastersActiveConnectionsIPs)

	case sharedConsts.PriorCabRequestsMessage:
		fmt.Print("sending prior cabrequests to elevatorchan")
		clientConn.Channels.ElevatorChan <- msg
		fmt.Print("sent prior cabrequests to elevatorchan")
	}
}

func (clientConn *ClientConnectionInfo) UpdateElevatorWorldview(fsm *elevator.FSM, msg sharedConsts.Message) {

	fmt.Println("At UpdateElevatorWorldview\n")
	clientID := clientConn.ID

	var mastersWorldview Worldview
	var priorCabRequestsWithID CabRequestsWithID
	err1 := json.Unmarshal(msg.Payload, &mastersWorldview)
	if err1 != nil {
		fmt.Println("Error decoding worldview message to elevator: ", err1)

	}

	err2 := json.Unmarshal(msg.Payload, &priorCabRequestsWithID)
	fmt.Println("my id in cabRequest stuff:\n", clientID)
	if err2 != nil {
		fmt.Println("Error decoding cabRequest message to elevator: ", err2)
	}

	fmt.Println("prior id", priorCabRequestsWithID.ID)
	if clientID == priorCabRequestsWithID.ID {
		mergedCabRequests := MergeCabRequests(fsm.El.ElevStates.CabRequests, priorCabRequestsWithID.CabRequests)
		fsm.Fsm_mtx.Lock()
		fsm.El.ElevStates.CabRequests = mergedCabRequests
		fsm.Fsm_mtx.Unlock()
	}

	elevatorData := UpdateElevatorData(mastersWorldview, clientID)

	assignedRequests := elevatorData.AssignedRequests
	globalHallRequests := elevatorData.GlobalHallRequests

	fsm.Fsm_mtx.Lock()
	// requestsToDo = assigendRequests + cabRequests
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		for button := 0; button < elevator.N_BUTTONS-1; button++ {
			if assignedRequests[floor][button] {
				fmt.Println("Assigned request at floor: ", floor, " button: ", button)
				fsm.El.RequestsToDo[floor][button] = true
			}
		}

		if fsm.El.ElevStates.CabRequests[floor] {
			fmt.Println("Assigned cab request at floor: ", floor)
			fsm.El.RequestsToDo[floor][elevator.B_Cab] = true
		} else {
			fmt.Println("No cab request at floor: ", floor)
		}
	}

	fsm.El.AssignedRequests = assignedRequests
	fsm.El.GlobalHallRequests = globalHallRequests
	fsm.Fsm_mtx.Unlock()

	sendMsg := "You have an updated worldview"
	clientConn.Channels.UpdateChan <- sendMsg
}

// This function returns only the assigned requests relevant to a particular elevator + globalHallRequests
func UpdateElevatorData(backupData Worldview, elevatorID string) ElevatorRequest {

	fmt.Println("My id: ", elevatorID)
	localAssignedRequests := backupData.AllAssignedRequests[elevatorID]
	fmt.Println("assigned requests to me", localAssignedRequests)
	globalHallRequests := backupData.GlobalHallRequests

	elevatorData := ElevatorRequest{
		GlobalHallRequests: globalHallRequests,
		AssignedRequests:   localAssignedRequests,
	}

	return elevatorData
}

func MergeCabRequests(currentCabRequests [elevator.N_FLOORS]bool, priorCabRequests [elevator.N_FLOORS]bool) [elevator.N_FLOORS]bool {
	fmt.Println("at merge cabrequests")
	fmt.Println("current cabrequests:", currentCabRequests)
	fmt.Println("prior cabrequests:", priorCabRequests)

	mergedCabRequests := [elevator.N_FLOORS]bool{false, false, false, false}
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		if currentCabRequests[floor] || priorCabRequests[floor] {
			mergedCabRequests[floor] = true
		}
	}

	return mergedCabRequests
}
