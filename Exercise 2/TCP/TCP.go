package main

import (
	"fmt"
	"net"
	"bufio"
)

var server_IP = "10.100.23.204"
var localIP = "10.100.23.29"

var fixedSizePort = 34933
var zeroTerminatedPort = 33546
var localPort = "20019"

func connectToServer(serverPort int) (net.Conn) {
	addr := net.TCPAddr{
		Port: serverPort, 
		IP: net.ParseIP(server_IP),
	}

	conn, err := net.Dial("tcp", addr.String())
	if err != nil {
		return nil
	}
	return conn
}

func receiverTCP(conn net.Conn){

	reader := bufio.NewReader(conn)
	for {
		response, err := reader.ReadString('\x00')
		if err != nil {
			fmt.Println("Error reading from connection: ", err)
			return
		}
		fmt.Println("Received message: ", response)
	}
}

func senderTCP(message string, conn net.Conn){
	
	_, err := conn.Write([]byte(message))
	if err != nil{
		fmt.Println("error: ", err)
	} else {
		fmt.Println("Message sent: ", message)
	}
}

func requestAndAccept(conn net.Conn) net.Conn{
	message := "Connect to: 10.100.23.29:20019\x00" // Why \x00?
	senderTCP(message, conn)


	// Listen for response
	ln, err := net.Listen("tcp", "10.100.23.29:20019")
	if err != nil {
		fmt.Println("Error listening: ", err)
		return nil
	}

	for {
		newConn, err := ln.Accept()
		if err != nil {	
			fmt.Println("Error accepting connection: ", err)
			return nil
		}
		return newConn
	}
	
}

func main() {

	message := "Hello World from desk 19 :)\x00"
	
	conn := connectToServer(zeroTerminatedPort)
	defer conn.Close()

	newConn := requestAndAccept(conn)
	defer newConn.Close()

	go receiverTCP(newConn)
	go senderTCP(message, newConn)

	select {}
}
