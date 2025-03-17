package main

import (
// 	"ElevatorProject/comm"
	//"ElevatorProject/roles"
 	"ElevatorProject/network"
	
// 	"fmt"
// 	"time"
// 
)

func main() {

	// Start network and store connections
	ac := network.CreateActiveConnections()

	var bcastPortInt = 16569
	// var bcastPortString = "16569"
	// For use on same computer?
	var bcastPortString = "9999"
	var peersPort = 15647
	var TCPPort = "8081"

	go network.StartNetwork(ac, bcastPortInt, bcastPortString, peersPort, TCPPort)
	// go network.InitNetwork(ac, bcastPortInt, bcastPortString, peersPort, TCPPort)
	// Start elevator
	//go roles.InitElevator()

	// Actual main program, ish
/*
	localIp := "127.0.0.1"

	// masse greier for å få det til å kjøre på samme pc, sikker helt unødvednig
	var listenPort string
	var broadcastPort string

	//Instance 1
	// broadcastPort = "8081"
	// listenPort = "8082"

	// //Instance 2
	// broadcastPort = "8082"
	// listenPort = "8081"
*/
	/*
	//Instance 3
	broadcastPort = "8081"
	listenPort = "8083"

	masterIP, found := comm.ListenForMaster(listenPort)

	if found {
		rank, conn, success := comm.ConnectToMaster(masterIP, "8081")
		if success {
			if rank == 2 {
				fmt.Println("Going to start backup")
				go roles.StartBackup(conn)
				time.Sleep(5 * time.Second)
			} else if rank > 2 {
				go roles.StartSlave(conn)
				time.Sleep(5 * time.Second)
			}
		}
	} else {
		go comm.AnnounceMaster(localIp, broadcastPort)
		go roles.StartMaster(broadcastPort)
	}
		*/

	// Test, sende til egen PC

	// ac := roles.CreateActiveConnections()
	// port := ":8080"
	// conn_ip := "127.0.0.1"
	// go ac.AddConnection(conn_ip)
	// go comm.Comm_listenAndAccept(conn_ip + port)

	// var elevst = elevator.ElevStates{Behaviour: "idle", Floor: 0, Direction: "down", CabRequests: []bool{true, false, false, false}, IP: conn_ip}

	// var allElevStates = make(map[string]elevator.ElevStates)

	// allElevStates[conn_ip] = elevst
	// ac.ListConnections()

	// assignedRequests := hra.SendStateToHRA(allElevStates, [][2]bool{{false, false}, {false, true}, {false, false}, {true, false}})

	// for k, v := range *assignedRequests {
	// 	fmt.Printf("%s :  %+v\n", k, v)
	// }

	// roles.SendAssignedRequests(assignedRequests, ac)

	// AddConnections test

	// ac := roles.CreateActiveConnections()
	// conn1_ip := "1.2.3.4"
	// conn2_ip := "5.6.7.8"
	// ac.AddConnection(conn1_ip)
	// ac.AddConnection(conn2_ip)

	// ac.ListConnections()

	// ac.RemoveConnection(conn1_ip)
	// ac.ListConnections()

	// // HRA test

	// var elevst1 = elevator.ElevStates{Behaviour: "moving", Floor: 0, Direction: "up", CabRequests: []bool{false, false, false, false}, IP: conn1_ip}
	// var elevst2 = elevator.ElevStates{Behaviour: "idle", Floor: 0, Direction: "down", CabRequests: []bool{true, false, false, false}, IP: conn2_ip}

	// var allElevStates = make(map[string]elevator.ElevStates)

	// allElevStates[conn1_ip] = elevst1
	// allElevStates[conn2_ip] = elevst2

	// assignedRequests := hra.SendStateToHRA(allElevStates, [][2]bool{{false, false}, {false, true}, {false, false}, {false, false}})

	// for k, v := range *assignedRequests {
	// 	fmt.Printf("%6v :  %+v\n", k, v)
	// }

	// 	var conn net.Conn

	// 	if len(os.Args) > 1 && os.Args[1] == "slave" {
	// 		conn = comm.Comm_slaveConnectToMaster()
	// 	} else {
	// 		conn = comm.Comm_masterConnectToSlave()
	// 	}

	// 	fmt.Println("Started!")

	// 	// burde vel ikke måtte definere denne på nytt, er jo definert i elevio?
	// 	numFloors := 4
	// 	const maxDuration time.Duration = 1<<63 - 1

	// 	elevio.Init("localhost:15657", numFloors)

	// 	fsm := fsm.FSM{El: elevator.Elevator_uninitialized(), Od: elevator_io_device.Elevio_getOutputDevice()}

	// 	elevator_logic.ElevLogic_runElevator(fsm, maxDuration, conn)
	//

	select {}
}
