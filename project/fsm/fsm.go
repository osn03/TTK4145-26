package fsm

import (
	"fmt"
	"project/constant"
	"project/elevator"
	"project/elevio"
	"project/request"
	"project/timer"
	"project/types"
)

func OnInitBetweenFloors(e *types.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	e.Dirn = elevio.MD_Down
	e.Behaviour = types.EB_Moving
}

// EvaluateMovement decides what the Local elevator should do next given its current request state.
// Call this after assignments are updated (network merge) and from OnDoorTimeout when door closes.
// trur kanskje denne må brukes i network?
func EvaluateMovement(e *types.Elevator) {
	if e.Behaviour == types.EB_Moving {
		return
	}

	if elevio.GetObstruction() {
		e.Dirn = elevio.MD_Stop
		e.Behaviour = types.EB_DoorOpen
		elevio.SetMotorDirection(elevio.MD_Stop)
		elevio.SetDoorOpenLamp(true)
		timer.Start(constant.DoorOpenDurationMS)
		return
	}

	pair := request.ChooseDirection(*e)
	e.Dirn = pair.Dirn
	e.Behaviour = pair.Behaviour

	switch e.Behaviour {

	case types.EB_DoorOpen:
		elevio.SetMotorDirection(elevio.MD_Stop)
		elevio.SetDoorOpenLamp(true)
		timer.Start(constant.DoorOpenDurationMS)

		*e = request.ClearAtCurrentFloor(*e)

	case types.EB_Moving:
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(e.Dirn)

	case types.EB_Idle:
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elevio.MD_Stop)
	}
}

func ClearAllRequests(e *types.Elevator) {
	for f := 0; f < constant.NumFloors; f++ {
		for b := elevio.ButtonType(0); b < constant.NumButtons; b++ {
			e.Requests[f][b] = types.ReqNone
		}
	}
}

func OnRequestButtonPress(e *types.Elevator, floor int, btnType elevio.ButtonType) {

	switch e.Requests[floor][btnType] {
	case types.ReqNone:
		e.Requests[floor][btnType] = types.ReqUnconfirmed
		return
	case types.ReqUnconfirmed:
		e.Requests[floor][btnType] = types.ReqUnconfirmed
		return
	case types.ReqConfirmed:
		e.Requests[floor][btnType] = types.ReqConfirmed
		return
	case types.ReqDeleting:
		e.Requests[floor][btnType] = types.ReqUnconfirmed
		return
	}
}

func OnFloorArrival(e *types.Elevator, newFloor int) {
	fmt.Printf("\n\nFSMOnFloorArrival(%d)\n", newFloor)
	elevator.ElevatorPrint(*e)

	e.Floor = newFloor
	elevio.SetFloorIndicator(e.Floor)

	switch e.Behaviour {
	case types.EB_Moving:
		if request.ShouldStop(*e) {
			elevio.SetMotorDirection(elevio.MD_Stop)

			elevio.SetDoorOpenLamp(true)

			*e = request.ClearAtCurrentFloor(*e)

			timer.Start(constant.DoorOpenDurationMS)

			e.Behaviour = types.EB_DoorOpen
		}
	default:
		// Do nothing for other behaviours
	}

	fmt.Println("\nNew state:")
	elevator.ElevatorPrint(*e)
}

func OnDoorTimeout(e *types.Elevator) {
	if e.Behaviour != types.EB_DoorOpen {
		return
	}
	elevio.SetDoorOpenLamp(false)

	EvaluateMovement(e)
}

// legge til case som registrerer om mottat melding over channel fra esm og velger retning
func RunLocalElevator(transfer chan types.Elevator, ordersFromCost chan [constant.NumFloors][constant.NumButtons]types.ReqState) {

	var e types.Elevator

	OnInitBetweenFloors(&e)

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

			OnRequestButtonPress(&e, a.Floor, a.Button)

			transfer <- e

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)

			OnFloorArrival(&e, a)

			transfer <- e

		case a := <-time_timeout:
			fmt.Printf("%+v\n", a)
			timer.Stop()
			OnDoorTimeout(&e)

			transfer <- e

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a && e.Behaviour == types.EB_DoorOpen {
				timer.Stop()
				//state and motordirection remain unchanged

			} else {
				timer.Start(constant.DoorOpenDurationSec)
				//restarts timer that will trigger door close
			}

			transfer <- e

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
				e.Behaviour = types.EB_Idle
				e.Dirn = elevio.MD_Stop
				//sets states to match stopped elevator
			} else {
				pair := request.ChooseDirection(e)
				e.Dirn = pair.Dirn
				e.Behaviour = pair.Behaviour
				elevio.SetMotorDirection(e.Dirn)
			}
			transfer <- e

		case a := <-ordersFromCost:
			fmt.Printf("%+v\n", a)
			ClearAllRequests(&e)
			e.Requests = a
			EvaluateMovement(&e)
			transfer <- e

		}

	}
}
