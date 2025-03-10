package roles

import (
	"encoding/json"
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

func (ac *ActiveConnections) AddConnection(ip string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Check if ip is already added
	for _, conn := range ac.conns {
		if conn.IP == ip {
			return
		}
	}

	// Må ha noe på denne formen?
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

	// Find index of ip to be removed
	index := -1
	for i, conn := range ac.conns {
		if conn.IP == ip {
			index = i
			// conn.Conn.Close() // Close the connection before removal
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

func SendAssignedRequests(assignedRequests *map[string][][2]bool, activeConnections *ActiveConnections) {
	activeConnections.mu.Lock()
	defer activeConnections.mu.Unlock()

	for _, connInfo := range activeConnections.conns {
		requests, exists := (*assignedRequests)[connInfo.IP]
		fmt.Printf("Checking requests for IP: '%s'\n", connInfo.IP)
		for k := range *assignedRequests {
			fmt.Printf("Available key in assignedRequests: '%s'\n", k)
		}

		if !exists {
			fmt.Printf("No requests found for %s", connInfo.IP)
			continue
		}

		data, err := json.Marshal(requests)
		if err != nil {
			fmt.Println("Failed to serialize data for ", connInfo.IP)
			continue
		}

		_, err = connInfo.Conn.Write(data)
		if err != nil {
			fmt.Println("Failed to send data to ", connInfo.IP, ":", err)
			continue
		} else {
			fmt.Println("Successfully sent data to ", connInfo.IP)
		}

	}
}
