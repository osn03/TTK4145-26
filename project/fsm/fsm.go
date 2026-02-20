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
			elevio.SetButtonLamp(elevio.ButtonType(button), floor, e.Requests[floor][button] == elevator.ReqConfirmed)
		}
	}
}

func OnInitBetweenFloors(e *elevator.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	e.Dirn = elevio.MD_Down
	e.Behaviour = elevator.EB_Moving
}


// EvaluateMovement decides what the local elevator should do next given its current request state.
// Call this after assignments are updated (network merge) and from OnDoorTimeout when door closes.
func EvaluateMovement(e *elevator.Elevator) {
	if e.Behaviour == elevator.EB_Moving {
		return
	}

	if elevio.GetObstruction() {
		e.Dirn = elevio.MD_Stop
		e.Behaviour = elevator.EB_DoorOpen
		elevio.SetMotorDirection(elevio.MD_Stop)
		elevio.SetDoorOpenLamp(true)
		timer.Start(constant.DoorOpenDurationMS)
		return
	}

	pair := request.ChooseDirection(*e)
	e.Dirn = pair.Dirn
	e.Behaviour = pair.Behaviour

	switch e.Behaviour {

	case elevator.EB_DoorOpen:
		elevio.SetMotorDirection(elevio.MD_Stop)
		elevio.SetDoorOpenLamp(true)
		timer.Start(constant.DoorOpenDurationMS)

		
		*e = request.ClearAtCurrentFloor(*e)

		SetAllLights(*e)

	case elevator.EB_Moving:
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(e.Dirn)

	case elevator.EB_Idle:
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elevio.MD_Stop)
	}
}

func OnRequestButtonPress(e *elevator.Elevator, floor int, btnType elevio.ButtonType) {

	switch e.Requests[floor][btnType] {
	case elevator.ReqNone, elevator.ReqDeleting:
		e.Requests[floor][btnType] = elevator.ReqUnconfirmed
	default:
		// already active (unconfirmed/confirmed), do nothing
	}

	SetAllLights(*e)
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

			timer.Start(constant.DoorOpenDurationMS)

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
	if e.Behaviour != elevator.EB_DoorOpen {
		return
	}
	elevio.SetDoorOpenLamp(false)

	EvaluateMovement(e)
}

func RunLocalElevator(e *elevator.Elevator){

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	time_timeout := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go timer.TimedOut(time_timeout)

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			OnRequestButtonPress(e, a.Floor, a.Button)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)

			OnFloorArrival(e, a)

		case a := <-time_timeout:
			fmt.Printf("%+v\n", a)
			timer.Stop()
			OnDoorTimeout(e)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a && e.Behaviour == elevator.EB_DoorOpen {
				timer.Stop()
				//state and motordirection remain unchanged

			} else {
				timer.Start(constant.DoorOpenDurationSec)
				//restarts timer that will trigger door close
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
				e.Behaviour = elevator.EB_Idle
				e.Dirn = elevio.MD_Stop
				//sets states to match stopped elevator
			} else {
				pair := request.ChooseDirection(*e)
				e.Dirn = pair.Dirn
				e.Behaviour = pair.Behaviour
				elevio.SetMotorDirection(e.Dirn)
			}

		}
	}
}