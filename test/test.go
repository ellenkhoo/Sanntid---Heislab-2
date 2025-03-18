package main

import ( 
	"net"
	"os"
	"fmt"
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