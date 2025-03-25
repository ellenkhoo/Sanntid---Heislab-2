package network

import (
	//"github.com/ellenkhoo/ElevatorProject/network/network_functions/bcast"
	"github.com/ellenkhoo/ElevatorProject/network/network_functions/localip"
	//"github.com/ellenkhoo/ElevatorProject/network/network_functions/peers"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"context"
	"github.com/ellenkhoo/ElevatorProject/elevator"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	//"time"
)

func CreateActiveConnections() *ActiveConnections {
	return &ActiveConnections{}
}

func CreateMasterData() *MasterData {
	return &MasterData{
		GlobalHallRequests:  [elevator.N_FLOORS][2]bool{},
		AllAssignedRequests: make(map[string][elevator.N_FLOORS][2]bool),
		AllElevStates:       make(map[string]elevator.ElevStates),
	}
}

func CreateBackupData() *BackupData {
	return &BackupData{
		GlobalHallRequests:  [elevator.N_FLOORS][2]bool{},
		AllAssignedRequests: make(map[string][elevator.N_FLOORS][2]bool),
	}
}
func SendMessage(client *ClientConnectionInfo, msg sharedConsts.Message, conn net.Conn) {
	fmt.Println("At SendMessage")
	if client.ID == client.HostIP {
		client.Channels.ReceiveChan <- msg
	}
	fmt.Println("The message is to a remote client")
	encoder := json.NewEncoder(conn)
	err := encoder.Encode(msg)
	if err != nil {
		fmt.Println("Error encoding message: ", err)
		return
	}
}

func ReceiveMessage(ctx context.Context, receiveChan chan sharedConsts.Message, conn net.Conn) {
	fmt.Println("At func ReceiveMessage!")
	decoder := json.NewDecoder(conn)

	for {
		select {
		case <- ctx.Done():
			fmt.Printf("Shutting down ReceiveMessage")
			return
		default:
			var msg sharedConsts.Message
			err := decoder.Decode(&msg)
			if err != nil {
				if err == io.EOF {
					fmt.Println("Connection closed")
				}
				fmt.Println("Error decoding message: ", err)
				continue // eller continue?
			}
			receiveChan <- msg
		}
	}
}

func RouteMessages(ctx context.Context, client *ClientConnectionInfo, networkChannels *sharedConsts.NetworkChannels) {
	fmt.Println("Router received msg")
	for msg := range networkChannels.ReceiveChan {
		select {
		case <- ctx.Done():
			fmt.Printf("Shutting down RouteMessage")
			return
		default:
			switch msg.Target {
			case sharedConsts.TargetMaster:
				fmt.Println("Msg is to master")
				networkChannels.MasterChan <- msg
			case sharedConsts.TargetClient:
				fmt.Println("Msg is to client")
				client.HandleReceivedMessageToClient(msg)
			case sharedConsts.TargetElevator:
				fmt.Println("Msg is to elevator")
				networkChannels.ElevatorChan <- msg
			default:
				fmt.Println("Unknown message target")
			}
		}
	}
}

func InitMasterSlaveNetwork(ctx context.Context, ac *ActiveConnections, client *ClientConnectionInfo, masterData *MasterData, backupData *BackupData, ackTracker *AcknowledgeTracker, bcastPort string, TCPPort string, networkChannels *sharedConsts.NetworkChannels, fsm *elevator.FSM) {
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

	go RouteMessages(ctx, client, networkChannels)
	// Listen for the master
	var masterID string = ""
	masterID, found := ListenForMaster(bcastPort)
	if found {
		//Try to connect to the master
		clientConn, success := ConnectToMaster(masterID, TCPPort)
		if success {
			client.AddClientConnection(id, clientConn, networkChannels)
			go ReceiveMessage(ctx, networkChannels.ReceiveChan, clientConn)
			go ClientSendMessagesFromSendChan(ctx, client, networkChannels.SendChan, clientConn)
		}
	} else {
		// No master found, announce ourselves as the master
		masterID = id
		client.ID = id // local client (elevator)
		client.HostIP = masterID
		fmt.Printf("Going to announce master. MasterID: %s\n", id)
		go AnnounceMaster(ctx, id, bcastPort)
		go ac.ListenAndAcceptConnections(ctx, TCPPort, networkChannels)
		go ac.MasterSendMessages(ctx, client)
	}

	for {
		select {
		case <- ctx.Done():
			fmt.Printf("Shutting down InitMasterSlaveNetwork")
			return
		default:
			select {
			case m := <-networkChannels.MasterChan:
				fmt.Println("Master received a message")
				go masterData.HandleReceivedMessagesToMaster(ctx, ac, m, client, ackTracker)
			case e := <-networkChannels.ElevatorChan:
				fmt.Println("Going to update my worldview")
				client.UpdateElevatorWorldview(fsm, e) //Tidligere blitt kjørt som go-routine. Gjør dte igjen om det ikke fungerer lenger
			}
		}
	}
}
