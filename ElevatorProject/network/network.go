package network

import (
	"ElevatorProject/comm"
	"ElevatorProject/network/network_functions/bcast"
	"ElevatorProject/network/network_functions/localip"
	"ElevatorProject/network/network_functions/peers"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

type HelloMsg struct {
	Message string
	Iter    int
}

type MessageType int

const (
	globalHallRequestMessage MessageType = iota
	assignedHallRequestsMessage
	backupAcknowledgeMessage
	localRequestMessage
	currentStateMessage
)

type MessageTarget int

const (
	TargetMaster MessageTarget = iota
	TargetBackup
	TargetElevator
)

type Message struct {
	Type    MessageType
	Target  MessageTarget
	Payload interface{}
}

// Keeping track of connections
type Connection struct {
	IP          string
	Rank        int
	HostConn    net.Conn
	ClientConn  net.Conn
	SendChan    chan Message
	ReceiveChan chan Message
}

type NetworkChannels struct {
	MasterChan   chan Message
	BackupChan   chan Message
	ElevatorChan chan Message
}

type ActiveConnections struct {
	// unødvendig med mutex?
	//mu    sync.Mutex
	conns []Connection
}

func CreateActiveConnections() *ActiveConnections {
	return &ActiveConnections{}
}

// When a new connection is established on the client side, this function adds it to the list of active connections
func (ac *ActiveConnections) AddClientConnection(conn net.Conn, sendChan chan Message, receiveChan chan Message) {
	//defer conn.Close()
	remoteIP, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	// ac.mu.Lock()
	// defer ac.mu.Unlock()

	//Check if IP is already added
	for _, connections := range ac.conns {
		if connections.IP == remoteIP {
			return
		}
	}

	// sendChan := make(chan Message)
	// receiveChan := make(chan Message)

	newConn := Connection{
		IP:          remoteIP,
		Rank:        len(ac.conns) + 1,
		ClientConn:  conn,
		SendChan:    sendChan,
		ReceiveChan: receiveChan,
	}

	ac.conns = append(ac.conns, newConn)

	go HandleConnection(newConn)
}

// Maybe not the most describing name
func HandleConnection(conn Connection) {
	// Read from TCP connection and send to the receive channel
	go func() {
		decoder := json.NewDecoder(conn.ClientConn)
		for {
			var msg Message
			err := decoder.Decode(&msg)
			if err != nil {
				fmt.Println("Error decoding message: ", err)
				return
			}
			conn.ReceiveChan <- msg
		}
	}()

	// Read from the send channel and write to the TCP connection
	go func() {
		encoder := json.NewEncoder(conn.ClientConn)
		for msg := range conn.SendChan {
			err := encoder.Encode(msg)
			if err != nil {
				fmt.Println("Error encoding message: ", err)
				return
			}
		}
	}()
}

// Adds the host's connection with the relevant client in the list of active connections
func (ac *ActiveConnections) AddHostConnection(conn net.Conn) {
	remoteIP, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	// Check if a connection already exists with this IP
	for i, connection := range ac.conns {
		if connection.IP == remoteIP {
			ac.conns[i].HostConn = conn
			return
		}
	}
}

// Removes a connection from the list of active connections when connection is lost
func (ac *ActiveConnections) RemoveConnection(ip string) {
	// ac.mu.Lock()
	// defer ac.mu.Unlock()

	// Find index of IP to be removed
	index := -1
	for i, conn := range ac.conns {
		if conn.IP == ip {
			index = i
			// Close the connection before removal
			conn.HostConn.Close()
			conn.ClientConn.Close()
			break
		}
	}

	// IP not found
	if index == -1 {
		return
	}

	// IP found, remove from list and adjust the ranks
	ac.conns = append(ac.conns[:index], ac.conns[index+1:]...)
	for i := range ac.conns {
		ac.conns[i].Rank = i + 1
	}

	fmt.Println("Successfully removed connection to", ip)
}

func (ac *ActiveConnections) ListConnections() {
	// ac.mu.Lock()
	// defer ac.mu.Unlock()

	for _, conn := range ac.conns {
		fmt.Printf("IP: %s, Rank: %d\n", conn.IP, conn.Rank)
	}
}

// Master listenes and accepts connections
func ListenAndAcceptConnections(ac ActiveConnections, port string) {
	ln, _ := net.Listen("tcp", ":"+port)

	for {
		hostConn, _ := ln.Accept()
		go ac.AddHostConnection(hostConn)
	}
}

func ReceiveandDistributeMessages(receiveChan chan Message, conn net.Conn, elevatorChan chan Message, masterChan chan Message, backupChan chan Message) {
	decoder := json.NewDecoder(conn)

	for {
		var msg Message
		err := decoder.Decode(&msg)
		if err != nil {
			fmt.Println("Error decoding message: ", err)
			return
		}

		switch msg.Target {
		case TargetMaster:
			masterChan <- msg
		case TargetBackup:
			backupChan <- msg
		case TargetElevator:
			elevatorChan <- msg
		default:
			fmt.Println("Unknown message target")
		}
	}
	// for msg := range receiveChan {
	// 	switch msg.Type {
	// 	case globalHallRequestMessage:
	// 		// do something
	// 	case assignedHallRequestsMessage:
	// 		// do something
	// 	default:
	// 		fmt.Println("Unknown message type")
	// 	}
	// }
}

func sendMessageOnChannel(sendChan chan Message, msg Message) {
	sendChan <- msg
}

func SendMessages(sendChan chan Message, conn net.Conn) {

	encoder := json.NewEncoder(conn)
	for msg := range sendChan {
		err := encoder.Encode(msg)
		if err != nil {
			fmt.Println("Error encoding message: ", err)
			return
		}
		// switch msg.Type {
		// case localRequestMessage:
		// 	// do something
		// case currentStateMessage:
		// 	// do something
		// case backupAcknowledgeMessage:
		// 	// do something
		// default:
		// 	fmt.Println("Unknown message type")
		// }
	}
}

// func hostReceiveMessages(receiveChan chan Message) {
// 	for msg := range receiveChan {
// 		switch msg.Type {
// 		case localRequestMessage:
// 			// do something
// 		case currentStateMessage:
// 			// do something
// 		case backupAcknowledgeMessage:
// 			// do something
// 		default:
// 			fmt.Println("Unknown message type")
// 		}
// 	}
// }

// func hostSendMessages(sendChan chan Message) {
// 	for msg := range sendChan {
// 		switch msg.Type {
// 		case globalHallRequestMessage:
// 			// do something
// 		case assignedHallRequestsMessage:
// 			// do something
// 		default:
// 			fmt.Println("Unknown message type")
// 		}
// 	}
// }

func RouteMessages(receiveChan chan Message, networkChannels NetworkChannels) {
	for msg := range receiveChan {
		switch msg.Target {
		case TargetMaster:
			networkChannels.MasterChan <- msg
		case TargetBackup:
			networkChannels.BackupChan <- msg
		case TargetElevator:
			networkChannels.ElevatorChan <- msg
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
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
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
	} else {
		// No master found, announce ourselves as the master
		id = masterID
		go comm.AnnounceMaster(id, bcastPortString)
		go ListenAndAcceptConnections(*ac, TCPPort)
		// A small delay to allow the master to start listening before trying to connect to itself
		time.Sleep(1 * time.Second)
		localIP := "127.0.0.1"
		clientConn, err := net.Dial("tcp", localIP+":"+TCPPort)
		if err != nil {
			fmt.Println("Master failed to connect to itself", err)
		}
		ac.AddClientConnection(clientConn, sendChan, receiveChan)
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

			// case a := <-helloRx:
			// 	fmt.Printf("Received: %#v\n", a)
		}
	}
}
