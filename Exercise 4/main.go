package main 

import(
	//"fmt"
	"net"
)

func main() {
	//sjekker om det allerede kjører en master
	conn, err := net.Dial ("tcp", address)
	if err != nil {
		startMaster(1) //ingen master funnet, starter som master
	} else {
		conn.Close()
		startBackup() //master funnet, start som backup
	}
}