package network

import (
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/ellenkhoo/ElevatorProject/elevator"
	"github.com/ellenkhoo/ElevatorProject/network/networkResources/peers"
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

func CreateWorldview() *GlobalRequestsWorldview {
	return &GlobalRequestsWorldview{
		GlobalHallRequests:  [elevator.N_FLOORS][2]bool{},
		AllAssignedRequests: make(map[string][elevator.N_FLOORS][2]bool),
	}
}

func CreateBackupData() *BackupData {
	return &BackupData{
		GlobalRequestsWorldview:     *CreateWorldview(),
		MastersActiveConnectionsIDs: []string{},
	}
}

func SendTCPMessage(client *ClientInfo, ac *ActiveConnections, msg sharedConsts.Message, conn net.Conn) {
	// If the message should be sent to the master's elevator
	if client.ID == client.HostID && !(msg.Type == sharedConsts.PriorCabRequestsMessage) {
		client.Channels.ReceiveChan <- msg
	}

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

func ReceiveTCPMessage(masterData *MasterData, client *ClientInfo, ac *ActiveConnections, networkChannels sharedConsts.NetworkChannels, conn net.Conn) {
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

			var clientID string
			err := json.Unmarshal(msg.Payload, &clientID)
			if err != nil {
				fmt.Println("Error unmarshalling payload: ", err)
				return
			}
			go ac.MasterAddConnection(masterData, clientID, conn, networkChannels.SendChan)
		}

		networkChannels.ReceiveChan <- msg
	}
}

func RouteMessages(client *ClientInfo, networkChannels *sharedConsts.NetworkChannels) {
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

func InitNetwork(ID string, ac *ActiveConnections, client *ClientInfo, masterData *MasterData,
	bcastPort string, TCPPort string, networkChannels *sharedConsts.NetworkChannels, fsm *elevator.FSM) {

	var masterID string = ""
	masterID, found := ListenForMaster(bcastPort)
	if found {
		go InitSlave(ID, masterID, ac, client, masterData, bcastPort, TCPPort, networkChannels, fsm)
	} else {
		go InitMaster(ID, ac, client, masterData, bcastPort, TCPPort, networkChannels, fsm)
	}
}

func InitMaster(masterID string, ac *ActiveConnections, client *ClientInfo, masterData *MasterData,
	bcastPort string, TCPPort string, networkChannels *sharedConsts.NetworkChannels, fsm *elevator.FSM) {

	// Initialize peer update to keep track of what connections are offline
	peerUpdateChan := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, masterID, peerTxEnable)
	go peers.Receiver(15647, peerUpdateChan)

	client.ID = masterID
	client.HostID = masterID
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

			// Remove lost connection from ActiveConnections
			for _, lostID := range p.Lost {
				for j, connInfo := range ac.Conns {
					if connInfo.ClientID == lostID {
						ac.Conns[j].ClientID = ""
						break
					}
				}
			}

			// Add new connection to ac
			if len(p.New) == len(masterID) {
				for j, connInfo := range ac.Conns {
					if connInfo.ClientID == "" {
						ac.Conns[j].ClientID = p.New
						break
					}
				}
			}
		}
	}
}

func InitSlave(ID string, masterID string, ac *ActiveConnections, client *ClientInfo, masterData *MasterData,
	bcastPort string, TCPPort string, networkChannels *sharedConsts.NetworkChannels, fsm *elevator.FSM) {

	// Initialize peer update to check if slave disconnects from
	// network at any time and should become its own master
	peerUpdateChan := make(chan peers.PeerUpdate)
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

		client.Channels.SendChan <- IDMessage

		client.ClientAddConnection(ID, clientConn, networkChannels)
		go ReceiveTCPMessage(masterData, client, ac, client.Channels, clientConn)
		go ClientSendMessagesFromSendChan(ac, client, networkChannels.SendChan, clientConn)

		for {
			select {
			case b := <-networkChannels.BackupChan:
				client.HandleReceivedMessageToClient(b)
			case e := <-networkChannels.ElevatorChan:
				fmt.Println("Going to update my worldview")
				go client.UpdateElevatorWorldview(fsm, e)
			case p := <-peerUpdateChan:

				if len(p.Peers) == 1 {
					// Become master
				}

			case r := <-networkChannels.RestartChan:
				if r == "master" {
					// Init master
					go InitMaster(ID, ac, client, masterData, bcastPort, TCPPort, networkChannels, fsm)
				} else if r == "slave" {
					// Stay slave, but listen for new master
					var masterID string = ""
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

func HandleClosedConnection(client *ClientInfo, ac *ActiveConnections, conn net.Conn) {
	fmt.Println("At handle disconnection")
	conn.Close()
	if client.ID == client.HostID {
		// Slave disconnected, remove connection from ActiveConnections
		for i, connInfo := range ac.Conns {
			if connInfo.HostConn == conn {
				ac.Conns = append(ac.Conns[:i], ac.Conns[i+1:]...)
			}
		}
		ac.SendActiveConnectionsToClient(client.Channels.SendChan)
	} else {
		// Master disconnected
		clientID := client.ID
		if ShouldBecomeMaster(clientID, client.BackupData.MastersActiveConnectionsIDs) {
			msg := "master"
			client.Channels.RestartChan <- msg
		} else {
			msg := "slave"
			client.Channels.RestartChan <- msg
		}
	}
}

func ShouldBecomeMaster(clientID string, mastersActiveConnections []string) bool {
	for _, ID := range mastersActiveConnections {
		if ID > clientID {
			return false
		}
	}
	return true
}
