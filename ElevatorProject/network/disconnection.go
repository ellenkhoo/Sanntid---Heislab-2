package network

import (
	"context"
	"fmt"

	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

func handleMasterDisconnection(ac *ActiveConnections, sendChan chan sharedConsts.Message, receiveChan chan sharedConsts.Message,
	 bcastPortString string, peersPort int, TCPPort string, networkChannels sharedConsts.NetworkChannels,
	ctx context.Context, cancel context.CancelFunc) {
	fmt.Println("handling master disconnection...")
	backupConn := findBackupConnection(ac)

	if backupConn != nil {
		fmt.Println("Backup is taking over as master")
		ac.CloseAllConnections()

		cancel()

		go InitMasterSlaveNetwork(ac, Client, Masterdata, bcastPortString, peersPort, TCPPort, networkChannels, fsm)
		go ListenForMasterAndReconnect(ac, bcastPortString, TCPPort, ctx)

	} else {
		fmt.Println("no backup available to taker over as master!")
	}
}

func ListenForMasterAndReconnect(ac *ActiveConnections, bcastPortString string, TCPPort string, ctx context.Context) {
	masterID, found := ListenForMaster(bcastPortString)
	if found {
		clientConn, success := ConnectToMaster(masterID, TCPPort)
		if success {
			sendChan := make(chan sharedConsts.Message)
			receiveChan := make(chan sharedConsts.Message)
			ac.AddClientConnection(masterID, clientConn, sendChan, receiveChan)

			go ReceiveMessage(receiveChan, clientConn, ctx) //go rutine
			go ClientSendMessages(sendChan, clientConn, ctx)
		} else {
			fmt.Println("error connectiong to master. try again...")
		}
	} else {
		fmt.Println("error connection to new master")
	}
}

func (ac *ActiveConnections) CloseAllConnections() {
	fmt.Println("closing all active connections...")

	for _, conn := range ac.Conns {
		if conn.HostConn != nil {
			fmt.Println("closing connection to %s\n", conn.HostConn.RemoteAddr().String())
			conn.HostConn.Close()
		}
	}
	ac.Conns = nil
}

func (ac *ActiveConnections) HandleClientDisconnection(clientip string) {
	fmt.Println("handling client disconnection for client:")
	index := -1
	var isBackup bool
	for i, conn := range ac.Conns {
		if conn.ClientIP == clientip {
			index = i
			fmt.Println("Removing connection to", clientip)

			if conn.HostConn != nil {
				conn.HostConn.Close()
				fmt.Println("closed hostconn for master:", conn.ClientIP)
			}

			if conn.Rank == 2 {
				isBackup = true
			}
			// if conn.ClientConn != nil {
			// 	conn.ClientConn.Close()
			// 	fmt.Println("closed clientconn for client:", conn.ClientIP)
			// }
			break
		}
	}

	if index == -1 {
		fmt.Println("Ip not found in active connections:", clientip)
		return
	}

	ac.Conns = append(ac.Conns[:index], ac.Conns[index+1:]...)

	for i := range ac.Conns {
		ac.Conns[i].Rank = i + 1
	}

	if len(ac.Conns) > 1 {
		if isBackup {
			fmt.Println("choosing new backup...")
			ac.Conns[1].Rank = 2
			fmt.Println("new backup is:", ac.Conns[1].ClientIP)
		} else {
			fmt.Println("Updated ranks for other clients")
		}
	} else {
		fmt.Println("no available backup!")
	}

}