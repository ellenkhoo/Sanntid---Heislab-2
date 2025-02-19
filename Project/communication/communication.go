package communicationpkg

import (
	"bufio"
	"fmt"
	"net"
	// "bufio"
	// "time"
	// "strconv"
)

const (
	lab_IP = "10.100.23.29:8080"
	sandra_IP = "10.22.216.146:8080"
)

func Comm_masterConnectToSlave () (conn net.Conn){
	ln, err := net.Listen("tcp", lab_IP)
	if err != nil {
		fmt.Println("Error starting master:", err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		} else {
			return conn
		}

	}
}

func Comm_slaveConnectToMaster () (conn net.Conn) {
	for {
		conn, err := net.Dial("tcp", lab_IP)
		if err != nil {
			fmt.Println("Error Starting backup:", err)
		} else {
			return conn
		}
		defer conn.Close()
	}
}

func Comm_sendMessage (message string, conn net.Conn) {

	_, err := conn.Write([]byte(message))
	if err != nil {
		fmt.Printf("Error writing message")
	}
}

func Comm_receiveMessage (conn net.Conn) {
	reader := bufio.NewReader(conn)
	
	for {
		message, err := reader.ReadString('\x00')
		if err != nil {
			fmt.Println("Error reading from connection: ", err)
			return
		}
		fmt.Println("Received message: ", message)
	}
}