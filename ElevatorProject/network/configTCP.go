package network

import (
	"fmt"
	"net"
)

func ConfigureTCPConn(conn net.Conn) (*net.TCPConn, error) {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return nil, fmt.Errorf("Failed to cast net.Conn to net.TCPConn")
	}

	// Apply TCP optimizations
	tcpConn.SetNoDelay(true)        // Disable Nagle's algorithm
	tcpConn.SetReadBuffer(2500000)  // Increase read buffer
	tcpConn.SetWriteBuffer(2500000) // Increase write buffer

	return tcpConn, nil
}
