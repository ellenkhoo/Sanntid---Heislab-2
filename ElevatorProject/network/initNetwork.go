package networkUDP

import (
	"ElevatorProject/network/network_functions/bcast"
	"ElevatorProject/network/network_functions/localip"
	"ElevatorProject/network/network_functions/peers"
	"flag"
	"fmt"
	"os"
	"time"
	"net"
	"ElevatorProject/comm"
	"ElevatorProject/roles"
)

type HelloMsg struct {
	Message string
	Iter    int
}

var bcastPortInt = 16569
// var bcastPortString = "16569"
// For use on same computer?
var bcastPortString = "9999"
var peersPort = 15647
var TCPPort = "8081"


func InitNetwork() {
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

	// Listen for the master 
	masterID, found := comm.ListenForMaster(bcastPortString)
	if found {
		// Try to connect to the master
		rank, conn, success := comm.ConnectToMaster(masterID, TCPPort)
		if success {
			// Based on rank, decide whether to become a backup or slave
			if rank == 2 {
				fmt.Println("Going to start backup")
				go roles.StartBackup(rank, conn)
				time.Sleep(5 * time.Second)
			} else if rank > 2 {
				go roles.StartSlave(rank, conn)
				time.Sleep(5 * time.Second)
			}
		}
	} else {
		// No master found, announce ourselves as the master
		go comm.AnnounceMaster(id, bcastPortString)
		// Connects to itself so that the elevator can communicate with it. Not sure if it works
		go roles.ListenForConnections(TCPPort)
		time.Sleep(3 * time.Second)
		rank := 1
		localIP := "127.0.0.1"
		conn, err := net.Dial("tcp", localIP+":"+TCPPort)
		if err != nil {
			fmt.Println("Master failed to connect to itself", err)
		}

		go roles.StartMaster(rank, TCPPort, conn)
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

