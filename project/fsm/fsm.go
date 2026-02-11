package fsm

import (
	"fmt"
	"project/constant"
	"project/elevator"
	"project/elevio"
	"project/request"
	"project/timer"
)

func SetAllLights(e elevator.Elevator) {
	for floor := 0; floor < constant.NumFloors; floor++ {
		for button := 0; button < constant.NumButtons; button++ {
			elevio.SetButtonLamp(elevio.ButtonType(button), floor, e.Requests[floor][button])
		}
	}
}

func OnInitBetweenFloors(e *elevator.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	e.Dirn = elevio.MD_Down
	e.Behaviour = elevator.EB_Moving
}

func OnRequestButtonPress(e *elevator.Elevator, floor int, btnType elevio.ButtonType) {
	fmt.Printf("\n\nFSMOnRequestButtonPress(%d, %s)\n", floor, elevator.ButtonToString(btnType))
	elevator.ElevatorPrint(*e)

	switch e.Behaviour {
	case elevator.EB_DoorOpen:
		if request.ShouldClearImmediately(*e, floor, btnType) {
			timer.Start(constant.DoorOpenDurationSec)
		} else {
			e.Requests[floor][btnType] = true
		}

	case elevator.EB_Moving:
		e.Requests[floor][btnType] = true

	case elevator.EB_Idle:
		e.Requests[floor][btnType] = true
		pair := request.ChooseDirection(*e)
		e.Dirn = pair.Dirn
		e.Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case elevator.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.Start(constant.DoorOpenDurationSec)
			*e = request.ClearAtCurrentFloor(*e)
		case elevator.EB_Moving:
			elevio.SetMotorDirection(e.Dirn)
		case elevator.EB_Idle:
			// nothing to do
		}
	}

	SetAllLights(*e)
	fmt.Println("\nNew state:")
	elevator.ElevatorPrint(*e)
}

func OnFloorArrival(e *elevator.Elevator, newFloor int) {
	fmt.Printf("\n\nFSMOnFloorArrival(%d)\n", newFloor)
	elevator.ElevatorPrint(*e)

	e.Floor = newFloor
	elevio.SetFloorIndicator(e.Floor)

	switch e.Behaviour {
	case elevator.EB_Moving:
		if request.ShouldStop(*e) {
			elevio.SetMotorDirection(elevio.MD_Stop)

			elevio.SetDoorOpenLamp(true)

			*e = request.ClearAtCurrentFloor(*e)

			timer.Start(constant.DoorOpenDurationSec)

			SetAllLights(*e)

			e.Behaviour = elevator.EB_DoorOpen
		}
	default:
		// Do nothing for other behaviours
	}

	fmt.Println("\nNew state:")
	elevator.ElevatorPrint(*e)
}

func OnDoorTimeout(e *elevator.Elevator) {
	fmt.Println("\n\nFSMOnDoorTimeout()")
	elevator.ElevatorPrint(*e)
	
	switch e.Behaviour {
	case elevator.EB_DoorOpen:
		pair := request.ChooseDirection(*e)
		e.Dirn = pair.Dirn
		e.Behaviour = pair.Behaviour

		switch e.Behaviour {
		case elevator.EB_DoorOpen:
			timer.Start(constant.DoorOpenDurationSec)
			*e = request.ClearAtCurrentFloor(*e)
			SetAllLights(*e)
		case elevator.EB_Moving, elevator.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(e.Dirn)
		}
	default:
		// Do nothing for other behaviours
	}

	fmt.Println("\nNew state:")
	elevator.ElevatorPrint(*e)
}
