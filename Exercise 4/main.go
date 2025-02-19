package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const (
	masterPort = ":8080"
	backupPort = ":8081"
)

var counter = 0

func main() {
	if len(os.Args) > 1 && os.Args[1] == "backup" {
		if len(os.Args) > 2 {
			counter, _ = strconv.Atoi(os.Args[2])
		}
		runBackup()
	} else {
		runMaster(counter)
	}
}

func runMaster(counter int) {
	ln, err := net.Listen("tcp", masterPort)
	if err != nil {
		fmt.Println("Error starting master:", err)
		return
	}
	defer ln.Close()
	fmt.Println("Master started, waiting for backup...")

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

func runBackup() {
	var backupCounter int
	for {
		conn, err := net.Dial("tcp", "localhost"+masterPort)
		if err == nil {
			fmt.Println("Connected to master")
			reader := bufio.NewScanner(conn)
			for reader.Scan() {
				text := reader.Text()
				num, err := strconv.Atoi(text)
				if err != nil {
					fmt.Println("Error reading from master:", err)
					break
				}
				backupCounter = num
			}
			if err := reader.Err(); err != nil {
				fmt.Println("Master connection lost, becoming master")
				conn.Close()
				runMaster(backupCounter)
				return
			}
			conn.Close()

		} else {
			fmt.Println("No master found, becoming master")
			go runNewBackup(backupCounter)
			runMaster(backupCounter)
		}
		time.Sleep(1 * time.Second)
	}
}

func runNewBackup(backupCounter int) {
	cmd := exec.Command("gnome-terminal", "--", "bash", "-c", fmt.Sprintf("go run main.go backup %d", backupCounter))
	err := cmd.Start()
	if err != nil {
		fmt.Println("Error starting new backup:", err)
	}
}
