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
	// addr, _ := net.ResolveUDPAddr("udp", "255.255.255.255:9999")
	conn, err := net.Dial("udp", broadcastAddr)
	//conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Error starting UDP listener:", err)
		return
	}
	defer conn.Close()

	for {
		msg := "I am Master"
		conn.Write([]byte(msg))
		time.Sleep(1 * time.Second) //announces every 2nd second, maybe it should happen more frequently?
	}
}

// Adds the host's connection with the relevant client in the list of active connections
func (ac *ActiveConnections) AddHostConnection(rank int, conn net.Conn, sendChan chan sharedConsts.Message) {

	remoteIP, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	newConn := MasterConnectionInfo{
		ClientIP: remoteIP,
		Rank:     rank,
		HostConn: conn,
	}

	fmt.Printf("NewConn. ClientIP: %s, Rank: %d", newConn.ClientIP, newConn.Rank)

	ac.mutex.Lock()
	ac.Conns = append(ac.Conns, newConn)
	ac.mutex.Unlock()

	// Marshal rank
	rankJSON, err := json.Marshal(rank)
	if err != nil {
		fmt.Println("Error marshalling rank: ", err)
		return
	}

	msg := sharedConsts.Message{
		Type:    sharedConsts.RankMessage,
		Target:  sharedConsts.TargetBackup,
		Payload: rankJSON,
	}

	fmt.Println("Sending rank message on channel")
	sendChan <- msg
}

// Master listenes and accepts connections
func (ac *ActiveConnections) ListenAndAcceptConnections(port string, networkChannels sharedConsts.NetworkChannels) {

	ln, _ := net.Listen("tcp", ":"+port)

	for {
		hostConn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error acepting connection:", err)
			continue
		}
		rank := len(ac.Conns) + 2

		go ReceiveMessage(networkChannels.ReceiveChan, hostConn)
		go ac.AddHostConnection(rank, hostConn, networkChannels.SendChan)
	}
}

func (ac *ActiveConnections) MasterSendMessages(networkChannels sharedConsts.NetworkChannels) {

	fmt.Println("Arrived at masterSend")

	var targetConn net.Conn
	for msg := range networkChannels.SendChan {

		switch msg.Target {
		case sharedConsts.TargetBackup:
			// Need to find the conn object connected to backup
			fmt.Println("Backup is target")
			for i := range ac.Conns {
				if ac.Conns[i].Rank == 2 {
					targetConn = ac.Conns[i].HostConn
					fmt.Println("Found backup conn")
					break
				}
			}

			if targetConn != nil {
				encoder := json.NewEncoder(targetConn)
				fmt.Println("Sending message:", msg)
				err := encoder.Encode(msg)
				if err != nil {
					fmt.Println("Error encoding message: ", err)
					return
				}
			} else {
				// If targetConn is nil, log a message or handle the case
				fmt.Println("No valid connection found for the message")
			}

		case sharedConsts.TargetMaster:
			networkChannels.MasterChan <- msg

		case sharedConsts.TargetElevator:
			// do something
		case sharedConsts.TargetClient:
			// Send to the local elevator
			networkChannels.ElevatorChan <- msg
			// Send to all other remote clients
			for clients := range ac.Conns {
				targetConn = ac.Conns[clients].HostConn
				encoder := json.NewEncoder(targetConn)
				fmt.Println("Sending message:", msg)
				err := encoder.Encode(msg)
				if err != nil {
					fmt.Println("Error encoding message: ", err)
					return
				}
			}
		}
	}
}

func (masterData *MasterData) HandleReceivedMessagesToMaster(msg sharedConsts.Message, networkChannels sharedConsts.NetworkChannels) {

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
			Type:    sharedConsts.MasterOrdersMessage,
			Target:  sharedConsts.TargetClient,
			Payload: clientDataJSON,
		}
		// Send message
		networkChannels.SendChan <- orderMsg

	}
}

// func StartMaster() {

// 	fmt.Println("Starting master")

// 	var allElevStates = make(map[string]elevator.ElevStates)
// 	var globalHallRequests [][2]bool

// 	select {}
// }
