package elevator_logicpkg


import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
	"Driver-go/elevio"
)

var allElevStates = make(map[int]elevio.ElevStates)
var globalHallRequest [][2]bool
var activeElevatorConnections = make(map[int]net.Conn)
var mutex sync.Mutex


func MasterLogic_StartMasterServer(port string) {
	ln, err := net.Listen("tcp" , port)
	if err != nil {
		fmt.Println("Error starting master: ", err)
		return
	}
	defer ln.Close()

	fmt.Println("Master listening on", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}
		go MasterLogic_handleElevatorConnection(conn)
	}
}



func MasterLogic_handleElevatorConnection(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)

	for {
		var state elevio.ElevStates
		err := decoder.Decode(&state)
		if err != nil {
			fmt.Println("Error decoding elevator state: ", err)
			return
		}
		mutex.Lock()
		allElevStates[state.ID] = state
		mutex.Unlock()

		fmt.Printf("Recieved state from elevator %d: %+v\n", state.ID, state)
	}
}


func MasterLogic_runHRAUpdater() {
	for {
		time.Sleep(1*time.Second)

		mutex.Lock()
		hallAssignments := SendStateToHRA(allElevStates, globalHallRequest)
		mutex.Unlock()

		if hallAssignments != nil {
			fmt.Println("New hall assignments: ", hallAssignments)

			//sending back to elevators
			for id, hallrequest := range *hallAssignments {
				elevatorID, err := strconv.Atoi(id)
				if err != nil {
					fmt.Println("Error parsing elevator ID: ", err)
					continue
				}

				conn, exists := activeElevatorConnections[elevatorID]
				if !exists {
					fmt.Printf("No connection found for elevator %d\n", elevatorID)
					continue
				}

				Comm_sendMessage("hall_request", hallrequest, conn)
			}
		}
	}
}

func MasterLogic_sendUpdatedRequestToElevators(output *map[string][][2]bool, connections map[int]net.Conn) {

}