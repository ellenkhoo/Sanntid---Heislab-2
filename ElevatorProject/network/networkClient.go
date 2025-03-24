package network

import (
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	"github.com/ellenkhoo/ElevatorProject/elevator"
	"encoding/json"
	"math/rand/v2"
	"time"
	"fmt"
	"net"	
)

func RandRange(min, max int) int {
	return rand.IntN(max-min) + min
}

func ListenForMaster(port string) (string, bool) {
	addr, _ := net.ResolveUDPAddr("udp", "0.0.0.0"+":"+port)
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting UDP listener:", err)
		return "", false //No existing master
	}

	defer conn.Close()

	buffer := make([]byte, 1024)
	t := time.Duration(RandRange(800, 1500))
	fmt.Printf("Waiting for %d ms\n", t)
	conn.SetReadDeadline(time.Now().Add(t * time.Millisecond)) //ensures that only one remains master
	_, remoteAddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("No master found, becoming master.")
		return "", false
	}

	fmt.Println("Master found at: ", remoteAddr.IP.String())
	return remoteAddr.IP.String(), true
}

func ConnectToMaster(masterIP string, listenPort string) (net.Conn, bool) {
	conn, err := net.Dial("tcp", masterIP+":"+listenPort)
	if err != nil {
		fmt.Println("Error connecting to master:", err)
		return nil, false
	}

	if err != nil {
		fmt.Println("Error reading from master:", err)
		conn.Close()
		return nil, false
	}

	fmt.Printf("Connected to master at %s\n: ", masterIP)
	return conn, true
}

// When a new connection is established on the client side, this function adds it to the list of active connections
func (client *ClientConnectionInfo) AddClientConnection(id string, clientConn net.Conn, networkChannels *sharedConsts.NetworkChannels) {
	//defer conn.Close()
	remoteIP, _, _ := net.SplitHostPort(clientConn.RemoteAddr().String())

	fmt.Println("Adding client connection")

	*client = ClientConnectionInfo{
		ID:         id,
		HostIP:     remoteIP,
		ClientConn: clientConn,
		Channels:   *networkChannels,
	}
}

func ClientSendMessagesFromSendChan(client *ClientConnectionInfo, sendChan chan sharedConsts.Message, conn net.Conn) {

	fmt.Println("Ready to send msg to master")
	for msg := range sendChan {
		SendMessage(client, msg, conn)
	}
}

// Messages sent to a client means that the data is meant both for an elevator thread and the potential backup
func (clientConn *ClientConnectionInfo) HandleReceivedMessageToClient(msg sharedConsts.Message) {

	clientID := clientConn.ID

	switch msg.Type {

	case sharedConsts.MasterWorldviewMessage:
		fmt.Println("Received master worldview message")
		data := msg.Payload
		var masterData BackupData
		err := json.Unmarshal(data, &masterData)
		if err != nil {
			fmt.Println("Error decoding message: ", err)
			return
		}

		backupData := UpdateBackupData(masterData)
		clientConn.ClientMtx.Lock()
		clientConn.Worldview = backupData
		clientConn.ClientMtx.Unlock()

		// Marshal backupData
		backupIDJSON, err := json.Marshal(clientID)
		if err != nil {
			fmt.Println("Error marshalling backup data: ", err)
			return
		}

		backupMsg := sharedConsts.Message{
			Type:    sharedConsts.AcknowledgeMessage,
			Target:  sharedConsts.TargetMaster,
			Payload: backupIDJSON,
		}

		fmt.Println("Sending ack")
		clientConn.Channels.SendChan <- backupMsg

	case sharedConsts.UpdateOrdersMessage:

		if clientID != clientConn.HostIP {
			fmt.Println("I am not on the master computer")
			elevatorData := UpdateElevatorData(clientConn.Worldview, clientID)

			elevatorDataJSON, err := json.Marshal(elevatorData)
			if err != nil {
				fmt.Println("Error marshalling backup data: ", err)
				return
			}

			elevatorMsg := sharedConsts.Message{
				Payload: elevatorDataJSON,
			}

			clientConn.Channels.ElevatorChan <- elevatorMsg
		}
		
		
	// case heartbeat: //
	// 	// start timer
	// case timeout:
	// 	// start master
	}
}

func (clientConn *ClientConnectionInfo) UpdateElevatorWorldview(fsm *elevator.FSM, msg sharedConsts.Message) {

	fmt.Println("At UpdateElevatorWorldview\n")
	fmt.Println("RequestsToDo before update:", fsm.El.RequestsToDo)
	clientID := clientConn.ID

	var masterData BackupData
	err := json.Unmarshal(msg.Payload, &masterData)
	if err != nil {
		fmt.Println("Error decoding message to elevator: ", err)
		return
	}

	elevatorData := UpdateElevatorData(masterData, clientID)
	fmt.Println("ElevatorData: ", elevatorData)
	fmt.Println("CabRequests: ", fsm.El.ElevStates.CabRequests)

	assignedRequests := elevatorData.AssignedRequests
	globalHallRequests := elevatorData.GlobalHallRequests

	fsm.Fsm_mtx.Lock()

	// requestsToDo = assigend requests + cab requests
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
	fmt.Println("RequestsToDo after update:", fsm.El.RequestsToDo)
	fmt.Println("GlobalHallRequests: ", fsm.El.GlobalHallRequests)
	fsm.Fsm_mtx.Unlock()

	sendMsg := "You are ready to do things"
	clientConn.Channels.UpdateChan <- sendMsg
}

// This function returns only the assigned requests relevant to a particular elevator + globalHallRequests
func UpdateElevatorData(backupData BackupData, elevatorID string) ElevatorRequest {

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

func UpdateBackupData(masterData BackupData) BackupData {

	AllAssignedRequests := masterData.AllAssignedRequests
	globalHallRequests := masterData.GlobalHallRequests

	backupData := BackupData{
		GlobalHallRequests:  globalHallRequests,
		AllAssignedRequests: AllAssignedRequests,
	}

	return backupData
}
