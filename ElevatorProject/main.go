package main

import (
	"time"
	"context"
	"fmt"
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

	// Start network and store connections
	// dumt at ac lages her? da vil alle pc-ene ha det
	ac := network.CreateActiveConnections()
	client := network.ClientConnectionInfo{}
	client.Channels = *networkChannels
	masterData := network.CreateMasterData()
	backupData := network.CreateBackupData()
	ackTracker := network.NewAcknowledgeTracker(client.Channels.SendChan, 1*time.Second)

	localIP, _ := localip.LocalIP()

	fsm := elevator.InitElevator(localIP, &client.Channels)

	ctx, cancel := context.WithCancel(context.Background())

	go network.InitMasterSlaveNetwork(ctx, ac, &client, masterData, backupData, ackTracker, network.BcastPort, network.TCPPort, networkChannels, fsm)

	time.Sleep(5*time.Second)

	fmt.Printf("Trying to cancel all go routines")
	cancel()

	select{}
}
