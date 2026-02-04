package elevator

const (
	numFloors  = 4
	numButtons = 3
)

type Dirn int

const (
	D_Down Dirn = iota - 1
	D_Stop
	D_Up
)

type Button int

const (
	B_HallUp Button = iota
	B_HallDown
	B_Cab
)

type ElevatorBehavior int

const (
	EB_Idle ElevatorBehavior = iota
	EB_DoorOpen
	EB_Moving
)

type Elevator struct {
	floor    int
	dirn     Dirn
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

func elevator_dirnToString(dirn Dirn) string {
	switch dirn {
	case D_Down:
		return "Down"
	case D_Up:
		return "Up"
	case D_Stop:
		return "Stop"
	default:
		return "undefined direction"
	}
}
