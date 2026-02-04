package request

import "Driver-go/elevio"

func requestAbove(e Elevator) bool {
	for f := e.Floor + 1; f < numFloors; f++ {
		for b := elevio.ButtonType(0); b < 4; b++ {
			if e.request(f, b) {
				return true
			}
		}
	}
	return false
}

func requestBelow(e Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ {
			if e.request(f, b) {
				return true
			}
		}
	}
	return false
}

func requestHere(e Elevator) bool {
	for b := elevio.ButtonType(0); b < 3; b++ {
		if e.request(e.Floor, b) {
			return true
		}
	}
	return false
}

func request_chooseDirection(e Elevator) {

	select {}
}
