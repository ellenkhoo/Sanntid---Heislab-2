package roles

import (
	"fmt"
	"sync"
)

type Connection struct {
	IP   string
	Rank int
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

	// Assign the next availiable rank
	rank := len(ac.conns) + 1
	ac.conns = append(ac.conns, Connection{IP: ip, Rank: rank})
}

func (ac *ActiveConnections) RemoveConnection(ip string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Find index of ip to be removed
	index := -1
	for i, conn := range ac.conns {
		if conn.IP == ip {
			index = i
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
}

func (ac *ActiveConnections) ListConnections() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	for _, conn := range ac.conns {
		fmt.Printf("IP: %s, Rank: %d\n", conn.IP, conn.Rank)
	}
}
