package roles

import (
	"ElevatorProject/comm"
	"fmt"
	"net"
	"sync"
)

//master-init()?
// ac := roles.CreateActiveConnections()
// var allElevStates = make(map[string]elevator.ElevStates)
// var globalHallRequests [][2]bool

// Keeping track of connections
type Connection struct {
	IP   string
	Rank int
	Conn net.Conn
}
type ActiveConnections struct {
	mu    sync.Mutex
	conns []Connection
}

func CreateActiveConnections() *ActiveConnections {
	return &ActiveConnections{}
}

// Replaces AddConnection ?
func (ac *ActiveConnections) handleConnection(conn net.Conn) {
	//defer conn.Close()
	remoteIP, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Check if IP is already added
	// for _, conn := range ac.conns {
	// 	if conn.IP == remoteIP {
	// 		return
	// 	}
	// }

	// Assign the next availiable rank
	// rank := len(ac.conns) + 2 // burde man legge til master ved oppstart? mulighvis man skal bruke rank til noe mer enn sjekk i main
	// ac.conns = append(ac.conns, Connection{IP: remoteIP, Rank: rank, Conn: conn})

	rank := len(ac.conns) + 2 // burde man legge til master ved oppstart? mulighvis man skal bruke rank til noe mer enn sjekk i main
	ac.conns = append(ac.conns, Connection{IP: remoteIP, Rank: rank, Conn: conn})

	conn.Write([]byte(fmt.Sprintf("You have rank %d\n", rank)))
}

func (ac *ActiveConnections) AddConnection(ip string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Check if IP is already added
	for _, conn := range ac.conns {
		if conn.IP == ip {
			return
		}
	}

	// Establish the connection
	conn, err := net.Dial("tcp", ip+":8080")
	if err != nil {
		fmt.Println("Failed to connect to", ip, ":", err)
		return
	}

	// Assign the next availiable rank
	rank := len(ac.conns) + 1
	ac.conns = append(ac.conns, Connection{IP: ip, Rank: rank, Conn: conn})
}

func (ac *ActiveConnections) RemoveConnection(ip string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Find index of IP to be removed
	index := -1
	for i, conn := range ac.conns {
		if conn.IP == ip {
			index = i
			conn.Conn.Close() // Close the connection before removal
			break
		}
	}

	// IP not found
	if index == -1 {
		return
	}

	// IP found, remove from list and adjust the ranks
	ac.conns = append(ac.conns[:index], ac.conns[index+1:]...)
	for i := range ac.conns {
		ac.conns[i].Rank = i + 1
	}

	fmt.Println("Successfully removed connection to", ip)
}

func (ac *ActiveConnections) ListConnections() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	for _, conn := range ac.conns {
		fmt.Printf("IP: %s, Rank: %d\n", conn.IP, conn.Rank)
	}
}

// func SendAssignedRequests(assignedRequests *map[string][][2]bool, activeConnections *ActiveConnections) {
// 	activeConnections.mu.Lock()
// 	defer activeConnections.mu.Unlock()

// 	for _, connInfo := range activeConnections.conns {
// 		requests, exists := (*assignedRequests)[connInfo.IP]
// 		fmt.Printf("Checking requests for IP: '%s'\n", connInfo.IP)

// 		if !exists {
// 			fmt.Printf("No requests found for %s", connInfo.IP)
// 			continuebroadcastPort = "8081"
// 		} else {
// 			fmt.Println("Successfully sent data to ", connInfo.IP)
// 		}
// 	}
// }


// NEW
func ListenForConnections(port string) {
	ac := CreateActiveConnections()
	ln, _ := net.Listen("tcp", ":"+port)

	for {
		conn, _ := ln.Accept()
		go ac.handleConnection(conn)
	}
}

func StartMaster(rank int, port string, conn net.Conn) {

	fmt.Println("Starting master")

	go InitElevator(rank, conn)

	// masterReceiveOrder is going to receive data from multiple connections, so conn can't be an argument as now
	go comm.Comm_masterReceive(conn)

	fmt.Println("Master is now ready to receive orders")

	select{}
}
