package request

import "project/elevio"
import "project/elevator"

type DirnBehaviourPair struct {
	Dirn      elevio.MotorDirection
	Behaviour elevator.ElevatorBehavior
}

func Above(e elevator.Elevator) bool {
	for f := e.Floor + 1; f < elevator.numFloors; f++ {
		for b := elevio.ButtonType(0); b < 4; b++ {
			if e.Requests[f][b]{
				return true 
			}
		}
	}
	return false
}

func Below(e elevator.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ {
			if e.Requests[f][b]{
				return true
			}
		}
	}
	return false
}

func Here(e elevator.Elevator) bool {
	for b := elevio.ButtonType(0); b < 3; b++ {
		if e.Requests[e.Floor][b]{
			return true
		}
	}
	return false
}
func ChooseDirection(e elevator.Elevator) DirnBehaviourPair {
	switch e.Dirn {
	case elevio.MD_Up:
		if Above(e) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		} else if Here(e) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_DoorOpen}
		} else if Below(e) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		}
		return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}

	case elevio.MD_Down:
		if Below(e) {
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

func ShouldStop(e Elevator) bool {
	switch e.Dirn {
	case D_up:
		return e.request[e.floor][B_HallUp] ||
			e.request[e.floor][B_Cab] ||
			!requestAbove(e)

	case D_Down:
		return e.request[e.floor][B_HallDown] ||
			e.request[e.floor][B_Cab] ||
			!requestBelow(e)

	case D_Stop:
		return true

	default:
		return true

	}
}

func ShouldClearImmediately(e Elevator, btnFloor int, btnType Button) bool {
	return e.Floor == btnFloor &&
		((e.Dirn == D_Up && btnType == B_HallUp) ||
			(e.Dirn == D_Down && btnType == B_HallDown) ||
			e.Dirn == D_Stop ||
			btnType == B_Cab)
}

func ClearAtCurrentFloor(e Elevator) Elevator {

	e.Requests[e.Floor][B_Cab] = false

	switch e.Dirn {

	case D_Up:
		if !requestsAbove(e) && !e.Requests[e.Floor][B_HallUp] {
			e.Requests[e.Floor][B_HallDown] = false
		}

		e.Requests[e.Floor][B_HallUp] = false

	case D_Down:

		if !requestsBelow(e) && !e.Requests[e.Floor][B_HallDown] {
			e.Requests[e.Floor][B_HallUp] = false
		}
		e.Requests[e.Floor][B_HallDown] = false

	case D_Stop:
		fallthrough
	default:
		e.Requests[e.Floor][B_HallUp] = false
		e.Requests[e.Floor][B_HallDown] = false
	}

	return e
}
