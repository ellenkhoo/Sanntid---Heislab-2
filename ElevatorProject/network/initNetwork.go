package networkUDP

import (
	"ElevatorProject/network/network_functions/bcast"
	"ElevatorProject/network/network_functions/localip"
	"ElevatorProject/network/network_functions/peers"
	"flag"
	"fmt"
	"os"
	"time"
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

	// We start the necessary channels for broadcasting, listening, and peer updates
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

	// Listen for the master (using your provided comm.ListenForMaster)
	masterID, found := comm.ListenForMaster(bcastPortString)
	if found {
		// Try to connect to the master
		rank, conn, success := comm.ConnectToMaster(masterID, TCPPort)
		if success {
			// Based on rank, decide whether to become a backup or slave
			if rank == 2 {
				fmt.Println("Going to start backup")
				go roles.StartBackup(conn)
				time.Sleep(5 * time.Second)
			} else if rank > 2 {
				go roles.StartSlave(conn)
				time.Sleep(5 * time.Second)
			}
		}
	} else {
		// If no master found, announce ourselves as the master
		go comm.AnnounceMaster(id, bcastPortString)
		go roles.StartMaster(TCPPort)
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

// Fra chat
// // Function to connect to the master using TCP
// func connectToMaster(masterID string) {
// 	fmt.Printf("Connecting to master %s...\n", masterID)

// 	// TCP connection setup (assuming master is on port 8081)
// 	conn, err := net.Dial("tcp", masterID+":8081")
// 	if err != nil {
// 		fmt.Println("Failed to connect to master:", err)
// 		return
// 	}
// 	defer conn.Close()

// 	fmt.Printf("Connected to master %s\n", masterID)

// 	// Now you can perform communication over TCP (optional)
// }
