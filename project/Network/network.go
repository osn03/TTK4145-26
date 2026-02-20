package network

import (
	"project/Network/Transform_elevator"
	"project/esm"
)

func NetworkCum(e *esm.ExternalElevator, out chan<- Transform_elevator.ElevatorMsg) {
	a := Transform_elevator.Set_up1(e)
	out <- a
}
