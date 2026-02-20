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

func OnRequestButtonPress(e *elevator.Elevator, floor int, btnType elevio.ButtonType) {
	fmt.Printf("\n\nFSMOnRequestButtonPress(%d, %s)\n", floor, elevator.ButtonToString(btnType))
	elevator.ElevatorPrint(*e)

	switch e.Behaviour {
	case elevator.EB_DoorOpen:
		if request.ShouldClearImmediately(*e, floor, btnType) {
			timer.Start(constant.DoorOpenDurationMS)
		} else {
			e.Requests[floor][btnType] = 1
		}

	case elevator.EB_Moving:
		e.Requests[floor][btnType] = 1

	case elevator.EB_Idle:
		e.Requests[floor][btnType] = 1
		pair := request.ChooseDirection(*e)
		e.Dirn = pair.Dirn
		e.Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case elevator.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.Start(constant.DoorOpenDurationMS)
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
	fmt.Println("\n\nFSMOnDoorTimeout()")
	elevator.ElevatorPrint(*e)

	switch e.Behaviour {

	case elevator.EB_DoorOpen:

		// NEW: obstruction policy belongs in FSM (not request logic)
		if elevio.GetObstruction() {
			// Keep door open, motor stopped, restart timer
			e.Dirn = elevio.MD_Stop
			e.Behaviour = elevator.EB_DoorOpen

			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			timer.Start(constant.DoorOpenDurationMS)

			// Optional: lights unchanged, but you can refresh if you want:
			SetAllLights(*e)

			fmt.Println("\nNew state (obstructed):")
			elevator.ElevatorPrint(*e)
			return
		}

		// Normal flow
		pair := request.ChooseDirection(*e)
		e.Dirn = pair.Dirn
		e.Behaviour = pair.Behaviour

		switch e.Behaviour {

		case elevator.EB_DoorOpen:
			timer.Start(constant.DoorOpenDurationMS)
			*e = request.ClearAtCurrentFloor(*e)
			SetAllLights(*e)

		case elevator.EB_Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(e.Dirn)

		case elevator.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MD_Stop)
			// Optional: SetAllLights(*e)

		}

	default:
		// Do nothing for other behaviours
	}

	fmt.Println("\nNew state:")
	elevator.ElevatorPrint(*e)
}


func RunLocalElevator(transfer chan elevator.Elevator){

	var e elevator.Elevator

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
			elevio.SetButtonLamp(a.Button, a.Floor, true)
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
			if a && e.Behaviour == elevator.EB_DoorOpen {
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
				e.Behaviour = elevator.EB_Idle
				e.Dirn = elevio.MD_Stop
				//sets states to match stopped elevator
			} else {
				pair := request.ChooseDirection(e)
				e.Dirn = pair.Dirn
				e.Behaviour = pair.Behaviour
				elevio.SetMotorDirection(e.Dirn)
			}
			transfer <- e

		}
	}
}