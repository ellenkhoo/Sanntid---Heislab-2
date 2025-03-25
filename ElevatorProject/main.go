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
		UpdateChan:   make(chan string),
	}

	ac := network.CreateActiveConnections()
	client := network.ClientConnectionInfo{}
	client.Channels = *networkChannels
	masterData := network.CreateMasterData()
	backupData := network.CreateBackupData()

	localIP, _ := localip.LocalIP()

	fsm := elevator.InitElevator(localIP, &client.Channels)
	go network.InitMasterSlaveNetwork(ac, &client, masterData, backupData, network.BcastPort, network.TCPPort, networkChannels, fsm)

	select {}
}
