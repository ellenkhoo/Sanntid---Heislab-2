package roles

import (
	"fmt"
	"net"
)

//kanskje unødvendig med en egen slave-funksjon, da den bare kaller initElevator

func StartSlave(rank int, conn net.Conn) {
	
	fmt.Println("Starting slave")
	
	InitElevator(rank, conn)
}
