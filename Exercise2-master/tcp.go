package main

import (
	"fmt"
	"net"
)

/*
	func receiveTCP(addr *net.TCPAddr) {
		conn, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		//var pay []byte



		for {
			_, err := conn.Read(buf)
			if err != nil {
				return
			}
			fmt.Println(string(buf))
		}

}
*/
func sendTCP(addr *net.TCPAddr) {
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return
	}
	defer conn.Close()
	buf := make([]byte, 1024)
	//var pay []byte

	n, err := conn.Read(buf)
	if err != nil {
		return
	}
	fmt.Println(string(buf[:n]))
	conn2, _ := net.Listen("tcp", ":10000")
	conn.Write([]byte("Connect to: 10.100.23.33:10000\000"))
	callback, _ := conn2.Accept()

	n, err = callback.Read(buf)

	if err != nil {
		return
	}
	fmt.Println(string(buf[:n]))
	callback.Write([]byte("hei igjen\000"))
	callback.Read(buf)
	fmt.Println(string(buf[:n]))

}

func main() {
	/*receiveaddr := &net.TCPAddr{
		IP:   net.ParseIP("10.100.23.11"),
		Port: 34933,
		Zone: "",
	}*/
	sendaddr := &net.TCPAddr{
		IP:   net.ParseIP("10.100.23.11"),
		Port: 33546,
	}

	sendTCP(sendaddr)
}

//10.100.23.3310.100.23.33
