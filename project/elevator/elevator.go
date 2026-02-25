package elevator

import (
	"fmt"
	"project/constant"
	"project/types"
)

const numFloors int = constant.NumFloors
const numButtons int = constant.NumButtons

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

func DirnToString(dirn types.MotorDirection) string {
	switch dirn {
	case types.MD_Down:
		return "Down"
	case types.MD_Up:
		return "Up"
	case types.MD_Stop:
		return "Stop"
	default:
		return "undefined direction"
	}
}

func ButtonToString(button types.ButtonType) string {
	switch button {
	case types.BT_HallUp:
		return "Hall Up"
	case types.BT_HallDown:
		return "Hall Down"
	case types.BT_Cab:
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
			if (f == constant.NumFloors-1 && btn == int(types.BT_HallUp)) ||
				(f == 0 && btn == int(types.BT_HallDown)) {
				fmt.Print("|     ")
				continue
			}

			switch e.Requests[f][btn] {
			case types.ReqNone:
				fmt.Print("|  -  ")
			case types.ReqUnconfirmed:
				fmt.Print("|  ?  ")
			case types.ReqConfirmed:
				fmt.Print("|  #  ")
			case types.ReqDeleting:
				fmt.Print("|  x  ")
			default:
				fmt.Print("|  !  ") // ukjent verdi -> bug
			}
		}
		fmt.Println("|")
	}

	fmt.Println("  +--------------------+")
}
