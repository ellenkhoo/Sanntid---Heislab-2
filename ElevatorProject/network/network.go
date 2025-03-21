package network

import (
	//"github.com/ellenkhoo/ElevatorProject/network/network_functions/bcast"
	"github.com/ellenkhoo/ElevatorProject/network/network_functions/localip"
	//"github.com/ellenkhoo/ElevatorProject/network/network_functions/peers"
	"encoding/json"
	"flag"
	"fmt"
	"net"

	"github.com/ellenkhoo/ElevatorProject/elevator"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	//"time"
)

func CreateActiveConnections() *ActiveConnections {
	return &ActiveConnections{}
}

func ReceiveMessage(receiveChan chan sharedConsts.Message, conn net.Conn) {
	fmt.Println("At func ReceiveMessage!")
	decoder := json.NewDecoder(conn)

	for {
		var msg sharedConsts.Message
		err := decoder.Decode(&msg)
		if err != nil {
			fmt.Println("Error decoding message: ", err)
			return
		}
		receiveChan <- msg
	}
}

func RouteMessages(client *(ClientConnectionInfo), receiveChan chan sharedConsts.Message, networkChannels sharedConsts.NetworkChannels) {
	for msg := range receiveChan {
		switch msg.Target {
		case sharedConsts.TargetMaster:
			networkChannels.MasterChan <- msg
		case sharedConsts.TargetBackup:
			networkChannels.BackupChan <- msg
		case sharedConsts.TargetElevator:
			networkChannels.ElevatorChan <- msg
		case sharedConsts.TargetClient:
			// Messages that all clients should receive
			client.HandleReceivedMessageToClient(msg)

		default:
			fmt.Println("Unknown message target")
		}
	}
}

func InitMasterSlaveNetwork(ac *ActiveConnections, client *ClientConnectionInfo, masterData MasterData, bcastPortInt int, bcastPortString string, peersPort int, TCPPort string, networkChannels sharedConsts.NetworkChannels, fsm *elevator.FSM) {
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
	masterID, found := ListenForMaster(bcastPortString)
	if found {
		//Try to connect to the master
		clientConn, success := ConnectToMaster(masterID, TCPPort)
		if success {
			client.AddClientConnection(id, clientConn, networkChannels)
		}
		go ReceiveMessage(networkChannels.ReceiveChan, clientConn)
		go ClientSendMessages(networkChannels.SendChan, clientConn)
		go client.ClientSendHeartbeats(networkChannels.SendChan)
		
	} else {
		// This whole part should be startMaster() ?
		// No master found, announce ourselves as the master
		masterID = id
		client.ID = id
		fmt.Printf("Going to announce master. MasterID: %s\n", id)
		go AnnounceMaster(id, bcastPortString)
		go ac.ListenAndAcceptConnections(TCPPort, networkChannels)
		go ac.MasterSendMessages(networkChannels)
		go ac.MasterSendHeartbeats(networkChannels.SendChan)
		//go startMaster()
	}

	// Main loop to handle peer updates and hello message reception
	fmt.Println("Started")
	for {
		select {
		// case r := <-networkChannels.ReceiveChan: // don't really need this case, just for temporary logging
		// 	fmt.Println("Received a message")
		// 	fmt.Printf("Received: %#v\n", r)

		case m := <-networkChannels.MasterChan:
			fmt.Println("Master received a message")
			masterData.HandleReceivedMessagesToMaster(m, networkChannels)

		// case b := <-networkChannels.BackupChan:
		// fmt.Println("Got a message from master to backup")
		// fmt.Printf("Received: %#v\n", b)
		// msg:= Message{
		// 	Type: currentStateMessage,
		// 	Target: TargetMaster,
		// 	Payload: elevst,
		// }

		//Overskriver requests lokalt om man får ny melding fra master
		case e := <-networkChannels.ElevatorChan:
			// fmt.Printf("Elevator received: %#v\n", e)
			client.HandleReceivedMessageToElevator(fsm, e)

			// case a := <-helloRx:
			// 	fmt.Printf("Received: %#v\n", a)
		}
	}
}
