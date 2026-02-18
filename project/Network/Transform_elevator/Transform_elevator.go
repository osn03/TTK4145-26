package Transform_elevator

import (
	"Project/Network/bcast"
	"Project/Network/peers"
	"Project/elevator"
	"Project/elevio"
	"flag"
	"fmt"
	"time"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//
//	will be received as zero-values.
type ElevatorMsg struct {
	Sender    int
	Floor     int
	Dirn      int
	Requests  [constant.numFloors][constant.numButtons]bool
	Behaviour int
}
func Transform_elevator(e elevator.Elevator, sender int) ElevatorMsg {
	return ElevatorMsg{
		Sender:    sender,
		Floor:     e.Floor,
		Dirn:      int(e.Dirn),
		Requests:  e.Requests,
		Behaviour: int(e.Behaviour),
	}
}
func Transform_back(msg ElevatorMsg) elevator.Elevator {
	return elevator.Elevator{
		Floor:     msg.Floor,
		Dirn:      elevio.MotorDirection(msg.Dirn),
		Requests:  msg.Requests,
		Behaviour: elevator.ElevatorBehavior(msg.Behaviour),
	}
}


func Set_up(e *elevator.Elevator, sender string) {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id int
	flag.IntVar(&id, "id", 0, "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	Tx := make(chan ElevatorMsg)
	Rx := make(chan ElevatorMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, Tx)
	go bcast.Receiver(16569, Rx)

	// The example message. We just send one of these every second.
	go func() {
    	for {
        msg := Transform_elevator(*e, id)
        Tx <- msg
        time.Sleep(100 * time.Millisecond)
    }
}()

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-Rx:
			if a.Sender == id {
				continue
			}
			*e = Transform_back(a)


		}
	}
}
