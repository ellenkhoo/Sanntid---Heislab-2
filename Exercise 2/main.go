package main

import (
	"fmt"
	"net"
)

var IPport = 30000
var seat = 19
var sendingPort = 20000 + seat

var server_IP = "10.100.23.204"
var localIP = "10.100.23.255"

func receiver(port int) {

	//Define address to listen to
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}

	//Create UDP Socket
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println("Error creating UDP socket:", err)
		return
	}

	//Closes the connection when finished?
	defer conn.Close()

	fmt.Println("Listening for UDP packets on port", IPport, "...")

	//Create buffer
	buffer := make([]byte, 1024)

	//Read data from server
	// for i := 0; i < 5; i++ {
	// 	n, RemoteAddr, err := conn.ReadFromUDP(buffer)
	// 	if err != nil {
	// 		fmt.Println("Error reading from UDP socket:", err)
	// 		continue
	// 	}

	// 	//Print received message
	// 	fmt.Printf("Received %d bytes from %s: %s:", n, RemoteAddr, string(buffer[:n]))
	// }

	n, RemoteAddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("Error reading from UDP socket:", err)
	}

	//Print received message
	fmt.Printf("Received %d bytes from %s: %s:", n, RemoteAddr, string(buffer[:n]))
}

func sender(message string, port int) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(server_IP),
	}

	conn, err := net.DialUDP("udp", nil, &addr)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	_, err = conn.Write([]byte(message))

	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		fmt.Println("Message sent: ", message)
	}

	defer conn.Close()
}

func main() {

	//make buffer
	// buffer := make(chan int, 1024)
	message := "Hello World"

	go receiver(sendingPort)
	go sender(message, sendingPort)

	select {}
}
