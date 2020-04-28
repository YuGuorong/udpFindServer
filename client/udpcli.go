package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const DISCVR_SRV_PORT = 9981

type UdpIP struct {
	ipHost net.IP
	brCast net.IP
}

var wg sync.WaitGroup
var szCmd = "GW_NETINFO"

func lookupNetInfs() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ipList := make([]UdpIP, 0)
	for idxInf, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				inf, err := net.InterfaceByIndex(idxInf)
				if err != nil || strings.Contains(inf.Name, "lxcbr") || strings.Contains(inf.Name, "docker") {
					continue
				}
				bcastip := net.IP{ip4[0], ip4[1], ip4[2], ip4[3]}
				for i := 0; i < 4; i++ {
					bcastip[i] = (ipnet.Mask[i] ^ 0xFF) | bcastip[i]
				}
				ipList = append(ipList, UdpIP{ip4, bcastip})
				//fmt.Println(ip4)
				//fmt.Println(bcastip)
			}
		}
	}

	for _, udpip := range ipList {
		wg.Add(1)
		go udpPing(udpip)
	}
}

func udpPing(uip UdpIP) {
	defer func() {
		wg.Done()
	}()
	fmt.Println("ping network :", uip.brCast.String())
	srcAddr := &net.UDPAddr{IP: uip.ipHost, Port: 0}
	dstAddr := &net.UDPAddr{IP: uip.brCast, Port: DISCVR_SRV_PORT}
	conn, err := net.ListenUDP("udp", srcAddr)
	if err != nil {
		fmt.Println(err)
	}
	n, err := conn.WriteToUDP([]byte(szCmd), dstAddr)
	if err != nil {
		fmt.Println(err)
	}
	data := make([]byte, 1024)
	if err := conn.SetReadDeadline(time.Now().Add(time.Second * 3)); err != nil {
		return
	}
	for {
		n, _, err = conn.ReadFrom(data)
		if err != nil {
			if _, ok := err.(net.Error); ok {
				conn.Close()
				return
			}
			fmt.Println(err)
		}
		fmt.Printf("Find server at %s\n", data[:n])
	}
	conn.Close()
}

func main() {
	if len(os.Args) > 1 && strings.Compare(os.Args[1], "ip") == 0 {
		szCmd = "GW_GETIP"
	}

	lookupNetInfs()
	wg.Wait()
}
