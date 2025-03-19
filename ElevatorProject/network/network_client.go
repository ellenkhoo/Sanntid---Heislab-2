package network

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net"
	"time"

	//"github.com/ellenkhoo/ElevatorProject/heartbeat"
	//"github.com/ellenkhoo/ElevatorProject/roles"
	"github.com/ellenkhoo/ElevatorProject/elevator"
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
func (client *ClientConnectionInfo) AddClientConnection(id string, clientConn net.Conn, sendChan chan sharedConsts.Message, receiveChan chan sharedConsts.Message) {
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
			var msg sharedConsts.Message
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

func ClientSendMessages(sendChan chan sharedConsts.Message, conn net.Conn, ctx context.Context) {

	fmt.Println("Ready to send msg to master")

	encoder := json.NewEncoder(conn)
	for {
		select{
		case msg := <- sendChan:
			fmt.Println("Sending message:", msg)
			err := encoder.Encode(msg)
			if err != nil {
				fmt.Println("Error encoding message: ", err)
				return
			}
		case <- ctx.Done():
			return
		}
	}
}

// Messages sent to a client means that the data is meant both for an elevator thread and the potential backup
func (clientConn *ClientConnectionInfo)HandleReceivedMessageToClient(msg sharedConsts.Message) {

	clientID := clientConn.ID

	switch msg.Type {
	case sharedConsts.RankMessage:
		data := msg.Payload
		if rank, ok := data.(int); ok {
			fmt.Println("Setting my rank to", rank)
			clientConn.Rank = rank
			if rank == 2 {
				fmt.Println("My rank is 2 and I will become backup")
				// start backup
			}
		}
	case sharedConsts.MasterOrdersMessage:
		data := msg.Payload
		if masterData, ok := data.(MasterData); ok {
			backupData:= CreateBackupData(masterData)
			elevatorData := CreateElevatorData(masterData, clientID)

			backupMsg := sharedConsts.Message{
				Type: sharedConsts.MasterOrdersMessage,
				Target: sharedConsts.TargetBackup,
				Payload: backupData,
			}


			elevatorMsg := sharedConsts.Message{
				Type: sharedConsts.MasterOrdersMessage,
				Target: sharedConsts.TargetElevator,
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

func HandleReceivedMessageToElevator(fsm elevator.FSM, msg sharedConsts.Message) {

	// assignedRequests := msg.Payload 
	// if assignedRequests, ok := assignedRequests.(fsm.El.AssignedRequests); ok {

	// }
	// fsm.El.AssignedRequests = assignedRequests
// 	fsm.El.AssignedRequests = e.Payload.AssignedRequests
// 	fsm.El.RequestsToDo = fsm.El.AssignedRequests.append(fsm.El.ElevStates.cabRequests)
}

