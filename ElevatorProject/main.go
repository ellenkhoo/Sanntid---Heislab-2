package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ellenkhoo/ElevatorProject/elevator"
	"github.com/ellenkhoo/ElevatorProject/network"
	"github.com/ellenkhoo/ElevatorProject/network/network_functions/localip"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

func main() {

	// Initialize network channels
	networkChannels := &sharedConsts.NetworkChannels{
		SendChan:     make(chan sharedConsts.Message, 100),
		ReceiveChan:  make(chan sharedConsts.Message),
		MasterChan:   make(chan sharedConsts.Message),
		BackupChan:   make(chan sharedConsts.Message),
		ElevatorChan: make(chan sharedConsts.Message, 100),
		RestartChan: make(chan string),
		UpdateChan:   make(chan string),
	}

	ac := network.CreateActiveConnections()
	client := network.ClientConnectionInfo{}
	client.Channels = *networkChannels
	masterData := network.CreateMasterData()

	go network.RouteMessages(&client, networkChannels)

	var id string
	var num int

	// Define command-line flags
	flag.StringVar(&id, "id", "", "ID of this peer")
	flag.IntVar(&num, "num", 0, "Custom number for the peer ID")
	flag.Parse()

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}

		// Use the provided number if given, otherwise fallback to process ID
		if num > 0 {
			id = fmt.Sprintf("peer-%s-%d", localIP, num)
		} else {
			id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
		}
	}

	fmt.Println("Assigned ID:", id)

	fsm := elevator.InitElevator(id, &client.Channels)
	go network.InitNetwork(id, ac, &client, masterData, network.BcastPort, network.TCPPort, networkChannels, fsm)

	select {}
}
