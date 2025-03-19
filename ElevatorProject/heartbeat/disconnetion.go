package heartbeat

import (
	"fmt"

	"github.com/ellenkhoo/ElevatorProject/network"
)

func handleMasterDisconnection(ac *network.ActiveConnections) {
	backupConn := findBackupConnection(ac)
	if backupConn == nil {
		fmt.Println("No backup found")
		return
	}

	//finner slave med høyest rang
	var newBackup *network.ClientConnectionInfo
	var maxRank int

	for _, conn := range ac.Conns {
		if conn.Rank > 2 {
			clientConnInfo, ok := conn.(*network.MasterConnectionInfo)
			if ok && clientConnInfo.Rank > maxRank {
				maxRank = clientConnInfo.Rank
				newBackup = clientConnInfo
			}
		}
	}

	if newBackup == nil {
		fmt.Println("no slave found to promote to backup")
		return
	}

	backupConn.Rank = 1
	fmt.Println("backup is now promoted to master")

	newBackup.Rank = 2
	fmt.Println("slave is now promoted to backup")

	go sendRoleUpdateToClients(ac, backupConn, newBackup)
}

func sendRoleUpdateToClients(ac *network.ActiveConnections, newMaster *network.ClientConnectionInfo, newBackup *network.ClientConnectionInfo) {
	message := network.Message{
		Type:    network.rankMessage,
		Target:  network.TargetClient,
		Payload: fmt.Sprintf("New master: %s, New backup: %s", newMaster.ID, newBackup.ID),
	}

	for _, conn := range ac.Conns {
		if conn.Rank != 1 {
			network.ClientSendMessages(conn.SendChan, conn.HostConn)
		}
	}

}
