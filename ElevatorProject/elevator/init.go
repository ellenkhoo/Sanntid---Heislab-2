package elevator

// Search for master

//

func initSystem() {

	masterIP, found := listenForMaster()

	if found {
		connectToMaster(masterIP) // Join as backup/slave
	} else {
		go announceMaster() // Start broadcasting as master
		startMaster()       // Accept connections
	}

}
