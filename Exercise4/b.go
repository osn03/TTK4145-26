package main

import (
	"net"
	"os/exec"
	"time"
	"fmt"
)

func main() {

	sendaddr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 20023,
		Zone: "",
	}
	println("hei")
	exec.Command("gnome-terminal", "--", "go", "run", "/home/student/Desktop/Heis21/Exercise4-master/a.go").Run()
	for i := 0; i < 10; i++ {
		send(sendaddr, fmt.Sprint(i))


		time.Sleep(time.Second * 1)
		
	}

}

func send(addr *net.UDPAddr, message string) {
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}

	defer conn.Close()

	for i := 0; i < 100; i++ {
		conn.Write([]byte(message))
		time.Sleep(10 * time.Millisecond)
	}
}
