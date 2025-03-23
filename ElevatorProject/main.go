package main

import (
	"github.com/ellenkhoo/ElevatorProject/network/network_functions/localip"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	"github.com/ellenkhoo/ElevatorProject/elevator"
	"github.com/ellenkhoo/ElevatorProject/network"
	"time"
)

func main() {

	// Start network and store connections
	// dumt at ac lages her? da vil alle pc-ene ha det
	ac := network.CreateActiveConnections()
	client := network.ClientConnectionInfo{}
	masterData := network.CreateMasterData()
	backupData := network.CreateBackupData()
	ackTracker := network.NewAcknowledgeTracker(5 * time.Second)

	localIP, _ := localip.LocalIP()

	// Initialize network channels
	networkChannels := &sharedConsts.NetworkChannels{
		SendChan:     make(chan sharedConsts.Message),
		ReceiveChan:  make(chan sharedConsts.Message),
		MasterChan:   make(chan sharedConsts.Message),
		BackupChan:   make(chan sharedConsts.Message),
		ElevatorChan: make(chan sharedConsts.Message),
		UpdateChan: make(chan string),
	}

	client.Channels = *networkChannels

	fsm := elevator.InitElevator(localIP, &client.Channels)
	time.Sleep(2 * time.Second)
	go network.InitMasterSlaveNetwork(ac, &client, masterData, backupData, ackTracker, network.BcastPort, network.TCPPort, networkChannels, fsm)
	
	select {}
}
