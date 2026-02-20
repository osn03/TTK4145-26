package network

import (
	"project/Network/Transform_elevator"
	"project/esm"
)

func NetworkCum(e *esm.ExternalElevator) (esm.ExternalElevator, string) {
	return Transform_elevator.Set_up1(e)
}
