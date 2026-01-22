package main

import (
	"fmt"
	"net"
	"time"
)

func receive(addr *net.UDPAddr) {
	/*addr := &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 30000,
		Zone: "",
	}*/
	conn, err := net.ListenUDP("udp", addr)

	if err != nil {
		return
	}
	defer conn.Close()

	var pay string
	n := 0
	for n == 0 {
		b := make([]byte, 1024)
		n2, addr2, err := conn.ReadFromUDP(b)
		fmt.Println(addr2.IP)
		n = n2
		if err != nil {
			return
		}
		pay = string(b[:n])
		fmt.Println(pay)
		fmt.Println(n)
	}

	return
}

func send(addr *net.UDPAddr) {
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}

	defer conn.Close()

	for {
		conn.Write([]byte("Sending back"))
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	sendaddr := &net.UDPAddr{
		IP:   net.ParseIP("10.100.23.11"),
		Port: 20023,
		Zone: "",
	}

	receiveaddr := &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 20023,
		Zone: "",
	}

	go receive(receiveaddr)
	send(sendaddr)
}
