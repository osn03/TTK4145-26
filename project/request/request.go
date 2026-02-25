package request

import (
	"project/constant"
	"project/elevator"
	"project/elevio"
	"project/types"
)

type DirnBehaviourPair struct {
	Dirn      elevio.MotorDirection
	Behaviour types.ElevatorBehavior
}

func Above(e types.Elevator) bool {
	for f := e.Floor + 1; f < constant.NumFloors; f++ {

		for b := elevio.ButtonType(0); b < constant.NumButtons; b++ {
			if elevator.ReqIsActive(e.Requests[f][b]) {
				return true
			}
		}
	}
	return false
}

func Below(e types.Elevator) bool {
	for f := 0; f < e.Floor; f++ {

		for b := elevio.ButtonType(0); b < constant.NumButtons; b++ {
			if elevator.ReqIsActive(e.Requests[f][b]) {

				return true
			}
		}
	}
	return false
}

func Here(e types.Elevator) bool {
	for b := elevio.ButtonType(0); b < 3; b++ {
		if elevator.ReqIsActive(e.Requests[e.Floor][b]) {
			return true
		}
	}
	return false
}

func ChooseDirection(e types.Elevator) DirnBehaviourPair {
	switch e.Dirn {

	case elevio.MD_Up:
		if Above(e) {
			return DirnBehaviourPair{elevio.MD_Up, types.EB_Moving}
		} else if Here(e) {
			return DirnBehaviourPair{elevio.MD_Up, types.EB_DoorOpen}
		} else if Below(e) {
			return DirnBehaviourPair{elevio.MD_Down, types.EB_Moving}
		}
		return DirnBehaviourPair{elevio.MD_Stop, types.EB_Idle}

	case elevio.MD_Down:
		if Below(e) {
			return DirnBehaviourPair{elevio.MD_Down, types.EB_Moving}
		} else if Here(e) {
			return DirnBehaviourPair{elevio.MD_Down, types.EB_DoorOpen}
		} else if Above(e) {
			return DirnBehaviourPair{elevio.MD_Up, types.EB_Moving}
		}
		return DirnBehaviourPair{elevio.MD_Stop, types.EB_Idle}

	case elevio.MD_Stop:
		if Here(e) {
			return DirnBehaviourPair{elevio.MD_Stop, types.EB_DoorOpen}
		} else if Above(e) {
			return DirnBehaviourPair{elevio.MD_Up, types.EB_Moving}
		} else if Below(e) {
			return DirnBehaviourPair{elevio.MD_Down, types.EB_Moving}
		}
		return DirnBehaviourPair{elevio.MD_Stop, types.EB_Idle}

	default:
		return DirnBehaviourPair{elevio.MD_Stop, types.EB_Idle}
	}
}

func ShouldStop(e types.Elevator) bool {
	switch e.Dirn {

	case elevio.MD_Up:
		return elevator.ReqIsActive(e.Requests[e.Floor][elevio.BT_HallUp]) ||
			elevator.ReqIsActive(e.Requests[e.Floor][elevio.BT_Cab]) ||
			!Above(e)

	case elevio.MD_Down:
		return elevator.ReqIsActive(e.Requests[e.Floor][elevio.BT_HallDown]) ||
			elevator.ReqIsActive(e.Requests[e.Floor][elevio.BT_Cab]) ||
			!Below(e)

	case elevio.MD_Stop:
		return true

	default:
		return true
	}
}

func ShouldClearImmediately(e types.Elevator, btnFloor int, btnType elevio.ButtonType) bool {
	return e.Floor == btnFloor &&
		((e.Dirn == elevio.MD_Up && btnType == elevio.BT_HallUp) ||
			(e.Dirn == elevio.MD_Down && btnType == elevio.BT_HallDown) ||
			e.Dirn == elevio.MD_Stop ||
			btnType == elevio.BT_Cab)
}

func ClearAtCurrentFloor(e types.Elevator) types.Elevator {

	e.Requests[e.Floor][elevio.BT_Cab] = 0

	switch e.Dirn {

	case elevio.MD_Up:
		if !Above(e) && !elevator.ReqIsActive(e.Requests[e.Floor][elevio.BT_HallUp]) {
			e.Requests[e.Floor][elevio.BT_HallDown] = types.ReqDeleting
		}
		e.Requests[e.Floor][elevio.BT_HallUp] = types.ReqDeleting

	case elevio.MD_Down:
		if !Below(e) && !elevator.ReqIsActive(e.Requests[e.Floor][elevio.BT_HallDown]) {
			e.Requests[e.Floor][elevio.BT_HallUp] = types.ReqDeleting
		}
		e.Requests[e.Floor][elevio.BT_HallDown] = types.ReqDeleting

	case elevio.MD_Stop:
		fallthrough
	default:
		e.Requests[e.Floor][elevio.BT_HallUp] = types.ReqDeleting
		e.Requests[e.Floor][elevio.BT_HallDown] = types.ReqDeleting
	}

	return e
}

func ClearAtCurrentFloorWithCallback(e types.Elevator, onClear func(btn elevio.ButtonType)) types.Elevator {

	// Helper: transition active -> deleting (and call onClear once)
	markDeleting := func(btn elevio.ButtonType) {
		if elevator.ReqIsActive(e.Requests[e.Floor][btn]) {
			onClear(btn)
			e.Requests[e.Floor][btn] = types.ReqDeleting
		}
	}

	// Cab: clear (mark deleting) if active
	markDeleting(elevio.BT_Cab)

	switch e.Dirn {

	case elevio.MD_Up:
		// If no requests above AND hallUp not active at this floor, clear hallDown too
		if !Above(e) && !elevator.ReqIsActive(e.Requests[e.Floor][elevio.BT_HallUp]) {
			markDeleting(elevio.BT_HallDown)
		}
		// Always clear hallUp in up-direction if active
		markDeleting(elevio.BT_HallUp)

	case elevio.MD_Down:
		// If no requests below AND hallDown not active at this floor, clear hallUp too
		if !Below(e) && !elevator.ReqIsActive(e.Requests[e.Floor][elevio.BT_HallDown]) {
			markDeleting(elevio.BT_HallUp)
		}
		// Always clear hallDown in down-direction if active
		markDeleting(elevio.BT_HallDown)

	case elevio.MD_Stop:
		fallthrough
	default:
		// Clear both hall buttons if active
		markDeleting(elevio.BT_HallUp)
		markDeleting(elevio.BT_HallDown)
	}

	return e
}
