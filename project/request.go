package request

import "Driver-go/elevio"

type DirnBehaviourPair struct {
	Dirn      Dirn
	Behaviour ElevatorBehaviour
}

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
func requestChooseDirection(e Elevator) DirnBehaviourPair {
	switch e.Dirn {
	case D_up:
		if requestAbove(e) {
			return DirnBehaviourPair{D_up, EB_moving}
		} else if requestHere(e) {
			return DirnBehaviourPair{D_down, EB_doorOpen}
		} else if requestBelow(e) {
			return DirnBehaviourPair{D_down, EB_moving}
		}
		return DirnBehaviourPair{D_stop, EB_idle}

	case D_down:
		if requestBelow(e) {
			return DirnBehaviourPair{D_down, EB_moving}
		} else if requestHere(e) {

			return DirnBehaviourPair{D_up, EB_doorOpen}
		} else if requestAbove(e) {
			return DirnBehaviourPair{D_up, EB_moving}
		}
		return DirnBehaviourPair{D_stop, EB_idle}

	case D_stop:
		if requestHere(e) {
			return DirnBehaviourPair{D_stop, EB_doorOpen}
		} else if requestAbove(e) {
			return DirnBehaviourPair{D_up, EB_moving}
		} else if requestBelow(e) {
			return DirnBehaviourPair{D_down, EB_moving}
		}
		return DirnBehaviourPair{D_stop, EB_idle}

	default:
		return DirnBehaviourPair{D_stop, EB_idle}
	}
}
