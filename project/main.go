package main

import (
	"project/constant"
	"project/elevio"
	"project/esm"
	"project/fsm"
	"project/network"
	"project/network/peers"
	"project/types"
)

func main() {

	numFloors := constant.NumFloors
	//numElevs := constant.NumElevators
	localid := "150"

	elevio.Init("localhost:15657", numFloors)

	outNetwork := make(chan types.ExternalElevator)
	elevatorin := make(chan network.Msg)
	statusin := make(chan peers.PeerUpdate)
	fromHardware := make(chan types.Elevator)
	toHardware := make(chan [4][3]types.ReqState)

	go network.NetworkCum(outNetwork, elevatorin, statusin, localid)
	go fsm.RunLocalElevator(fromHardware, toHardware)
	esm.RunESM(fromHardware, elevatorin, outNetwork, statusin, localid, toHardware)
}
