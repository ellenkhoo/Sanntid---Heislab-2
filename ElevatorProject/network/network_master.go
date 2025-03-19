package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/ellenkhoo/ElevatorProject/elevator"
	elevio "github.com/ellenkhoo/ElevatorProject/elevator/Driver"
	"github.com/ellenkhoo/ElevatorProject/hra"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
)


func AnnounceMaster(localIP string, port string, ctx context.Context) {
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

		for{
			select{
			case <- ctx.Done():
				return
			default:
				msg := "I am Master"
			conn.Write([]byte(msg))
			time.Sleep(1 * time.Second) //announces every 2nd second, maybe it should happen more frequently?
			}
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

	msg := sharedConsts.Message{
		Type:    sharedConsts.RankMessage,
		Target:  sharedConsts.TargetBackup,
		Payload: rank,
	}

	fmt.Println("Sending rank message on channel")
	SendMessageOnChannel(sendChan, msg)
}

// Master listenes and accepts connections
func (ac *ActiveConnections) ListenAndAcceptConnections(port string, sendChan chan sharedConsts.Message, receiveChan chan sharedConsts.Message, ctx context.Context) {

	ln, _ := net.Listen("tcp", ":"+port)

	for {
		select{
		case <- ctx.Done():
			return
		default:
			hostConn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error acepting connection:", err)
			continue
		}
		rank := len(ac.Conns) + 2

		go ReceiveMessage(receiveChan, hostConn, ctx)
		go ac.AddHostConnection(rank, hostConn, sendChan)
		}
	}
}

func (ac *ActiveConnections) MasterSendMessages(sendChan chan sharedConsts.Message, ctx context.Context) {

	fmt.Println("Arrived at masterSend")

	var targetConn net.Conn

	for{
		select{
		case msg := <- sendChan:
			fmt.Println("target: ", msg.Target)
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
	
			case sharedConsts.TargetElevator:
				// do something
			case sharedConsts.TargetClient:
	
				// do something
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
		case <- ctx.Done():
			return
		}
	}
}


func (masterData *MasterData)HandleReceivedMessagesToMaster(msg sharedConsts.Message) {

	switch msg.Type {
	case sharedConsts.LocalRequestMessage:
		// Update GlobalHallRequests
		request := msg.Payload
		if request, ok := request.(elevio.ButtonEvent); ok {
			fmt.Println("Received request: ", request)
			floor := request.Floor
			button := request.Button
			masterData.mutex.Lock()
			masterData.GlobalHallRequests[floor][button] = true
			masterData.mutex.Unlock()
		}
		
	case sharedConsts.CurrentStateMessage:
		// Update allElevStates
		elevState := msg.Payload
		if elevState, ok := elevState.(elevator.ElevStates); ok {
			fmt.Printf("Received current state from elevator: %#v\n", elevState)
			ID := elevState.IP
			masterData.mutex.Lock()
			masterData.AllElevStates[ID] = elevState
			masterData.mutex.Unlock()
			hra.SendStateToHRA(masterData.AllElevStates, masterData.GlobalHallRequests)
		}
	}
}

// func StartMaster() {

// 	fmt.Println("Starting master")

// 	var allElevStates = make(map[string]elevator.ElevStates)
// 	var globalHallRequests [][2]bool

// 	select {}
// }
