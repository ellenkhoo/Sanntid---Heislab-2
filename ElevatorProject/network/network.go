package network

import (
	"github.com/ellenkhoo/ElevatorProject/comm"

	"github.com/ellenkhoo/ElevatorProject/network/network_functions/bcast"
	"github.com/ellenkhoo/ElevatorProject/network/network_functions/localip"
	"github.com/ellenkhoo/ElevatorProject/network/network_functions/peers"
	"github.com/ellenkhoo/ElevatorProject/elevator"

	"encoding/json"
	"flag"
	"fmt"
	"net"

	// "os"
	"time"
)

func CreateActiveConnections() *ActiveConnections {
	return &ActiveConnections{}
}


func ReceiveMessage(receiveChan chan Message, conn net.Conn) {
	fmt.Println("At func ReceiveMessage!")
	decoder := json.NewDecoder(conn)

	for {
		var msg Message
		err := decoder.Decode(&msg)
		if err != nil {
			fmt.Println("Error decoding message: ", err)
			return
		}

		receiveChan <-msg
	}
}

func SendMessageOnChannel(sendChan chan Message, msg Message) {
	fmt.Println("Sending msg on chan")
	sendChan <- msg
}

func RouteMessages(client *ClientConnectionInfo, receiveChan chan Message, networkChannels NetworkChannels) {
	for msg := range receiveChan {
		switch msg.Target {
		case TargetMaster:
			networkChannels.MasterChan <- msg
		case TargetBackup:
			fmt.Print("Sending msg on backup chan")
			networkChannels.BackupChan <- msg
		case TargetElevator:
			networkChannels.ElevatorChan <- msg
		case TargetClient:
			// Messages that all clients should receive
			client.HandleReceivedMessageToClient(msg)
			
		default:
			fmt.Println("Unknown message target")
		}
	}
}

func StartNetwork(ac *ActiveConnections, client ClientConnectionInfo, masterData MasterData, bcastPortInt int, bcastPortString string, peersPort int, TCPPort string, fsm elevator.FSM) NetworkChannels {
	networkChannels := NetworkChannels{
		sendChan : make(chan Message),
		receiveChan : make(chan Message),
		MasterChan:   make(chan Message),
		BackupChan:   make(chan Message),
		ElevatorChan: make(chan Message),
	}

	go InitMasterSlaveNetwork(ac, client, masterData, bcastPortInt, bcastPortString, peersPort, TCPPort, networkChannels, fsm)
	go StartHeartbeat(ac, networkChannels.MasterChan, networkChannels.BackupChan, bcastPortInt, bcastPortString, peersPort, TCPPort, networkChannels)

	return networkChannels
}


func InitMasterSlaveNetwork(ac *ActiveConnections, client ClientConnectionInfo, masterData MasterData, bcastPortInt int, bcastPortString string, peersPort int, TCPPort string, networkChannels NetworkChannels, fsm elevator.FSM) {
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

	// Start necessary channels for broadcasting, listening, and peer updates
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(peersPort, id, peerTxEnable)
	go peers.Receiver(peersPort, peerUpdateCh)

	helloTx := make(chan HelloMsg)
	helloRx := make(chan HelloMsg)
	go bcast.Transmitter(bcastPortInt, helloTx)
	go bcast.Receiver(bcastPortInt, helloRx)

	// Track discovered peers
	peersMap := make(map[string]bool)
	var masterID string

	// Send hello message every second
	go func() {
		helloMsg := HelloMsg{"Hello from " + id, 0}
		for {
			helloMsg.Iter++
			helloTx <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()

	go RouteMessages(&client, networkChannels.receiveChan, networkChannels)
	// Listen for the master
	masterID, found := comm.ListenForMaster(bcastPortString)
	if found {
		// Try to connect to the master
		clientConn, success := comm.ConnectToMaster(masterID, TCPPort)
		if success {
			client.AddClientConnection(id, clientConn, networkChannels.sendChan, networkChannels.receiveChan)
		}
		go ReceiveMessage(networkChannels.receiveChan, clientConn)
		go ClientSendMessages(networkChannels.sendChan, clientConn)
	} else {
		// This whole part should be startMaster() ?
		// No master found, announce ourselves as the master
		masterID = id
		fmt.Printf("Going to announce master. MasterID: %s\n", id)
		go comm.AnnounceMaster(id, bcastPortString)
		go ac.ListenAndAcceptConnections(TCPPort, networkChannels.sendChan, networkChannels.receiveChan)
		go ac.MasterSendMessages(networkChannels.sendChan)
		//go startMaster()
	}

	// Main loop to handle peer updates and hello message reception
	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			// Update the list of discovered peers
			for _, newPeer := range p.New {
				peerID := string(newPeer)
				peersMap[peerID] = true
			}
			for _, lostPeer := range p.Lost {
				delete(peersMap, lostPeer)
			}

			// Once peers are discovered, select the master
			if len(peersMap) > 1 && masterID == "" {
				// Select the master (smallest lexicographically)
				for peerID := range peersMap {
					if masterID == "" || peerID < masterID {
						masterID = peerID
					}
				}
				fmt.Printf("Master selected: %s\n", masterID)

				// If we're not the master, connect to the master using TCP
				if id != masterID {
					go comm.ConnectToMaster(masterID, TCPPort)
				}
			}
		case r := <-networkChannels.receiveChan: // don't really need this case, just for temporary logging
			fmt.Println("Received a message")
			fmt.Printf("Received: %#v\n", r)

		case m := <-networkChannels.MasterChan:
			fmt.Println("Master received a message")
			fmt.Printf("Received: %#v\n", m)
			go masterData.HandleReceivedMessagesToMaster(m)

		case b := <-networkChannels.BackupChan:
			fmt.Println("Got a message from master to backup")
			fmt.Printf("Received: %#v\n", b)

			//data := "I received rank from master"
			var elevst = elevator.ElevStates{Behaviour: "idle", Floor: 0, Direction: "down", CabRequests: []bool{true, false, false, false}, IP: "1.1.1.1"}

			msg:= Message{
				Type: currentStateMessage,
				Target: TargetMaster,
				Payload: elevst,
			}

			SendMessageOnChannel(networkChannels.sendChan, msg)
		
		//Overskriver requests lokalt om man får ny melding fra master
		case e := <-networkChannels.ElevatorChan:
			fsm.El.AssignedRequests = msg.Payload.AssignedRequests
			fsm.El.RequestsToDo = fsm.El.AssignedRequests.append(fsm.El.ElevStates.cabRequests)
			


			// case a := <-helloRx:
			// 	fmt.Printf("Received: %#v\n", a)
		}
	}
}
