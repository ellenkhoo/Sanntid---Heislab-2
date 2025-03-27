package network

import (
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/ellenkhoo/ElevatorProject/elevator"
	"github.com/ellenkhoo/ElevatorProject/network/network_functions/peers"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
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

func CreateWorldview() *Worldview {
	return &Worldview{
		GlobalHallRequests:  [elevator.N_FLOORS][2]bool{},
		AllAssignedRequests: make(map[string][elevator.N_FLOORS][2]bool),
	}
}

func CreateBackupData() *BackupData {
	return &BackupData{
		Worldview:                   *CreateWorldview(),
		MastersActiveConnectionsIPs: []string{},
	}
}

func SendMessage(client *ClientConnectionInfo, ac *ActiveConnections, msg sharedConsts.Message, conn net.Conn) {
	fmt.Println("At SendMessage")
	if client.ID == client.HostIP && !(msg.Type == sharedConsts.PriorCabRequestsMessage) {
		client.Channels.ReceiveChan <- msg
	}
	fmt.Println("The message is to a remote client")
	encoder := json.NewEncoder(conn)
	err := encoder.Encode(msg)
	if err != nil {
		fmt.Println("Error encoding message: ", err)
		if netErr, ok := err.(*net.OpError); ok && netErr.Op == "write" {
			if netErr.Err.Error() == "use of closed network connection" {
				fmt.Println("The connection is closed, unable to send message")
				HandleClosedConnection(client, ac, conn)
			}
		} else if err == io.EOF {
			fmt.Println("Connection closed")
			HandleClosedConnection(client, ac, conn)
		}
		return
	}
}

func ReceiveMessage(masterData *MasterData, client *ClientConnectionInfo, ac *ActiveConnections, networkChannels sharedConsts.NetworkChannels, conn net.Conn) {
	fmt.Println("At func ReceiveMessage!")
	decoder := json.NewDecoder(conn)

	for {
		var msg sharedConsts.Message
		err := decoder.Decode(&msg)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed")
				HandleClosedConnection(client, ac, conn)
			} else if opErr, ok := err.(*net.OpError); ok && opErr.Op == "read" {
				fmt.Println("Connection reset by peer")
				HandleClosedConnection(client, ac, conn)
			}
			fmt.Println("Error decoding message: ", err)
			break
		}
		if msg.Type == sharedConsts.ClientIDMessage {

			fmt.Println("Received clinetID msg")
			var clientID string
			err := json.Unmarshal(msg.Payload, &clientID)
			if err != nil {
				fmt.Println("Error unmarshalling payload: ", err)
				return
			}
			go ac.AddHostConnection(masterData, clientID, conn, networkChannels.SendChan)
		}

		networkChannels.ReceiveChan <- msg
	}
}

func RouteMessages(client *ClientConnectionInfo, networkChannels *sharedConsts.NetworkChannels) {
	fmt.Println("Router received msg")
	for msg := range networkChannels.ReceiveChan {
		switch msg.Target {
		case sharedConsts.TargetMaster:
			fmt.Println("Msg is to master")
			networkChannels.MasterChan <- msg
		case sharedConsts.TargetClient:
			fmt.Println("Msg is to client")
			client.HandleReceivedMessageToClient(msg)
		case sharedConsts.TargetElevator:
			fmt.Println("Msg is to elevator")
			networkChannels.ElevatorChan <- msg
		default:
			fmt.Println("Unknown message target")
		}
	}
}
func InitNetwork(ID string, ac *ActiveConnections, client *ClientConnectionInfo, masterData *MasterData,
	bcastPort string, TCPPort string, networkChannels *sharedConsts.NetworkChannels, fsm *elevator.FSM) {

	var masterID string = ""
	masterID, found := ListenForMaster(bcastPort)
	if found {
		go InitSlave(ID, masterID, ac, client, masterData, bcastPort, TCPPort, networkChannels, fsm)
	} else {
		go InitMaster(ID, ac, client, masterData, bcastPort, TCPPort, networkChannels, fsm)
	}
}

func InitMaster(masterID string, ac *ActiveConnections, client *ClientConnectionInfo, masterData *MasterData, bcastPort string, TCPPort string, networkChannels *sharedConsts.NetworkChannels, fsm *elevator.FSM) {

	peerUpdateChan := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, masterID, peerTxEnable)
	go peers.Receiver(15647, peerUpdateChan)

	client.ID = masterID
	client.HostIP = masterID
	fmt.Printf("Going to announce master. MasterID: %s\n", masterID)
	go AnnounceMaster(masterID, bcastPort)
	go ac.ListenAndAcceptConnections(masterData, client, TCPPort, networkChannels)
	go ac.MasterSendMessages(client)
	fmt.Println("AC: ", ac.Conns)

	for {
		select {
		case m := <-networkChannels.MasterChan:
			fmt.Println("Master received a message")
			go masterData.HandleReceivedMessagesToMaster(ac, m, client)
		case e := <-networkChannels.ElevatorChan:
			fmt.Println("Going to update my worldview")
			go client.UpdateElevatorWorldview(fsm, e)
		case p := <-peerUpdateChan:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			// Remove lost connection from ActiveConnections
			for _, lostID := range p.Lost {
				for j, connInfo := range ac.Conns {
					if connInfo.ClientIP == lostID {
						ac.Conns[j].ClientIP = ""
						fmt.Println("Removed IP of connection. AC now:", ac.Conns)
						break
					}
				}
			}

			// Add new connection to ac
			for _, newID := range p.New {
				for j, connInfo := range ac.Conns {
					if connInfo.ClientIP == "" {
						ac.Conns[j].ClientIP = string(newID)
						fmt.Println("Added client ID back to AC", string(newID))
						break
					}
				}
			}
		}
	}
}

func InitSlave(ID string, masterID string, ac *ActiveConnections, client *ClientConnectionInfo, masterData *MasterData,
	bcastPort string, TCPPort string, networkChannels *sharedConsts.NetworkChannels, fsm *elevator.FSM) {

	peerUpdateChan := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, ID, peerTxEnable)
	go peers.Receiver(15647, peerUpdateChan)

	clientConn, success := ConnectToMaster(masterID, TCPPort)
	if success {
		// Send message to master with ID
		idJSON, err := json.Marshal(ID)
		if err != nil {
			fmt.Println("Error marshaling ID")
		}

		IDMessage := sharedConsts.Message{
			Type:    sharedConsts.ClientIDMessage,
			Target:  sharedConsts.TargetMaster,
			Payload: idJSON,
		}

		fmt.Println("Sending ID message to master", ID)
		client.Channels.SendChan <- IDMessage
		fmt.Println("Sent ID message to master")

		client.AddClientConnection(ID, clientConn, networkChannels)
		go ReceiveMessage(masterData, client, ac, client.Channels, clientConn)
		go ClientSendMessagesFromSendChan(ac, client, networkChannels.SendChan, clientConn)

		for {
			select {
			case b := <-networkChannels.BackupChan:
				client.HandleReceivedMessageToClient(b)
			case e := <-networkChannels.ElevatorChan:
				fmt.Println("Going to update my worldview")
				go client.UpdateElevatorWorldview(fsm, e)
			case p := <-peerUpdateChan:
				fmt.Printf("Peer update:\n")
				fmt.Printf("  Peers:    %q\n", p.Peers)
				fmt.Printf("  New:      %q\n", p.New)
				fmt.Printf("  Lost:     %q\n", p.Lost)

				if len(p.Peers) == 1 {
					// start master
					fmt.Println("I am alone on the network and should become master")
				}
			case r := <-networkChannels.RestartChan:
				fmt.Println("Received message on restartChan:", r)
				if r == "master" {
					fmt.Print("AC: ", ac)
					fmt.Println("Going to try to init master")
					go InitMaster(ID, ac, client, masterData, bcastPort, TCPPort, networkChannels, fsm)
				} else if r == "slave" {
					var masterID string = ""
					fmt.Println("Listening for a new master")
					masterID, found := ListenForMaster(bcastPort)
					if found {
						go InitSlave(ID, masterID, ac, client, masterData, bcastPort, TCPPort, networkChannels, fsm)
					}
				}
				return
			}
		}
	}
}

func HandleClosedConnection(client *ClientConnectionInfo, ac *ActiveConnections, conn net.Conn) {
	fmt.Println("At handle disconnection")
	conn.Close()
	if client.ID == client.HostIP {
		// Slave disconnected
		fmt.Println("Slave disconnected")
		// Remove from active connections
		for i, connInfo := range ac.Conns {
			if connInfo.HostConn == conn {
				ac.Conns = append(ac.Conns[:i], ac.Conns[i+1:]...)
				fmt.Println("Removed connection. AC now:", ac.Conns)
			}
		}
		fmt.Println("Going to send active connections")
		ac.SendActiveConnections(client.Channels.SendChan)
	} else {
		// Master disconnected
		fmt.Println("Master disconnected")
		clientID := client.ID
		if ShouldBecomeMaster(clientID, client.BackupData.MastersActiveConnectionsIPs) {
			fmt.Println("I should become master")
			msg := "master"
			client.Channels.RestartChan <- msg
		} else {
			fmt.Println("I should become slave")
			msg := "slave"
			client.Channels.RestartChan <- msg
		}
	}
}

func ShouldBecomeMaster(clientID string, mastersActiveConnections []string) bool {
	fmt.Println("AC at master dead:", mastersActiveConnections)
	for _, ID := range mastersActiveConnections {
		fmt.Println("My ID:", clientID, "connID:", ID)
		if ID > clientID {
			return false
		}
	}
	return true
}
