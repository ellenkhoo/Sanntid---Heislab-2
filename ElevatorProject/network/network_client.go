package network

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net"
	"time"

	//"github.com/ellenkhoo/ElevatorProject/heartbeat"
	//"github.com/ellenkhoo/ElevatorProject/roles"
	"github.com/ellenkhoo/ElevatorProject/elevator"
	elevio "github.com/ellenkhoo/ElevatorProject/elevator/Driver"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
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
func (client *ClientConnectionInfo) AddClientConnection(id string, clientConn net.Conn, channels sharedConsts.NetworkChannels) {
	//defer conn.Close()
	remoteIP, _, _ := net.SplitHostPort(clientConn.RemoteAddr().String())

	fmt.Println("Adding client connection")

	*client = ClientConnectionInfo{
		ID:          id,
		HostIP:      remoteIP,
		ClientConn:  clientConn,
		Channels: channels,
		HeartbeatTimer: time.NewTimer(5 * time.Second),
	}

	fmt.Println("Going to handle connection")
	go HandleConnection(*client)
}

// Maybe not the most describing name
func HandleConnection(client ClientConnectionInfo) {
	// Read from TCP connection and send to the receive channel
	fmt.Println("Ready to read from TCP")
	go func() {
		decoder := json.NewDecoder(client.ClientConn)
		for {
			var msg sharedConsts.Message
			err := decoder.Decode(&msg)
			if err != nil {
				fmt.Println("Error decoding message: ", err)
				return
			}
			client.Channels.ReceiveChan <- msg
		}
	}()

	// Read from the send channel and write to the TCP connection
	fmt.Println("Ready to send on TCP")
	go func() {
		encoder := json.NewEncoder(client.ClientConn)
		for msg := range client.Channels.SendChan {
			err := encoder.Encode(msg)
			if err != nil {
				fmt.Println("Error encoding message: ", err)
				return
			}
		}
	}()
}

func ClientSendMessages(sendChan chan sharedConsts.Message, conn net.Conn) {

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
func (clientConn *ClientConnectionInfo) HandleReceivedMessageToClient(msg sharedConsts.Message) {

	clientID := clientConn.ID
	clientConn.HeartbeatTimer = time.NewTimer(5 * time.Second)

	switch msg.Type {
	case sharedConsts.RankMessage:
		var rank int
		err := json.Unmarshal(msg.Payload, &rank)
		if err != nil {
			fmt.Println("Error decoding rank message: ", err)
			return
		}

		fmt.Println("Setting my rank to", rank)
		clientConn.Rank = rank
		if rank == 2 {
			fmt.Println("My rank is 2 and I will become backup")
			// start backup
		}

	case sharedConsts.MasterOrdersMessage:
		data := msg.Payload
		var masterData BackupData
		err := json.Unmarshal(data, &masterData)
		if err != nil {
			fmt.Println("Error decoding master orders message: ", err)
			return
		}

		backupData := CreateBackupData(masterData)
		elevatorData := CreateElevatorData(masterData, clientID)

		// Marshal backupData and elevatorData

		backupDataJSON, err := json.Marshal(backupData)
		if err != nil {
			fmt.Println("Error marshalling backup data: ", err)
			return
		}

		elevatorDataJSON, err := json.Marshal(elevatorData)
		if err != nil {
			fmt.Println("Error marshalling elevator data: ", err)
			return
		}

		backupMsg := sharedConsts.Message{
			Type:    sharedConsts.MasterOrdersMessage,
			Target:  sharedConsts.TargetBackup,
			Payload: backupDataJSON,
		}

		elevatorMsg := sharedConsts.Message{
			Type:    sharedConsts.MasterOrdersMessage,
			Target:  sharedConsts.TargetElevator,
			Payload: elevatorDataJSON,
		}

		fmt.Println("Sending messages to backup and elevator")
		clientConn.Channels.BackupChan <- backupMsg
		clientConn.Channels.ElevatorChan <- elevatorMsg

	case sharedConsts.Heartbeat: 
		var heartbeat string 
		err := json.Unmarshal(msg.Payload, &heartbeat)
		if err != nil {
			fmt.Println("Error decoding heartbeat message: ", err)
			return
		}
		if heartbeat == "HB" {	
		fmt.Println("Received heartbeat from master")
			if !clientConn.HeartbeatTimer.Stop(){
				select {
				case <-clientConn.HeartbeatTimer.C:
				default:
				}
			}
			clientConn.HeartbeatTimer.Reset(5 * time.Second)

			go func() {
				<-clientConn.HeartbeatTimer.C
				fmt.Println("⏳ Timeout! Assuming master is dead...")
				//HandleMasterDisconnection() // Kall en funksjon for å håndtere failover
			}()
			}

		// 	// start timer
		// case timeout:
		// 	// start master
	}
}

func (clientConn *ClientConnectionInfo) HandleReceivedMessageToElevator(fsm *elevator.FSM, msg sharedConsts.Message) {

	fmt.Println("At handleMessageToElevator\n")
	fmt.Println("Before update:", fsm.El.RequestsToDo)
	clientID := clientConn.ID
	fmt.Println("Client ID: ", clientID) // returnerer ingenting akkurat nå
	var masterData BackupData
	err := json.Unmarshal(msg.Payload, &masterData)
	if err != nil {
		fmt.Println("Error decoding message to elevator: ", err)
		return
	}

	fmt.Println("MasterData: ", masterData)
	elevatorData := CreateElevatorData(masterData, clientID)
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
			fsm.El.RequestsToDo[floor][elevio.BT_Cab] = true
		} else {
			fmt.Println("No cab request at floor: ", floor)
		}
	}

	fsm.El.AssignedRequests = assignedRequests
	fsm.El.GlobalHallRequests = globalHallRequests
	fmt.Println("After update:", fsm.El.RequestsToDo)
	fsm.Fsm_mtx.Unlock()

	clientConn.Channels.UpdateChan <- "You are ready to do things"
}

// This function returns only the assigned requests relevant to a particular elevator + globalHallRequests
func CreateElevatorData(masterData BackupData, elevatorID string) ElevatorRequest {

	localAssignedRequests := masterData.AllAssignedRequests[elevatorID]
	globalHallRequests := masterData.GlobalHallRequests

	elevatorData := ElevatorRequest{
		GlobalHallRequests: globalHallRequests,
		AssignedRequests:   localAssignedRequests,
	}

	return elevatorData
}

func CreateBackupData(masterData BackupData) BackupData {

	AllAssignedRequests := masterData.AllAssignedRequests
	globalHallRequests := masterData.GlobalHallRequests

	backupData := BackupData{
		GlobalHallRequests:  globalHallRequests,
		AllAssignedRequests: AllAssignedRequests,
	}

	return backupData
}
