package main

import (
	"ElevatorProject/comm"
	"ElevatorProject/elevator"
	"ElevatorProject/hra"
	"ElevatorProject/roles"
	"fmt"
)

func main() {

	// Actual main program, ish

	masterIP, found := comm.ListenForMaster()

	if found {
		rank, conn, success := comm.ConnectToMaster(masterIP) // bør man returnere conn også og sende den inn i backup/slave?
		if success {
			if rank == 2 {
				go roles.StartBackup(conn)
			} else if rank > 2 {
				go roles.StartSlave(conn)
			}
		}
	} else {
		go comm.AnnounceMaster()
		roles.StartMaster()
	}

	// Test, sende til egen PC

	ac := roles.CreateActiveConnections()
	port := ":8080"
	conn_ip := "127.0.0.1"
	go ac.AddConnection(conn_ip)
	go comm.Comm_listenAndAccept(conn_ip + port)

	var elevst = elevator.ElevStates{Behaviour: "idle", Floor: 0, Direction: "down", CabRequests: []bool{true, false, false, false}, IP: conn_ip}

	var allElevStates = make(map[string]elevator.ElevStates)

	allElevStates[conn_ip] = elevst
	ac.ListConnections()

	assignedRequests := hra.SendStateToHRA(allElevStates, [][2]bool{{false, false}, {false, true}, {false, false}, {true, false}})

	for k, v := range *assignedRequests {
		fmt.Printf("%s :  %+v\n", k, v)
	}

	roles.SendAssignedRequests(assignedRequests, ac)

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
}
