package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const DISCVR_SRV_PORT = 9981

var ipList []net.IP

func queryIpList() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ipList = make([]net.IP, 0)
	for idxInf, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				inf, err := net.InterfaceByIndex(idxInf)
				if err != nil || strings.Contains(inf.Name, "lxcbr") || strings.Contains(inf.Name, "docker") {
					continue
				}
				ipList = append(ipList, ip4)
				fmt.Println(ip4)
				for i := 0; i < 4; i++ {
					ip4[i] = (ipnet.Mask[i] ^ 0xFF) | ip4[i]
				}
				fmt.Println("Broadcast:", ip4)
				fmt.Println("IP：", ipnet.IP.String())
				go udpSvr(ipnet.IP, ip4)
			}
		}
	}

}

func udpSvr(hostip net.IP, bcastIP net.IP) {
	fmt.Println("new solution")
	//address := "192.168.3.255" + ":" + "9981"
	addr := &net.UDPAddr{IP: bcastIP, Port: 9981}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer conn.Close()

	for {
		// Here must use make and give the lenth of buffer
		data := make([]byte, 256)
		nrr, rAddr, err := conn.ReadFromUDP(data)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println(nrr)
		strData := string(data[0:nrr])
		fmt.Println("Received:", strData)
		fmt.Println(rAddr)

		upper := "R-" + strData
		n, err := conn.WriteToUDP([]byte(upper), rAddr)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("Send:[", n, "]:", upper)
	}
}

func main() {
	queryIpList()
}
