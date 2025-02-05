package main

import(
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

const(
	address = "localhost:9000"   //TCP-adressen for kommunikasjon 
	heartbeatTime = 2*time.Second //maks tid å vente før backup tar over 
)

//starter en TCP-server (master)
func startMaster(startNumber int) {
	fmt.Println ("Starting as Master...")
	ln, err := net.Listen("tcp", address)
	if err != nil{
		fmt.Println("Error starting Master:", err)
		os.Exit(1)
	}
	defer ln.Close()

	number := startNumber
	for {
		//venter på at en backup kobler til 
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error connecting backup:", err)
			continue
		}
		fmt.Println("Backup connected!")

		//sender telling til backup kontinuerlig 
		writer := bufio.NewWriter(conn)
		for {
			fmt.Println(number)
			_, err := writer.WriteString(strconv.Itoa(number) + "\n")
			writer.Flush()
			if err != nil{
				fmt.Println("Backup lost connection. Starting a new backup...")
				break
			}
			number++
			time.Sleep(1*time.Second) //venter 1sek mellom hver telling 
		}
	}	
}


//starter en backup som venter på at master feiler
func startBackup() {
	fmt.Println("Starting as backup...")
	for {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			fmt.Println("Master down! Starting as new master...")
			startMaster(1) //starter som master med verdi 1
			return
		}

	//leser tallet fra master
	reader := bufio.NewScanner(conn)
	lastNumber := 1
	for reader.Scan() {
		num, err := strconv.Atoi(reader.Text())
		if err != nil {
			fmt.Println("Error passing numbers:", err)
			break
		}
		lastNumber = num + 1
	}

	fmt.Println("Connection to master lost. Starting as new Master...")
	startMaster(lastNumber) //fortsetter tellingen fra siste mottatte tall
	}
}