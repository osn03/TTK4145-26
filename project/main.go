package main

import (
	network "project/Network"
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
	in := make(chan network.Msg)
	local := make(chan elevator.Elevator)

	go network.NetworkCum(in, out)
	go fsm.RunLocalElevator(local)
	//go esm.RunESM(local, in, out)

}
