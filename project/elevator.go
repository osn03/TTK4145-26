package elevator

import (
	"elevio"
)

const (
	numFloors  = 4
	numButtons = 3
)

type ElevatorBehavior int

const (
	EB_Idle ElevatorBehavior = iota
	EB_DoorOpen
	EB_Moving
)

type Elevator struct {
	floor    int
	dirn     elevio.MotorDirection
	requests [numFloors][numButtons]int
}

func elevator_behaviorToString(eb ElevatorBehavior) string {
	switch eb {
	case EB_Idle:
		return "EB_idle"
	case EB_DoorOpen:
		return "DoorOpen"
	case EB_Moving:
		return "Moving"
	default:
		return "undefined behavior"
	}
}

func elevator_dirnToString(dirn elevio.MotorDirection) string {
	switch dirn {
	case elevio.D_Down:
		return "Down"
	case elevio.D_Up:
		return "Up"
	case elevio.D_Stop:
		return "Stop"
	default:
		return "undefined direction"
	}
}

func elevator_buttonToString(button elevio.ButtonType) string {
	switch button {
	case elevio.BT_HallUp:
		return "Hall Up"
	case elevio.BT_HallDown:
		return "Hall Down"
	case elevio.BT_Cab:
		return "Cab"
	default:
		return "undefined button"
	}
}
