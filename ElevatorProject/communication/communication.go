package comm

import (
	"Driver-go/elevio"
	"bufio"
	"fmt"
	"net"
	// "bufio"
	// "time"
	// "strconv"
	"encoding/json"
)

const (
	lab_IP = "10.100.23.29:8080"
	local_IP = "10.22.216.146:8080"
)

func Comm_findOwnIP() net.IP{
	var own_ipv4 net.IP
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
    	if ipv4 = addr.To4(); ipv4 != nil {
        	fmt.Println("IPv4: ", ipv4)
    	}   
	}
}

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
			//Denne blokken setter keepalive, en slags "heartbeat" som automatisk gjør jobben
			//Tror denne bør settes inn hver gang en kobling gjøres så følger de med på hverandre
			aliveErr = ln.SetKeepAlive(true) 
			if aliveErr != nil {
				fmt.Printf("Unable to set keepalive: %s \n", aliveErr)
			} else {
				//Sjekker om andre siden av koblingen fortsatt er i live etter 5 sekunder
				aliveErr2 = ln.SetKeepAlivePeriod(5 * time.Second) 
				if aliveErr2 != nil {
					fmt.Printf("Unable to set keepalive interval: %s", aliveErr2)
				}
			}
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
			//Denne blokken setter keepalive, en slags "heartbeat" som automatisk gjør jobben
			//Tror denne bør settes inn hver gang en kobling gjøres så følger de med på hverandre
			aliveErr = ln.SetKeepAlive(true) 
			if aliveErr != nil {
				fmt.Printf("Unable to set keepalive: %s \n", aliveErr)
			} else {
				//Sjekker om andre siden av koblingen fortsatt er i live etter 5 sekunder
				aliveErr2 = ln.SetKeepAlivePeriod(5 * time.Second) 
				if aliveErr2 != nil {
					fmt.Printf("Unable to set keepalive interval: %s", aliveErr2)
				}
			}
			return conn
		}
		defer conn.Close()
	}
}

func Comm_sendMessage (message interface{}, conn net.Conn) {
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


func Comm_sendCurrentState (state interface{}, conn net.Conn) {
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


