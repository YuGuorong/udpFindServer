package main

import (
	"fmt"
	"net"
	"os"
)

const DISCVR_SRV_PORT = 9981

func main() {
	ip := net.ParseIP("192.168.3.255")
	//	ip := net.ParseIP("172.24.14.255")
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: ip, Port: 9981}
	conn, err := net.ListenUDP("udp", srcAddr)
	if err != nil {
		fmt.Println(err)
	}
	n, err := conn.WriteToUDP([]byte("hello"), dstAddr)
	if err != nil {
		fmt.Println(err)
	}
	data := make([]byte, 1024)
	n, _, err = conn.ReadFrom(data)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("read %s from <%s>\n", data[:n], conn.RemoteAddr())
	b := make([]byte, 1)
	os.Stdin.Read(b)
}
