package TransformElevator

import (
	"flag"
	"fmt"
	"project/network/bcast"
	"project/network/peers"
	"project/constant"
	"project/types"
	"time"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//
//	will be received as zero-values.
type ElMsg struct {
	Sender    string
	Status    bool
	Floor     int
	Dirn      types.MotorDirection
	Requests  [constant.NumFloors][constant.NumButtons]types.ReqState
	Behaviour types.ElevatorBehavior
}

func Transform_elevator(sender_id string, e types.ExternalElevator) ElMsg {
	return ElMsg{
		Sender:    sender_id,
		Status:    e.Status,
		Floor:     e.Elevator.Floor,
		Dirn:      e.Elevator.Dirn,
		Requests:  e.Elevator.Requests,
		Behaviour: e.Elevator.Behaviour,
	}
}
func Transform_back(msg ElMsg) (e types.ExternalElevator, sender_id string) {
	return types.ExternalElevator{
			Status: msg.Status,
			Elevator: types.Elevator{
				Floor:     msg.Floor,
				Dirn:      types.MotorDirection(msg.Dirn),
				Requests:  msg.Requests,
				Behaviour: types.ElevatorBehavior(msg.Behaviour),
			},
		},
		msg.Sender
}

func Set_up1(e types.ExternalElevator, id string) (outMsg chan ElMsg, outNoder chan peers.PeerUpdate) {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	

	// ... or alternatively, we can use the Local IP address.
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
	Tx := make(chan ElMsg)
	Rx := make(chan ElMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, Tx)
	go bcast.Receiver(16569, Rx)

	// The example message. We just send one of these every second.
	go func() {
		for {
			msg := Transform_elevator(id, e)
			Tx <- msg
			time.Sleep(5000 * time.Millisecond)
		}
	}()

	returnNoder := make(chan peers.PeerUpdate)
	returnMsg := make(chan ElMsg)

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			returnNoder <- p
			return returnMsg, returnNoder

		case a := <-Rx:

			if a.Sender == id {
				continue
			}

			returnMsg <- a
			return returnMsg, returnNoder

		}
	}

}

func Set_up2() {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string

	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the Local IP address.
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
	//Tx := make(chan ElMsg)
	Rx := make(chan ElMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	//go bcast.Transmitter(16569, Tx)
	go bcast.Receiver(16569, Rx)

	// The example message. We just send one of these every second.
	/*
		go func() {
			for {
				msg := Transform_elevator(id, *e)
				Tx <- msg
				time.Sleep(100 * time.Millisecond)
			}
		}()
	*/
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
			e, reciver_id := Transform_back(a)
			println("Received message from peer", reciver_id)
			fmt.Printf("  Floor: %d\n", e.Elevator.Floor)
			fmt.Printf("  Dirn: %d\n", e.Elevator.Dirn)
			fmt.Printf("  Behaviour: %d\n", e.Elevator.Behaviour)

		}
	}
}

/* For å teste com kan denne main-funk brukes:

func main() {
	var elev1 esm.ExternalElevator
	elev1.Status = true
	var elev2 elevator.Elevator


	elev2.Floor = 0
	elev2.Dirn = 1
	elev2.Behaviour = elevator.EB_Idle
	elev2.Requests = [4][3]int{
		{0, 0, 0},
		{0, 0, 0},
		{0, 0, 0},
		{0, 0, 0},
	}
	elev1.Elevator = elev2

	Transform_elevator.Set_up1(&elev1)

}


#I case a:=<-Rx, blir ut-printen slikt (from 1 vil være id til den noden som sendte):
Received message from 1: {Status:true Elevator:{Floor:0 Dirn:1 Requests:[[0 0 0] [0 0 0] [0 0 0] [0 0 0]] Behaviour:0}}
OBS, sender_id er en string, for å kunne bruke peers.Transmitter. I utgangspunktet ikke krise, men kan også byttes over til int om nødvendig.

*/
