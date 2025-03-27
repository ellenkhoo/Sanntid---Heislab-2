package network

import (
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/ellenkhoo/ElevatorProject/elevator"
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

func CreateBackupData() *BackupData {
	return &BackupData{
		GlobalHallRequests:  [elevator.N_FLOORS][2]bool{},
		AllAssignedRequests: make(map[string][elevator.N_FLOORS][2]bool),
	}
}
func SendMessage(client *ClientConnectionInfo, ac *ActiveConnections, msg sharedConsts.Message, conn net.Conn) {
	fmt.Println("At SendMessage")
	if client.ID == client.HostIP {
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

func ReceiveMessage(client *ClientConnectionInfo, ac *ActiveConnections, receiveChan chan sharedConsts.Message, conn net.Conn) {
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
		receiveChan <- msg
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
		}
	}
}

func InitSlave(ID string, masterID string, ac *ActiveConnections, client *ClientConnectionInfo, masterData *MasterData,
	bcastPort string, TCPPort string, networkChannels *sharedConsts.NetworkChannels, fsm *elevator.FSM) {

	clientConn, success := ConnectToMaster(masterID, TCPPort)
	if success {
		client.AddClientConnection(ID, clientConn, networkChannels)
		go ReceiveMessage(client, ac, networkChannels.ReceiveChan, clientConn)
		go ClientSendMessagesFromSendChan(ac, client, networkChannels.SendChan, clientConn)

		for {
			select {
			case b := <-networkChannels.BackupChan:
				client.HandleReceivedMessageToClient(b)
			case e := <-networkChannels.ElevatorChan:
				fmt.Println("Going to update my worldview")
				client.UpdateElevatorWorldview(fsm, e)
				// case r := <-networkChannels.RestartChan:
				// 	fmt.Println("Received message on restartChan:", r)
				// 	if r == "master" {
				// 		fmt.Print("AC: ", ac)
				// 		go InitMaster(ID, ac, client, masterData, bcastPort, TCPPort, networkChannels, fsm)
				// 	} else if r == "slave" {
				// 		var masterID string = ""
				// 		masterID, found := ListenForMaster(bcastPort)
				// 		if found {
				// 			go InitSlave(ID, masterID, ac, client, masterData, bcastPort, TCPPort, networkChannels, fsm)
				// 		}
				// 	}
				// 	return
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
		// for i, connInfo := range ac.Conns {
		// 	if connInfo.HostConn == conn {
		// 		ac.Conns = append(ac.Conns[:i], ac.Conns[i+1:]...)
		// 		fmt.Println("Removed connection. AC now:", ac.Conns)
		// 	}
		// }
		// fmt.Println("Going to send active connections")
		// ac.SendActiveConnections(client.Channels.SendChan)
	} else {
		// Master disconnected
		fmt.Println("Master disconnected")
		// clientID := client.ID
		// if ShouldBecomeMaster(clientID, ac) {
		// 	fmt.Println("I should become master")
		// 	msg := "master"
		// 	client.Channels.RestartChan <- msg
		// } else {
		// 	fmt.Println("I should become slave")
		// 	msg := "slave"
		// 	client.Channels.RestartChan <- msg
		// }
	}
}
