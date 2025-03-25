package network

import (
	"fmt"
	"net"
)

func ConfigureTCPConn(conn net.Conn) (*net.TCPConn, error) {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return nil, fmt.Errorf("failed to cast net.Conn to net.TCPConn")
	}

	// Apply TCP optimizations
	tcpConn.SetNoDelay(true)        // Disable Nagle's algorithm
	tcpConn.SetReadBuffer(2500000)  // Increase read buffer
	tcpConn.SetWriteBuffer(2500000) // Increase write buffer
	// tcpConn.SetKeepAlive(true)                         // Enable keep-alive
	// tcpConn.SetKeepAlivePeriod(10 * time.Second)       // Keep-alive interval

	// Set congestion control to BBR (Linux only)
	// fd, err := tcpConn.File()
	// if err == nil {
	// 	defer fd.Close()
	// 	unix.SetsockoptString(int(fd.Fd()), unix.IPPROTO_TCP, unix.TCP_CONGESTION, "bbr")
	// }

	fmt.Println("TCP settings applied successfully!")
	return tcpConn, nil
}
