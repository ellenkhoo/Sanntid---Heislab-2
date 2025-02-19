package communicationpkg

import (
	"net"
	"fmt"
	"bufio"
	"time"
	"strconv"
)

const (
	lab_IP = "10.100.23.29"
	sandra_IP = "10.22.216.146"
)

func Comm_masterConnectToSlave () {
	var counter int = 0
	ln, err := net.Listen("tcp", lab_IP)
	if err != nil {
		fmt.Println("Error starting master:", err)
		return
	}
	defer ln.Close()

	number := counter
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		fmt.Printf("Master accepted connection\n")

		writer := bufio.NewWriter(conn)

		for {
			fmt.Printf("Master sent: %d\n", number)
			_, err := writer.WriteString(strconv.Itoa(number) + "\n")
			writer.Flush()
			if err != nil {
				fmt.Println("Error writing to backup:", err)
				break
			}
			number++
			time.Sleep(1 * time.Second)
		}

	}
}

func Comm_slaveConnectToMaster () {
	for {
		conn, err := net.Dial("tcp", sandra_IP)
		if err != nil {
			fmt.Println("Error Starting backup:", err)
		} else {
			fmt.Println("Backup connected succesfully")
		}
		defer conn.Close()

		reader := bufio.NewScanner(conn)

		for reader.Scan() {
			text := reader.Text()
			num, err := strconv.Atoi(text)
			if err != nil {
				fmt.Println("Error reading from master:", err)
				break
			}
			fmt.Printf("Number received: %d \n", num)
		}
		if err := reader.Err(); err != nil {
			fmt.Println("Master connection lost")
			conn.Close()
			return
		}
		conn.Close()
	
	time.Sleep(1 * time.Second)
	}
}