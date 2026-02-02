package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
	"strconv"
)

func receive(addr *net.UDPAddr, message string) (string, bool){
	/*addr := &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 30000,
		Zone: "",
	}*/
	conn, err := net.ListenUDP("udp", addr)

	if err != nil {
		return message, false
	}
	defer conn.Close()

	var pay string
	n := 0

	for n == 0 {
		b := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(3000 * time.Millisecond))
		n2, _, err := conn.ReadFromUDP(b)
		n = n2
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return message, true
		}
		if err != nil {
			return message, false
		}
		pay = string(b[:n])
		if message == pay {
			return message, false
		}
		fmt.Println(pay)
	}

	return pay, false
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

func main() {
	addr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 20023,
		Zone: "",
	}

	println("h")
	var message string
	var dead bool
	for {

		message, dead = receive(addr, message)
		if dead{
			break
		}
	}
	println("hei")
	exec.Command("gnome-terminal", "--", "go", "run", "/home/student/Desktop/Heis21/Exercise4-master/a.go").Run()
	for i,_ := strconv.Atoi(message); i < 23; i++ {
		send(addr, fmt.Sprint(i+1))

		time.Sleep(time.Millisecond* 100)

	} 
}
