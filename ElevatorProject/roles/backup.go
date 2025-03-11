package roles

import (
	"fmt"
	"net"
	"time"
)

func StartBackup(conn net.Conn) {
	fmt.Println("Starting backup")
	for {
		time.Sleep(5 * time.Second)
		fmt.Println("Still backup")
	}
	// Placeholder
}
