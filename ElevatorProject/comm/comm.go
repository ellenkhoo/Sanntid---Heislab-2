package comm

import (
	"github.com/ellenkhoo/ElevatorProject/elevator/Driver"
	"bufio"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"time"

	// "bufio"
	// "time"
	// "strconv"
	"encoding/json"
)

// const (
// 	lab_IP = "10.100.23.29:8080"
// 	local_IP = "10.22.216.146:8080"
// )

// should this be moved to someplace else?
func RandRange(min, max int) int {
	return rand.IntN(max-min) + min
}

func Comm_listenAndAccept(IP string) (conn net.Conn) {
	ln, err := net.Listen("tcp", IP)
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

func ListenForMaster(port string) (string, bool) {
	// addr := net.UDPAddr{Port: 9999, IP: net.ParseIP("0.0.0.0")} //change port?
	addr, _ := net.ResolveUDPAddr("udp", "0.0.0.0"+":"+port)
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting UDP listener:", err)
		return "", false //no existing master
	}

	defer conn.Close()

	buffer := make([]byte, 1024)
	t := time.Duration(RandRange(800, 1500))
	fmt.Printf("Waiting for %d ms\n", t)
	conn.SetReadDeadline(time.Now().Add(t * time.Millisecond)) //ensures that only one remains master
	_, remoteAddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("No master found, becoming master.")
		return "", false
	}

	fmt.Println("Master found at: ", remoteAddr.IP.String())
	return remoteAddr.IP.String(), true
}

func AnnounceMaster(localIP string, port string) {
	fmt.Println("Announcing master")
	broadcastAddr := "255.255.255.255" + ":" + port
	// addr, _ := net.ResolveUDPAddr("udp", "255.255.255.255:9999")
	conn, err := net.Dial("udp", broadcastAddr)
	//conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Error starting UDP listener:", err)
		return
	}
	defer conn.Close()

	for {
		msg := "I am Master"
		conn.Write([]byte(msg))
		time.Sleep(1 * time.Second) //announces every 2nd second, maybe it should happen more frequently?
	}
}

func ConnectToMaster(masterIP string, listenPort string) (net.Conn, bool) {
	conn, err := net.Dial("tcp", masterIP+":"+listenPort)
	if err != nil {
		fmt.Println("Error connecting to master:", err)
		return nil, false
	}

	// defer conn.Close() //want to use conn later in the program

	// buffer := make([]byte, 1024)
	// n, _ := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from master:", err)
		conn.Close()
		return nil, false
	}

	// Tror ikke vi trenger dette
	// var rank int
	// _, err = fmt.Sscanf(string(buffer[:n]), "You have rank %d\n", &rank)
	// if err != nil {
	// 	fmt.Println("Error parsing rank:", err)
	// 	conn.Close()
	// 	return 0, nil, false
	// }

	fmt.Printf("Connected to master at %s\n: ", masterIP)
	return conn, true
}

// func ReceiveAssignedRequests(conn net.Conn) {
// 	defer conn.Close() //why do we need to do that? Might have to remove in order to keep the conn for the entirety of the program

// 	for {
// 		buffer := make([]byte, 4096) //might have to adjust size
// 		n, err := conn.Read(buffer)
// 		if err != nil {
// 			if err == io.EOF {
// 				fmt.Println("Connection closed by master")
// 			} else {
// 				fmt.Println("Error reading data:", err)
// 			}
// 			return
// 		}

// 		var receivedRequests [][2]bool
// 		err = json.Unmarshal(buffer[:n], &receivedRequests)
// 		if err != nil {
// 			fmt.Println("Failed to decode assigned requests:", err)
// 			continue
// 		}

// 		// do somthing with receivedRequests
// 	}
// }

// func Comm_sendMessage(message interface{}, conn net.Conn) {
// 	data, err1 := json.Marshal(message)
// 	if err1 != nil {
// 		fmt.Println("Error encoding message: ", err1)
// 		return
// 	}

// 	_, err2 := conn.Write(data)
// 	if err2 != nil {
// 		fmt.Println("Error writing message: ", err2)
// 	}
// }

// func Comm_receiveMessage(conn net.Conn) {
// 	reader := bufio.NewReader(conn)

// 	for {
// 		message, err := reader.ReadString('\x00')
// 		if err != nil {
// 			fmt.Println("Error reading from connection: ", err)
// 			return
// 		}
// 		fmt.Println("Received message: ", message)
// 	}
// }

// func Comm_sendReceivedOrder(order elevio.ButtonEvent, conn net.Conn) {
// 	//sender ordre til master når en ordre er motatt
// 	if conn == nil {
// 		fmt.Println("Connection is nil")
// 		return
// 	}

// 	data, err := json.Marshal(order)
// 	if err != nil {
// 		fmt.Println("Failed to encode order: ", err)
// 		return
// 	}

// 	// TEST
// 	message := "order:" + string(data)
// 	_, err = conn.Write([]byte(message))

// 	fmt.Println("Sending data:", message) // Log the raw JSON data

// 	//_, err = conn.Write(data)
// 	if err != nil {
// 		fmt.Println("Failed to send order: ", err)
// 		return
// 	}
// 	fmt.Println("Sent current order to master.\n")
// }

// func Comm_masterReceive(conn net.Conn) {
// 	//Finn ut hvilken type melding som har kommet
// 	//send ordreliste/matrise til backup
// 	//avvent bekreftelse fra backup
// 	//Kjør fordelingsalgoritme
// 	//Send ordreliste -> setAllLights()

// 	fmt.Println("Arrived at Comm_masterReceiveOrder")

// 	if conn == nil {
// 		fmt.Println("Connection is nil")
// 		return
// 	}

// 	// Buffer to hold incoming data
// 	buffer := make([]byte, 1024)

// 	for {
// 		fmt.Println("Reading loop")
// 		// Read data from the connection
// 		n, err := conn.Read(buffer)
// 		if err != nil {
// 			fmt.Println("Error receiving data:", err)
// 			return
// 		}

// 		// Read the message and check its prefix
// 		message := string(buffer[:n])

// 		// Check if the message starts with "order" or "state"
// 		if len(message) >= 5 && message[:5] == "order" {
// 			var order elevio.ButtonEvent
// 			err := json.Unmarshal([]byte(message[6:]), &order) // Skip "order:"
// 			if err != nil {
// 				fmt.Println("Failed to unmarshal order:", err)
// 				return
// 			}
// 			fmt.Printf("Received order: Floor %d, Button %d\n", order.Floor, order.Button)
// 			continue

// 		}
// 		if len(message) >= 5 && message[:5] == "state" {
// 			fmt.Println("Received state message")
// 			// var state elevator.ElevStates
// 			// err := json.Unmarshal([]byte(message[6:]), &state) // Skip "state:"
// 			// if err != nil {
// 			// 	fmt.Println("Failed to unmarshal state:", err)
// 			// 	return
// 			// }
// 			// fmt.Printf("Received state: Behaviour %s, Floor %d, Direction %s\n", state.Behaviour, state.Floor, state.Direction)
// 			continue
// 		}

// 		fmt.Println("Received unrecognized message type")
// 	}

// }

// func Comm_slaveReceiveRequests() {
// 	//Bekrefte lister
// 	//Oppdater RequestsToDo
// 	//setAllLights()
// 	//Utføre requests (gjøres kontinuerlig)
// }

// // might be easier to understand by changing to 'state ElevStates' instead of interface{}
// func Comm_sendCurrentState(state interface{}, conn net.Conn) {

// 	if conn == nil {
// 		fmt.Println("Connection is nil")
// 		return
// 	}

// 	data, err := json.Marshal(state)
// 	if err != nil {
// 		fmt.Println("Failed to encode state: ", err)
// 		return
// 	}

// 	// TEST
// 	message := "state:" + string(data)
// 	_, err = conn.Write([]byte(message))

// 	//_, err = conn.Write(data)
// 	if err != nil {
// 		fmt.Println("Failed to send state: ", err)
// 		return
// 	}
// 	fmt.Println("Sent current state to master.")
// }
