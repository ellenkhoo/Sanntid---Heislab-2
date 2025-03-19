package network

import (
	"github.com/ellenkhoo/ElevatorProject/network/network_functions/bcast"
	"github.com/ellenkhoo/ElevatorProject/network/network_functions/localip"
	"github.com/ellenkhoo/ElevatorProject/network/network_functions/peers"
	"github.com/ellenkhoo/ElevatorProject/elevator"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"time"
	"context"
)

func CreateActiveConnections() *ActiveConnections {
	return &ActiveConnections{}
}


func ReceiveMessage(receiveChan chan sharedConsts.Message, conn net.Conn, ctx context.Context) {
	fmt.Println("At func ReceiveMessage!")
	decoder := json.NewDecoder(conn)
	
	var msg sharedConsts.Message 
	for {
		select{
		case <- ctx.Done():
			return
		default: 
			err := decoder.Decode(&msg)
			if err != nil {
				fmt.Println("Error decoding message: ", err)
				return
			}
		}
		receiveChan <-msg
	}
}

func SendMessageOnChannel(sendChan chan sharedConsts.Message, msg sharedConsts.Message) {
	fmt.Println("Sending msg on chan")
	sendChan <- msg
}

func RouteMessages(client *(ClientConnectionInfo), receiveChan chan sharedConsts.Message, networkChannels sharedConsts.NetworkChannels, ctx context.Context) {
	for {
		select{
		case msg := <- receiveChan:
				switch msg.Target {
				case sharedConsts.TargetMaster:
					networkChannels.MasterChan <- msg
				case sharedConsts.TargetBackup:
					fmt.Print("Sending msg on backup chan")
					networkChannels.BackupChan <- msg
				case sharedConsts.TargetElevator:
					networkChannels.ElevatorChan <- msg
				case sharedConsts.TargetClient:
					// Messages that all clients should receive
					client.HandleReceivedMessageToClient(msg)
					
				default:
					fmt.Println("Unknown message target")
				}
			case <- ctx.Done():
				return
		}
	}
}


func InitMasterSlaveNetwork(ac *ActiveConnections, client ClientConnectionInfo, masterData MasterData, bcastPortInt int, bcastPortString string, peersPort int, TCPPort string, networkChannels sharedConsts.NetworkChannels, fsm elevator.FSM) {
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
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

	helloTx := make(chan sharedConsts.HelloMsg)
	helloRx := make(chan sharedConsts.HelloMsg)
	go bcast.Transmitter(bcastPortInt, helloTx)
	go bcast.Receiver(bcastPortInt, helloRx)

	// Track discovered peers
	peersMap := make(map[string]bool)
	var masterID string

	// Send hello message every second
	go func() {
		helloMsg := sharedConsts.HelloMsg{"Hello from " + id, 0}
		for {
			select {
			case <- ctx.Done():
				return
			default:
				helloMsg.Iter++
				helloTx <- helloMsg
				time.Sleep(1 * time.Second)
			}
		}
	}()

	go RouteMessages(&client, networkChannels.ReceiveChan, networkChannels, ctx)
	// Listen for the master
	masterID, found := ListenForMaster(bcastPortString)
	if found {
		// Try to connect to the master
		clientConn, success := ConnectToMaster(masterID, TCPPort)
		if success {
			client.AddClientConnection(id, clientConn, networkChannels.SendChan, networkChannels.ReceiveChan)
		}
		go ReceiveMessage(networkChannels.ReceiveChan, clientConn, ctx)
		go ClientSendMessages(networkChannels.SendChan, clientConn, ctx)
	} else {
		// This whole part should be startMaster() ?
		// No master found, announce ourselves as the master
		masterID = id
		fmt.Printf("Going to announce master. MasterID: %s\n", id)
		go AnnounceMaster(id, bcastPortString, ctx)
		go ac.ListenAndAcceptConnections(TCPPort, networkChannels.SendChan, networkChannels.ReceiveChan, ctx)
		go ac.MasterSendMessages(networkChannels.SendChan, ctx)
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
					go ConnectToMaster(masterID, TCPPort)
				}
			}
		case r := <-networkChannels.ReceiveChan: // don't really need this case, just for temporary logging
			fmt.Println("Received a message")
			fmt.Printf("Received: %#v\n", r)

		case m := <-networkChannels.MasterChan:
			fmt.Println("Master received a message")
			fmt.Printf("Received: %#v\n", m)
			masterData.HandleReceivedMessagesToMaster(m)

		case b := <-networkChannels.BackupChan:
			fmt.Println("Got a message from master to backup")
			fmt.Printf("Received: %#v\n", b)
			// msg:= Message{
			// 	Type: currentStateMessage,
			// 	Target: TargetMaster,
			// 	Payload: elevst,
			// }

			// SendMessageOnChannel(networkChannels.sendChan, msg)
		
		//Overskriver requests lokalt om man får ny melding fra master
		case e := <-networkChannels.ElevatorChan:
			fmt.Printf("Received: %#v\n", e)
			//HandleReceivedMessageToElevator(e)


			// case a := <-helloRx:
			// 	fmt.Printf("Received: %#v\n", a)
		}
	}
}
