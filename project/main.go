package main

import (
	"project/constant"
	"project/elevator"
	"project/elevio"
	"project/fsm"
)

func main() {

	numFloors := constant.NumFloors
	//numElevs := constant.NumElevators

	elevio.Init("localhost:15657", numFloors)

	var e elevator.Elevator
	fsm.OnInitBetweenFloors(&e)

	fsm.RunLocalElevator(&e)

}
