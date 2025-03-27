package network

import (
	"fmt"
	"net"
)

// Usikker på om det faktisk hjalp noe
func ConfigureTCPConn(conn net.Conn) (*net.TCPConn, error) {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return nil, fmt.Errorf("failed to cast net.Conn to net.TCPConn")
	}

	// Apply TCP optimizations
	tcpConn.SetNoDelay(true)        // Disable Nagle's algorithm
	tcpConn.SetReadBuffer(2500000)  // Increase read buffer
	tcpConn.SetWriteBuffer(2500000) // Increase write buffer

	fmt.Println("TCP settings applied successfully!")
	return tcpConn, nil
}
