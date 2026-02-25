package elevator

import (
	"fmt"
	"project/constant"
	"project/elevio"
	"project/types"
)

const numFloors int = constant.NumFloors
const numButtons int = constant.NumButtons

func ReqIsActive(s types.ReqState) bool {
	return s == types.ReqConfirmed
}

func BehaviorToString(eb types.ElevatorBehavior) string {
	switch eb {
	case types.EB_Idle:
		return "EB_idle"
	case types.EB_DoorOpen:
		return "DoorOpen"
	case types.EB_Moving:
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

func ElevatorPrint(e types.Elevator) {
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
