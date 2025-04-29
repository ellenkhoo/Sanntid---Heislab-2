<h1>ElevatorProject</h1>

**ElevatorProject** is a distributed control system for a network of elevators, developed for the TTK4145 Sanntidsprogrammering course at NTNU. 

Usage
-----------------------
The program can be run using an elevatorserver, which connects to a physical elevator, or a simulator. Both can be found in the **TTK4145** git repository, linked below.

Running the program requires a go compiler.
Run the program by entering ```go run main.go --num <value>```. The specified value is used to differentiate the clients, so make sure all clients run with different values.

In **initElevator.go**, the function **InitializeElevatorDriver** is passed an address. When using the elevatorserver, make sure the port is set to '15657'. If running multiple elevators on the same computer, make sure the port matches those used by the server. Each server should use a unique port.

Project overview
-----------------------
The project is based on a master-slave architecture, where all TCP communication goes through the master. Each client, including the master itself, controls an elevator through internal channels.

When the program is run, the first to connect on the network is considered the master and will broadcast itself, using UDP, allowing clients to connect to it, by creating a TCP connection. These will be considered backups.

<h3> Message flow </h3>

When an event, like a button press, happens on an elevator, a message is sent to the master. The master then checks what type of message was received and reacts accordingly. It then notifies all clients of its updated worldview. All elevators run continuously, reacting to updates as they happen.


Modules 
-----------------------
Our project uses two main modules; **elevator** and **network**. Additionally, there are modules named **timers** and **sharedConsts**.

<h4>elevator</h4>

This module has been derived from the project resources algorithm for a single elevator.
It has been modified to work with networking.
- **elevator.go** includes definitions of the elevator struct and all other structs it needs to run
- **elevatorIO.go** binds elevator hardware functions. Some functions are taken from **driver-go** in the **TTK4145** git repository.
- **elevatorLogic.go** contains the main loop which allows the FSM to react and act upon the events which may affect its state
- **fsm.go** makes the finite state machine and all functions needed to transition between states
- **initElevator.go** is used for initializing the elevator object when the program is first run
- **request.go** includes all functions relating to orders, and helps the FSM determine the elevator's next destination based on pending requests

<h4>hallRequestAssigner</h4>

The module includes two executable files, one for Windows and one for Linux. When a new order has been placed and sent to the master, the function **HallRequestAssigner** is used to assign hall requests to the clients on the network. It has been built from the source code, which is part of the 'TTK4145/Project-resources' git repository, in the directory **cost_fns/hall_request_assigner**.

<h4>network </h4>

This module includes all functions relating to the communication of the elevators on the network
- **configTCP.go** is used to ensure all our TCP connections are using the settings we wish
- **network.go** includes all networking functions that are used by all users of the network
- **networkClient.go** includes networking functions that are used by all clients on the network
- **networkConsts.go** includes definitions of structs that are used by the network. 
    - If you encounter problems with creating a TCP connection, try changing the TCPPort variable.
- **networkMaster.go** includes all networking functions that only the master uses

<h4>sharedConsts</h4>

This module defines shared constants, message types, and communication channels. It is used by both the **network** and **elevator** modules.

<h4>timers</h4>

This module implements timer functionality. It is used exclusively by the elevator to handle the door's behaviour.


<h2>Limitations and known bugs</h2>

- **Failover:** If the master dies, the system struggles to transition to a new master, as we struggle to shut down previous goroutines.
- **Packet loss:** The network does not handle packet loss very well. 

<h2>Resources</h2>

All resources provided through the course can be found here:
https://github.com/TTK4145
