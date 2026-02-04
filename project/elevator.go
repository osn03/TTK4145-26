package elevator

import (
	"elevio"
	"fmt"
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
	behaviour ElevatorBehavior
}

func behaviorToString(eb ElevatorBehavior) string {
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

func dirnToString(dirn elevio.MotorDirection) string {
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

func buttonToString(button elevio.ButtonType) string {
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


func elevator_print(el Elevator){
	fmt.Println("  +--------------------+")
	fmt.Printf(
		"  |floor = %-2d          |\n"+
			"  |dirn  = %-12.12s|\n"+
			"  |behav = %-12.12s|\n",
		el.floor,
		dirnToString(el.dirn),
		behaviorToString(el.behaviour),
	)
	fmt.Println("  +--------------------+")
	fmt.Println("  |  | up  | dn  | cab |")

	for f := numFloors - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < numButtons; btn++ {
			if (f == numFloors-1 && btn == int(elevio.BT_HallUp)) ||
				(f == 0 && btn == int(elevio.BT_HallDown)) {
				fmt.Print("|     ")
			} else {
				if el.requests[f][btn] != 0 {
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