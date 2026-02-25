package main

import (
	network "project/Network"
	"project/Network/peers"
	"project/constant"
	"project/elevio"
	"project/fsm"
	"project/types"
)

func main() {

	numFloors := constant.NumFloors
	//numElevs := constant.NumElevators

	elevio.Init("Localhost:15657", numFloors)

	out := make(chan types.ExternalElevator)
	elevatorin := make(chan network.Msg)
	statusin := make(chan peers.PeerUpdate)
	Local := make(chan types.Elevator)

	go network.NetworkCum(out, elevatorin, statusin)
	go fsm.RunLocalElevator(Local)
	//go esm.RunESM(Local, in, out)

}
