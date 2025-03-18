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

func RouteMessages(receiveChan chan Message, networkChannels NetworkChannels) {
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
			// messages that all types of clients should receive
			networkChannels.ElevatorChan <- msg
			networkChannels.BackupChan <- msg
		default:
			fmt.Println("Unknown message target")
		}
	}
}

func StartNetwork(ac *ActiveConnections, bcastPortInt int, bcastPortString string, peersPort int, TCPPort string) NetworkChannels {
	networkChannels := NetworkChannels{
		MasterChan:   make(chan Message),
		BackupChan:   make(chan Message),
		ElevatorChan: make(chan Message),
	}

	go InitMasterSlaveNetwork(ac, bcastPortInt, bcastPortString, peersPort, TCPPort, networkChannels)

	return networkChannels
}

func InitMasterSlaveNetwork(ac *ActiveConnections, bcastPortInt int, bcastPortString string, peersPort int, TCPPort string, networkChannels NetworkChannels) {
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

	sendChan := make(chan Message)
	receiveChan := make(chan Message)
	go RouteMessages(receiveChan, networkChannels)

	// Listen for the master
	masterID, found := comm.ListenForMaster(bcastPortString)
	if found {
		// Try to connect to the master
		clientConn, success := comm.ConnectToMaster(masterID, TCPPort)
		if success {
			ac.AddClientConnection(clientConn, sendChan, receiveChan)
		}
		go ReceiveMessage(receiveChan, clientConn)
		go ClientSendMessages(sendChan, clientConn)
	} else {
		// No master found, announce ourselves as the master
		masterID = id
		fmt.Printf("Going to announce master. MasterID: %s\n", id)
		go comm.AnnounceMaster(id, bcastPortString)
		go ac.ListenAndAcceptConnections(TCPPort, sendChan, receiveChan)
		go ac.MasterSendMessages(sendChan)
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
		case r := <-receiveChan:
			fmt.Println("Got a message from master")
			fmt.Printf("Received: %#v\n", r)

		case m := <-networkChannels.MasterChan:
			fmt.Println("Got a message to master")
			go HandleReceivedMessagesToMaster(m)

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

			SendMessageOnChannel(sendChan, msg)

			// case a := <-helloRx:
			// 	fmt.Printf("Received: %#v\n", a)
		}
	}
}
