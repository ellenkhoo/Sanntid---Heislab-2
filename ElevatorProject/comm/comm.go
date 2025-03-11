package comm

import (
	elevio "ElevatorProject/Driver"
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
	t := time.Duration(RandRange(500, 1000))
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
		time.Sleep(2 * time.Second) //announces every 2nd second, maybe it should happen more frequently?
	}
}

func ConnectToMaster(masterIP string, listenPort string) (int, net.Conn, bool) {
	conn, err := net.Dial("tcp", masterIP+":"+listenPort)
	if err != nil {
		fmt.Println("Error connecting to master:", err)
		return 0, nil, false
	}

	defer conn.Close() //want to use conn later in the program

	buffer := make([]byte, 1024)
	n, _ := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from master:", err)
		return 0, nil, false
	}

	var rank int
	_, err = fmt.Sscanf(string(buffer[:n]), "You have rank %d\n", &rank)
	if err != nil {
		fmt.Println("Error parsing rank:", err)
		return 0, nil, false
	}

	fmt.Printf("Connected to master at %s and received rank %d\n: ", masterIP, rank)
	return rank, conn, true
}

func ReceiveAssignedRequests(conn net.Conn) {
	defer conn.Close() //why do we need to do that?

	for {
		buffer := make([]byte, 4096) //might have to adjust size
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed by master")
			} else {
				fmt.Println("Error reading data:", err)
			}
			return
		}

		var receivedRequests [][2]bool
		err = json.Unmarshal(buffer[:n], &receivedRequests)
		if err != nil {
			fmt.Println("Failed to decode assigned requests:", err)
			continue
		}

		// do somthing with receivedRequests
	}
}

// func Comm_slaveConnectToMaster() (conn net.Conn) {
// 	for {
// 		conn, err := net.Dial("tcp", lab_IP)
// 		if err != nil {
// 			fmt.Println("Error Starting backup:", err)
// 		} else {
// 			return conn
// 		}
// 		defer conn.Close()
// 	}
// }

func Comm_sendMessage(message interface{}, conn net.Conn) {
	data, err1 := json.Marshal(message)
	if err1 != nil {
		fmt.Println("Error encoding message: ", err1)
		return
	}

	_, err2 := conn.Write(data)
	if err2 != nil {
		fmt.Println("Error writing message: ", err2)
	}
}

func Comm_receiveMessage(conn net.Conn) {
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

func Comm_sendReceivedOrder(order elevio.ButtonEvent, IP int, conn net.Conn) {
	//sender ordre til master når en ordre er motatt
	//Sender også med heisens IP-adresse, slik at cab-calls registreres på riktig heis

}

func Comm_masterReceiveOrder() {
	//Finn ut hvilken type melding som har kommet
	//send ordreliste/matrise til backup
	//avvent bekreftelse fra backup
	//Kjør fordelingsalgoritme
	//Send ordreliste -> setAllLights()
}

func Comm_slaveReceiveRequests() {
	//Bekrefte lister
	//Oppdater RequestsToDo
	//setAllLights()
	//Utføre requests (gjøres kontinuerlig)
}

func Comm_sendCurrentState(state interface{}, conn net.Conn) {

	data, err := json.Marshal(state)
	if err != nil {
		fmt.Println("Failed to encode state: ", err)
		return
	}

	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("Failed to send state: ", err)
		return
	}

}
