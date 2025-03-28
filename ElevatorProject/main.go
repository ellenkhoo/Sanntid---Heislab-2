package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ellenkhoo/ElevatorProject/elevator"
	"github.com/ellenkhoo/ElevatorProject/network"
	"github.com/ellenkhoo/ElevatorProject/network/networkResources/localip"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

func main() {

	// Define client's ID
	var ID string
	var num int

	flag.StringVar(&ID, "id", "", "ID of this peer")
	flag.IntVar(&num, "num", 0, "Custom number for the peer ID")
	flag.Parse()

	if ID == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		if num > 0 {
			ID = fmt.Sprintf("peer-%s-%d", localIP, num)
		} else {
			ID = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
		}
	}

	// Initialize the system
	networkChannels := &sharedConsts.NetworkChannels{
		SendChan:     make(chan sharedConsts.Message, 100),
		ReceiveChan:  make(chan sharedConsts.Message),
		MasterChan:   make(chan sharedConsts.Message),
		BackupChan:   make(chan sharedConsts.Message),
		ElevatorChan: make(chan sharedConsts.Message, 100),
		RestartChan:  make(chan string),
		UpdateChan:   make(chan string),
	}

	client := network.ClientInfo{}
	client.Channels = *networkChannels
	activeConnections := network.CreateActiveConnections()
	masterData := network.CreateMasterData()

	fsm := elevator.InitElevator(ID, &client.Channels)
	go network.RouteMessagesToCorrectChannel(&client, networkChannels)
	go network.InitNetwork(ID, activeConnections, &client, masterData, network.BcastPort, network.TCPPort, networkChannels, fsm)

	select {}
}
