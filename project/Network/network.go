package network

import (
	"project/Network/Transform_elevator"
	"project/esm"
)

func NetworkCum(in <-chan esm.ExternalElevator, out chan<- Transform_elevator.ElevatorMsg) {
	a := Transform_elevator.Set_up1(<-in)
	out <- a
}
