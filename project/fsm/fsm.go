package fsm

import (
	"fmt"
	"project/elevator"
	"project/elevio"
	"project/request"
	"project/timer"
)

func SetAllLights(es Elevator) {
	for floor := 0; floor < numFloors; floor++ {
		for button := 0; button < elevio._numButtons; button++ {
			elevio.SetButtonLamp(elevio.ButtonType(button), floor, es.Requests[floor][button])
		}
	}
}

func OnInitBetweenFloors(e *Elevator) {
	elevio.SetMotorDirection(D_Down)
	e.Dirn = D_Down
	e.Behavior = EB_moving
}

func OnRequestButtonPress(e *Elevator, floor int, btnType Button) {
	fmt.Printf("\n\nFSMOnRequestButtonPress(%d, %s)\n", floor, elevator.ButtonToString(btnType))
	elevator.ElevatorPrint(*e)

	switch e.Behaviour {
	case EB_DoorOpen:
		if request.ShouldClearImmediately(*e, floor, btnType) {
			timer.Start(e.Config.DoorOpenDurationSec)
		} else {
			e.Requests[floor][btnType] = 1
		}

	case EB_Moving:
		e.Requests[floor][btnType] = 1

	case EB_Idle:
		e.Requests[floor][btnType] = 1
		pair := request.ChooseDirection(*e)
		e.Dirn = pair.Dirn
		e.Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case EB_DoorOpen:
			elevio.SetDoorOpenLamp(1)
			timer.Start(e.Config.DoorOpenDurationSec)
			*e = request.ClearAtCurrentFloor(*e)
		case EB_Moving:
			elevio.SetMotorDirection(e.Dirn)
		case EB_Idle:
			// nothing to do
		}
	}

	SetAllLights(*e)
	fmt.Println("\nNew state:")
	elevator.ElevatorPrint(*e)
}

func OnFloorArrival(e *Elevator, newFloor int) {
	fmt.Printf("\n\nFSMOnFloorArrival(%d)\n", newFloor)
	elevator.ElevatorPrint(*e)

	e.Floor = newFloor
	elevio.SetFloorIndicator(e.Floor)

	switch e.Behaviour {
	case EB_Moving:
		if request.ShouldStop(*e) {
			elevio.SetMotorDirection(D_Stop)

			elevio.SetDoorOpenLamp(1)

			*e = request.ClearAtCurrentFloor(*e)

			timer.Start(e.Config.DoorOpenDurationSec)

			SetAllLights(*e)

			e.Behaviour = EB_DoorOpen
		}
	default:
		// Do nothing for other behaviours
	}

	fmt.Println("\nNew state:")
	elevator.ElevatorPrint(*e)
}

func FSMOnDoorTimeout(e *Elevator) {
	fmt.Println("\n\nFSMOnDoorTimeout()")
	elevator.ElevatorPrint(*e)

	switch e.Behaviour {
	case EB_DoorOpen:
		pair := request.ChooseDirection(*e)
		e.Dirn = pair.Dirn
		e.Behaviour = pair.Behaviour

		switch e.Behaviour {
		case EB_DoorOpen:
			timer.Start(e.Config.DoorOpenDurationSec)
			*e = request.ClearAtCurrentFloor(*e)
			SetAllLights(*e)
		case EB_Moving, EB_Idle:
			elevio.SetDoorOpenLamp(0)
			elevio.SetMotorDirection(e.Dirn)
		}
	default:
		// Do nothing for other behaviours
	}

	fmt.Println("\nNew state:")
	elevator.ElevatorPrint(*e)
}
