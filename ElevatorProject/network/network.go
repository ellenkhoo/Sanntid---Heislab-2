package network

import (
	//"github.com/ellenkhoo/ElevatorProject/network/network_functions/bcast"
	"github.com/ellenkhoo/ElevatorProject/network/network_functions/localip"
	//"github.com/ellenkhoo/ElevatorProject/network/network_functions/peers"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"

	"github.com/ellenkhoo/ElevatorProject/elevator"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	//"time"
)

func CreateActiveConnections() *ActiveConnections {
	return &ActiveConnections{}
}

func CreateMasterData() *MasterData {
	return &MasterData{
		GlobalHallRequests:  [elevator.N_FLOORS][2]bool{},
		AllAssignedRequests: make(map[string][elevator.N_FLOORS][2]bool),
		AllElevStates:       make(map[string]elevator.ElevStates),
	}
}

func SendMessage(msg sharedConsts.Message, conn net.Conn) {
	fmt.Println("At SendMessage")
	encoder := json.NewEncoder(conn)
	err := encoder.Encode(msg)
	if err != nil {
		fmt.Println("Error encoding message: ", err)
		return
	}
}

func ReceiveMessage(receiveChan chan sharedConsts.Message, conn net.Conn) {
	fmt.Println("At func ReceiveMessage!")
	decoder := json.NewDecoder(conn)

	for {
		var msg sharedConsts.Message
		err := decoder.Decode(&msg)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed")
			}
			fmt.Println("Error decoding message: ", err)
			return
		}
		receiveChan <- msg
	}
}

func RouteMessages(client *ClientConnectionInfo, networkChannels *sharedConsts.NetworkChannels) {
	for msg := range networkChannels.ReceiveChan {
		switch msg.Target {
		case sharedConsts.TargetMaster:
			networkChannels.MasterChan <- msg
		case sharedConsts.TargetClient:
			client.HandleReceivedMessageToClient(msg)
		default:
			fmt.Println("Unknown message target")
		}
	}
}

func InitMasterSlaveNetwork(ac *ActiveConnections, client *ClientConnectionInfo, masterData *MasterData, ackTracker *AcknowledgeTracker, bcastPort string, TCPPort string, networkChannels *sharedConsts.NetworkChannels, fsm *elevator.FSM) {
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// If no ID is given, use the local IP and process ID
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = localIP
		//id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
		fmt.Printf("id: %s", id)
	}

	go RouteMessages(client, networkChannels.ReceiveChan, networkChannels)
	// Listen for the master
	var masterID string = ""
	masterID, found := ListenForMaster(bcastPort)
	if found {
		//Try to connect to the master
		clientConn, success := ConnectToMaster(masterID, TCPPort)
		if success {
			client.AddClientConnection(id, clientConn, networkChannels)
		}
	} else {
		// No master found, announce ourselves as the master
		masterID = id
		client.ID = id // local client (elevator)
		fmt.Printf("Going to announce master. MasterID: %s\n", id)
		go AnnounceMaster(id, bcastPort)
		go ac.ListenAndAcceptConnections(TCPPort, networkChannels)
		go ac.MasterSendMessages(networkChannels)
	}

	for {
		select {
		case m := <-networkChannels.MasterChan:
			fmt.Println("Master received a message")
			masterData.HandleReceivedMessagesToMaster(ac, m, networkChannels, ackTracker)
		case e := <-networkChannels.ElevatorChan:
			client.UpdateElevatorWorldview(fsm, e)
		}
	}
}
