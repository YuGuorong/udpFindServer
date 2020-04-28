package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

const DISCVR_SRV_PORT = 9981

type UdpIP struct {
	ipHost net.IP
	brCast net.IP
}

var wg sync.WaitGroup
var ipList []UdpIP
var sInfInfo string

func lookupNetInfs() {
	sInfInfo = ""
	infs, err := net.Interfaces()
	for _, ninf := range infs {
		if ninf.HardwareAddr != nil &&
			!strings.Contains(ninf.Name, "lxcbr") && !strings.Contains(ninf.Name, "docker") {
			sInfInfo += fmt.Sprintf("\n  [%12s | ", ninf.Name)
			sInfInfo += ninf.HardwareAddr.String() + "]"
		}
	}

	sInfInfo += " \n"
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ipList = make([]UdpIP, 0)
	for idxInf, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				inf, err := net.InterfaceByIndex(idxInf + 1)
				if err != nil || ipnet.IP.IsLoopback() || strings.Contains(inf.Name, "lxcbr") || strings.Contains(inf.Name, "docker") {
					continue
				}
				bcastip := net.IP{ip4[0], ip4[1], ip4[2], ip4[3]}
				for i := 0; i < 4; i++ {
					bcastip[i] = (ipnet.Mask[i] ^ 0xFF) | bcastip[i]
				}
				ipList = append(ipList, UdpIP{ipHost: ip4, brCast: bcastip})
				//fmt.Println(ip4)
				//fmt.Println(bcastip)
			}
		}
	}
	udpSvr(ipList)
}

func runSvr(ctx context.Context, uip UdpIP) {
	defer func() {
		fmt.Println("proc:", uip.ipHost.String, " Exit!")
		wg.Done()
	}()

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: uip.brCast, Port: DISCVR_SRV_PORT})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Starting listening at :", uip.brCast.String())
	for {
		select {
		case <-ctx.Done():
			conn.Close()
			return
		default:
			cmd := 0
			data := make([]byte, 256)
			nrr, rAddr, err := conn.ReadFromUDP(data)
			if err != nil {
				fmt.Println(err)
				continue
			}

			strData := string(data[0:nrr])
			fmt.Println(strData)
			if strings.Compare(strData, "GW_GETIP") == 0 {
				cmd = 1
			} else if strings.Compare(strData, "GW_NETINFO") == 0 {
				cmd = 2
			} else {
				continue
			}

			upper := fmt.Sprintf("%s", uip.ipHost.String())
			if cmd == 2 {
				upper += "{"
				for i, uip := range ipList {
					if i != 0 {
						upper += "; "
					}
					upper = upper + fmt.Sprintf("%s", uip.ipHost.String())
				}
				upper += " } | " + sInfInfo
			}
			n, err := conn.WriteToUDP([]byte(upper), rAddr)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("Send:[", n, "]:", upper)
		}

	}
}

func freeUdpSvr(uip UdpIP) {
	srcAddr := &net.UDPAddr{IP: uip.ipHost, Port: 0}
	dstAddr := &net.UDPAddr{IP: uip.brCast, Port: DISCVR_SRV_PORT}
	conn, err := net.ListenUDP("udp", srcAddr)
	if err != nil {
		fmt.Println(err)
	}
	_, err = conn.WriteToUDP([]byte("exit Read"), dstAddr)
	conn.Close()
}

func udpSvr(ipList []UdpIP) {
	fmt.Println("starting service...")

	ctx := context.Background()
	ctxWithCancel, cancelFunction := context.WithCancel(ctx)

	for _, uip := range ipList {
		wg.Add(1)
		go runSvr(ctxWithCancel, uip)
	}

	var wait string
	fmt.Scanln(&wait)
	fmt.Println("Waiting process exit...")
	cancelFunction()
	for _, uip := range ipList {
		freeUdpSvr(uip)
	}

	wg.Wait()
}

func main() {
	lookupNetInfs()
}
