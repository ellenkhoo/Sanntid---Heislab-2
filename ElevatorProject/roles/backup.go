package roles

import (
	"fmt"
	"net"
)



func StartBackup(rank int, conn net.Conn) {
	fmt.Println("Starting backup")

	// var allElevStates = make(map[string]elevator.ElevStates)
	// var globalHallRequests [][2]bool	

	InitElevator(rank, conn)
	
}
