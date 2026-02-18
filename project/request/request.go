package request

import (
	"project/constant"
	"project/elevator"
	"project/elevio"
)

type DirnBehaviourPair struct {
	Dirn      elevio.MotorDirection
	Behaviour elevator.ElevatorBehavior
}

func Above(e elevator.Elevator) bool {
	for f := e.Floor + 1; f < constant.NumFloors; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ {
			if e.Requests[f][b] {
				return true
			}
		}
	}
	return false
}

func Below(e elevator.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ {
			if e.Requests[f][b] {
				return true
			}
		}
	}
	return false
}

func Here(e elevator.Elevator) bool {
	for b := elevio.ButtonType(0); b < 3; b++ {
		if e.Requests[e.Floor][b] {
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
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		} else if Here(e) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_DoorOpen}
		} else if Above(e) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		}
		return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}

	case elevio.MD_Stop:
		if Here(e) {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_DoorOpen}
		} else if Above(e) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		} else if Below(e) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		}
		return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}

	default:
		return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
	}
}

func ShouldStop(e elevator.Elevator) bool {
	switch e.Dirn {

	case elevio.MD_Up:
		return e.Requests[e.Floor][elevio.BT_HallUp] ||
			e.Requests[e.Floor][elevio.BT_Cab] ||
			!Above(e)

	case elevio.MD_Down:
		return e.Requests[e.Floor][elevio.BT_HallDown] ||
			e.Requests[e.Floor][elevio.BT_Cab] ||
			!Below(e)

	case elevio.MD_Stop:
		return true

	default:
		return true
	}
}

func ShouldClearImmediately(e elevator.Elevator, btnFloor int, btnType elevio.ButtonType) bool {
	return e.Floor == btnFloor &&
		((e.Dirn == elevio.MD_Up && btnType == elevio.BT_HallUp) ||
			(e.Dirn == elevio.MD_Down && btnType == elevio.BT_HallDown) ||
			e.Dirn == elevio.MD_Stop ||
			btnType == elevio.BT_Cab)
}

func ClearAtCurrentFloor(e elevator.Elevator) elevator.Elevator {

	e.Requests[e.Floor][elevio.BT_Cab] = false

	switch e.Dirn {

	case elevio.MD_Up:
		if !Above(e) && !e.Requests[e.Floor][elevio.BT_HallUp] {
			e.Requests[e.Floor][elevio.BT_HallDown] = false
		}
		e.Requests[e.Floor][elevio.BT_HallUp] = false

	case elevio.MD_Down:
		if !Below(e) && !e.Requests[e.Floor][elevio.BT_HallDown] {
			e.Requests[e.Floor][elevio.BT_HallUp] = false
		}
		e.Requests[e.Floor][elevio.BT_HallDown] = false

	case elevio.MD_Stop:
		fallthrough
	default:
		e.Requests[e.Floor][elevio.BT_HallUp] = false
		e.Requests[e.Floor][elevio.BT_HallDown] = false
	}

	return e
}
