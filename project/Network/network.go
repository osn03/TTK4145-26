package network

import (
	"project/Network/Transform_elevator"
	"project/Network/peers"
	"project/constant"
	"project/elevator"
	"project/esm"
)

type ElevatorMsg struct {
	Sender    string
	Status    bool
	Floor     int
	Dirn      int
	Requests  [constant.NumFloors][constant.NumButtons]elevator.ReqState
	Behaviour int
}

func NetworkCum(in <-chan esm.ExternalElevator, outMsg chan<- Transform_elevator.ElevatorMsg, outNoder chan<- peers.PeerUpdate) {
	a, b := Transform_elevator.Set_up1(<-in)
	outMsg = a
	outNoder = b

}
