package elevator

import (
	"fmt"
	"project/elevio"
	"project/constant"
)

const numFloors int = constant.NumFloors
const numButtons int = constant.NumButtons


type ElevatorBehavior int
type ReqState int

const (
	ReqNone        ReqState = 0
	ReqUnconfirmed ReqState = 1
	ReqConfirmed   ReqState = 2
	ReqDeleting    ReqState = 3
)

func ReqIsActive(s ReqState) bool {
	return s == ReqUnconfirmed || s == ReqConfirmed
}

func ReqLampOn(s ReqState) bool {
	return s == ReqConfirmed
}

const (
	EB_Idle ElevatorBehavior = iota
	EB_DoorOpen
	EB_Moving
)

type Elevator struct {
	Floor    int
	Dirn     elevio.MotorDirection
	Requests [numFloors][numButtons]ReqState
	Behaviour ElevatorBehavior
}

func BehaviorToString(eb ElevatorBehavior) string {
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

func DirnToString(dirn elevio.MotorDirection) string {
	switch dirn {
	case elevio.MD_Down:
		return "Down"
	case elevio.MD_Up:
		return "Up"
	case elevio.MD_Stop:
		return "Stop"
	default:
		return "undefined direction"
	}
}

func ButtonToString(button elevio.ButtonType) string {
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


func ElevatorPrint(e Elevator){
	fmt.Println("  +--------------------+")
	fmt.Printf(
		"  |floor = %-2d          |\n"+
			"  |dirn  = %-12.12s|\n"+
			"  |behav = %-12.12s|\n",
		e.Floor,
		DirnToString(e.Dirn),
		BehaviorToString(e.Behaviour),
	)
	fmt.Println("  +--------------------+")
	fmt.Println("  |  | up  | dn  | cab |")

	for f := constant.NumFloors - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < constant.NumButtons; btn++ {
			if (f == constant.NumFloors-1 && btn == int(elevio.BT_HallUp)) ||
				(f == 0 && btn == int(elevio.BT_HallDown)) {
				fmt.Print("|     ")
			} else {
				if e.Requests[f][btn] > 0 {
					fmt.Print("|  #  ")
				} else {
					fmt.Print("|  -  ")
				}
			}
		}
		fmt.Println("|")
	}

	fmt.Println("  +--------------------+")
}