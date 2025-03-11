package roles

import (
	"fmt"
	"time"
	"net"
)

func StartSlave(conn net.Conn) {
	for {
		time.Sleep(5 * time.Second)
		fmt.Println("Still slave")
	}
	// Placeholder
}
