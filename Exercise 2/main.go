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

	//bør egt connecte til serveren i main-func, med dial

	//make buffer
	// buffer := make(chan int, 1024)
	message := "Hello World"

	go receiver(sendingPort)
	go sender(message, sendingPort)

	select {}
}


/*
//1.2: TCP - newest 

func recevierTCP(conn net.Conn, port int){
	//define address to listen to 
	addr:=net.Addr{
		Port: port, 
		IP: net.ParseIP("0.0.0.0"),
	}

	buffer := make([]byte, 1024)
	for{
		n, err := conn.Read(buffer) //listen brukes når en selv er server
		if err != nil{
			fmt.Println("error receiving message: ", err)
			return
		} //blir man stuck
		fmt.Println("message recevied from server: ", string(buffer[:net]))
	}
}

func senderTCP(conn net.Conn, message String, port int){
	//define address to listen to 
	addr:=net.Addr{
		Port: port, 
		IP: net.ParseIP("0.0.0.0"),
	}
	//oppretter forbindelse med server
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	//defer.Close()

	//create a fixed size message
	const size = 1024
	if len(message) > size{
		message = message[: size]
	}
	fixedSizeMessage := message + string(make([]byte, size-len(message)))

	_, err = conn.Write([]byte(fixedSizeMessage))
	if err != nil{
		fmt.Println("error: ", err)
	} else {
		fmt.Println("Message sent: ", message)
	}
	defer.Close()
}
/*
//i main funksjon 
conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()
*/

//___________________________________________________________________




//1.2:TCP 
func receiverTCP(port int){
	//define address to listen to 
	addr:=net.TCPAddr{
		Port: port, 
		IP: net.ParseIP("0.0.0.0"),
	}

	//create TCP socket 
	listener, err := net.Listen("tcp", &addr)
	if err != nil{
		fmt.Println("error creating TCP socket:", err)
		return
	}

	//closes the connetction 
	defer listener.Close()

	fmt.Println("Listen for TCP packets on port:", port)

	//trenger egt kun når vi selv er server
	for {
		//accept connections //tcp er protokoll, så en forbindelse må oppdrettes
		//mellom klient og server før data sendes eller mottas
		//ved en udp tilkobling trengs det ingen forbindelse og dermed ingen accept
		conn, err := listener.Accept()
		if err != nil{
			fmt.Println("Error acceptinp connection: ", err)
			continue
		}
		//handle incoming message in a separat goroutine
		go func (c net.Conn){ //etter at tilkoblingen er etablert, opprettes 
			//en dedikert kommunikasjonskanal mellom klienten og serveren
			defer c.Close()
			buffer := make([]byte, 1024)
			n, err := c.Read(buffer)
			if err != nil {
				fmt.Println("error reading from connection: ", err)
				return 
			}
			fmt.Println("Received message: ", string(bufer[:n]))
		} (conn)
	}

	
	
}

func senderTCP(message string, serverIP string, port int){
	addr := net.TCPAddr{
		Port: port //34933, 
		IP: net.ParseIP(server_IP),
	}

	conn, err := net.Dial("tcp", &addr)
	if err != nil{
		fmt.Println("error: ", err)
	}
	defer conn.Close()
	
	//create a fixed size message
	const size = 1024
	if len(message) > size{
		message = message[: size]
	}
	fixedSizeMessage := message + string(make[]byte, size-len(message))

	_, err = conn.Write([]byte(fixedSizeMessage))
	if err != nil{
		fmt.Println("error: ", err)
	} else {
		fmt.Println("Message sent: ", message)
	}
}
