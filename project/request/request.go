package request

import (
	"project/constant"
	"project/types"
)

type DirnBehaviourPair struct {
	Dirn      types.MotorDirection
	Behaviour types.ElevatorBehavior
}

func ReqIsActive(s types.ReqState) bool {
	return s == types.ReqConfirmed
}

func Above(e types.Elevator) bool {
	for f := e.Floor + 1; f < constant.NumFloors; f++ {

		for b := types.ButtonType(0); b < constant.NumButtons; b++ {
			if ReqIsActive(e.Requests[f][b]) {
				return true
			}
		}
	}
	return false
}

func Below(e types.Elevator) bool {
	for f := 0; f < e.Floor; f++ {

		for b := types.ButtonType(0); b < constant.NumButtons; b++ {
			if ReqIsActive(e.Requests[f][b]) {

				return true
			}
		}
	}
	return false
}

func Here(e types.Elevator) bool {
	for b := types.ButtonType(0); b < 3; b++ {
		if ReqIsActive(e.Requests[e.Floor][b]) {
			return true
		}
	}
	return false
}

func ChooseDirection(e types.Elevator) DirnBehaviourPair {
	switch e.Dirn {

	case types.MD_Up:
		if Above(e) {
			return DirnBehaviourPair{types.MD_Up, types.EB_Moving}
		} else if Here(e) {
			return DirnBehaviourPair{types.MD_Up, types.EB_DoorOpen}
		} else if Below(e) {
			return DirnBehaviourPair{types.MD_Down, types.EB_Moving}
		}
		return DirnBehaviourPair{types.MD_Stop, types.EB_Idle}

	case types.MD_Down:
		if Below(e) {
			return DirnBehaviourPair{types.MD_Down, types.EB_Moving}
		} else if Here(e) {
			return DirnBehaviourPair{types.MD_Down, types.EB_DoorOpen}
		} else if Above(e) {
			return DirnBehaviourPair{types.MD_Up, types.EB_Moving}
		}
		return DirnBehaviourPair{types.MD_Stop, types.EB_Idle}

	case types.MD_Stop:
		if Here(e) {
			return DirnBehaviourPair{types.MD_Stop, types.EB_DoorOpen}
		} else if Above(e) {
			return DirnBehaviourPair{types.MD_Up, types.EB_Moving}
		} else if Below(e) {
			return DirnBehaviourPair{types.MD_Down, types.EB_Moving}
		}
		return DirnBehaviourPair{types.MD_Stop, types.EB_Idle}

	default:
		return DirnBehaviourPair{types.MD_Stop, types.EB_Idle}
	}
}

func ShouldStop(e types.Elevator) bool {
	switch e.Dirn {

	case types.MD_Up:
		return ReqIsActive(e.Requests[e.Floor][types.BT_HallUp]) ||
			ReqIsActive(e.Requests[e.Floor][types.BT_Cab]) ||
			!Above(e)

	case types.MD_Down:
		return ReqIsActive(e.Requests[e.Floor][types.BT_HallDown]) ||
			ReqIsActive(e.Requests[e.Floor][types.BT_Cab]) ||
			!Below(e)

	case types.MD_Stop:
		return true

	default:
		return true
	}
}

func ShouldClearImmediately(e types.Elevator, btnFloor int, btnType types.ButtonType) bool {
	return e.Floor == btnFloor &&
		((e.Dirn == types.MD_Up && btnType == types.BT_HallUp) ||
			(e.Dirn == types.MD_Down && btnType == types.BT_HallDown) ||
			e.Dirn == types.MD_Stop ||
			btnType == types.BT_Cab)
}

func ClearAtCurrentFloor(e types.Elevator) types.Elevator {

	e.Requests[e.Floor][types.BT_Cab] = 0

	switch e.Dirn {

	case types.MD_Up:
		if !Above(e) && !ReqIsActive(e.Requests[e.Floor][types.BT_HallUp]) {
			e.Requests[e.Floor][types.BT_HallDown] = types.ReqDeleting
		}
		e.Requests[e.Floor][types.BT_HallUp] = types.ReqDeleting

	case types.MD_Down:
		if !Below(e) && !ReqIsActive(e.Requests[e.Floor][types.BT_HallDown]) {
			e.Requests[e.Floor][types.BT_HallUp] = types.ReqDeleting
		}
		e.Requests[e.Floor][types.BT_HallDown] = types.ReqDeleting

	case types.MD_Stop:
		fallthrough
	default:
		if ReqIsActive(e.Requests[e.Floor][types.BT_HallUp]) {
			e.Requests[e.Floor][types.BT_HallUp] = types.ReqDeleting
		}
		if ReqIsActive(e.Requests[e.Floor][types.BT_HallDown]) {
			e.Requests[e.Floor][types.BT_HallDown] = types.ReqDeleting
		}
	}

	return e
}

func ClearAtCurrentFloorWithCallback(e types.Elevator, onClear func(btn types.ButtonType)) types.Elevator {

	// Helper: transition active -> deleting (and call onClear once)
	markDeleting := func(btn types.ButtonType) {
		if ReqIsActive(e.Requests[e.Floor][btn]) {
			onClear(btn)
			e.Requests[e.Floor][btn] = types.ReqDeleting
		}
	}

	// Cab: clear (mark deleting) if active
	markDeleting(types.BT_Cab)

	switch e.Dirn {

	case types.MD_Up:
		// If no requests above AND hallUp not active at this floor, clear hallDown too
		if !Above(e) && !ReqIsActive(e.Requests[e.Floor][types.BT_HallUp]) {
			markDeleting(types.BT_HallDown)
		}
		// Always clear hallUp in up-direction if active
		markDeleting(types.BT_HallUp)

	case types.MD_Down:
		// If no requests below AND hallDown not active at this floor, clear hallUp too
		if !Below(e) && !ReqIsActive(e.Requests[e.Floor][types.BT_HallDown]) {
			markDeleting(types.BT_HallUp)
		}
		// Always clear hallDown in down-direction if active
		markDeleting(types.BT_HallDown)

	case types.MD_Stop:
		fallthrough
	default:
		// Clear both hall buttons if active
		markDeleting(types.BT_HallUp)
		markDeleting(types.BT_HallDown)
	}

	return e
}
