package types

import (
	"project/constant"
	"project/elevio"
)

type ElevatorBehavior int

const (
	EB_Idle ElevatorBehavior = iota
	EB_DoorOpen
	EB_Moving
)

type ReqState int

const (
	ReqNone        ReqState = 0
	ReqUnconfirmed ReqState = 1
	ReqConfirmed   ReqState = 2
	ReqDeleting    ReqState = 3
)

type ExternalElevator struct {
	Status   bool
	Elevator Elevator
}

type WorldView struct {
	Elevators       map[string]ExternalElevator
	OnlineElevators int
	Local           Elevator

	AssignedLocal [constant.NumFloors][2]bool // Vet ikke om vi må ha denne
}

type Elevator struct {
	Floor     int
	Dirn      elevio.MotorDirection
	Requests  [constant.NumFloors][constant.NumButtons]ReqState
	Behaviour ElevatorBehavior
}
