package fsm

import (
	"Driver-go/elevio"
	"elevator"
	"ftm"
	"time"
)

func SetAllLights(es Elevator){
	for floor := 0; floor < numFloors;floor++{
		for button := 0; button < elevio._numButtons; button++{
			elevio.SetButtonLamp(elevio.ButtonType(button), floor, es.Requests[floor][button])
		}	
	}
}

func fsm_onInitBetweenFloors(e *Elevator) {
	elevator_motorDirection(D_Down)
	e.Dirn = D_Down
	e.Behavior = EB_moving
}

func FSMOnRequestButtonPress(e *Elevator, floor int, btnType Button) {
	fmt.Printf("\n\nFSMOnRequestButtonPress(%d, %s)\n", floor, ElevatorButtonToString(btnType))
	ElevatorPrint(*e)

	switch e.Behaviour {
	case EB_DoorOpen:
		if RequestsShouldClearImmediately(*e, floor, btnType) {
			TimerStart(e.Config.DoorOpenDurationSec)
		} else {
			e.Requests[floor][btnType] = 1
		}

	case EB_Moving:
		e.Requests[floor][btnType] = 1

	case EB_Idle:
		e.Requests[floor][btnType] = 1
		pair := RequestsChooseDirection(*e)
		e.Dirn = pair.Dirn
		e.Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case EB_DoorOpen:
			ElevatorDoorLight(1)
			TimerStart(e.Config.DoorOpenDurationSec)
			*e = RequestsClearAtCurrentFloor(*e)
		case EB_Moving:
			ElevatorMotorDirection(e.Dirn)
		case EB_Idle:
			// nothing to do
		}
	}

	SetAllLights(*e)
	fmt.Println("\nNew state:")
	ElevatorPrint(*e)
}

func FSMOnFloorArrival(e *Elevator, newFloor int) {
	fmt.Printf("\n\nFSMOnFloorArrival(%d)\n", newFloor)
	ElevatorPrint(*e)

	
	e.Floor = newFloor
	ElevatorFloorIndicator(e.Floor)

	switch e.Behaviour {
	case EB_Moving:
		if RequestsShouldStop(*e) {
			ElevatorMotorDirection(D_Stop)

			ElevatorDoorLight(1)
			
			*e = RequestsClearAtCurrentFloor(*e)
			
			TimerStart(e.Config.DoorOpenDurationSec)
			
			SetAllLights(*e)
			
			e.Behaviour = EB_DoorOpen
		}
	default:
		// Do nothing for other behaviours
	}

	fmt.Println("\nNew state:")
	ElevatorPrint(*e)
}

func FSMOnDoorTimeout(e *Elevator) {
	fmt.Println("\n\nFSMOnDoorTimeout()")
	ElevatorPrint(*e)

	switch e.Behaviour {
	case EB_DoorOpen:
		pair := RequestsChooseDirection(*e)
		e.Dirn = pair.Dirn
		e.Behaviour = pair.Behaviour

		switch e.Behaviour {
		case EB_DoorOpen:
			TimerStart(e.Config.DoorOpenDurationSec)
			*e = RequestsClearAtCurrentFloor(*e)
			SetAllLights(*e)
		case EB_Moving, EB_Idle:
			ElevatorDoorLight(0)
			ElevatorMotorDirection(e.Dirn)
		}
	default:
		// Do nothing for other behaviours
	}

	fmt.Println("\nNew state:")
	ElevatorPrint(*e)
}