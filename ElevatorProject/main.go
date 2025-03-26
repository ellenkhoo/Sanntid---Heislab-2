package main

import (
	"github.com/ellenkhoo/ElevatorProject/elevator"
	"github.com/ellenkhoo/ElevatorProject/network"
	"github.com/ellenkhoo/ElevatorProject/network/network_functions/localip"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

func main() {

	// Initialize network channels
	networkChannels := &sharedConsts.NetworkChannels{
		SendChan:     make(chan sharedConsts.Message),
		ReceiveChan:  make(chan sharedConsts.Message),
		MasterChan:   make(chan sharedConsts.Message),
		BackupChan:   make(chan sharedConsts.Message),
		ElevatorChan: make(chan sharedConsts.Message),
		RestartChan:  make(chan string),
		UpdateChan:   make(chan string),
	}

	ac := network.CreateActiveConnections()
	client := network.ClientConnectionInfo{}
	client.Channels = *networkChannels
	masterData := network.CreateMasterData()

	id, _ := localip.LocalIP()

	fsm := elevator.InitElevator(id, &client.Channels)
	go network.RouteMessages(&client, networkChannels)

	go network.InitNetwork(id, ac, &client, masterData, network.BcastPort, network.TCPPort, networkChannels, fsm)

	select {}
}
