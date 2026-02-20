package main

import (
	"project/Network"
	"project/Network/peers"
	"project/constant"
	"project/elevator"
	"project/elevio"
	"project/esm"
	"project/fsm"
)

func main() {

	numFloors := constant.NumFloors
	//numElevs := constant.NumElevators

	elevio.Init("localhost:15657", numFloors)

	out := make(chan esm.ExternalElevator)
	elevatorin := make(chan network.Msg)
	statusin := make(chan peers.PeerUpdate)
	local := make(chan elevator.Elevator)

	go network.NetworkCum(out, elevatorin, statusin)
	go fsm.RunLocalElevator(local)
	//go esm.RunESM(local, in, out)

}
