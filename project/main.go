package main

import (
	"fmt"
	"project/constant"
	"project/elevator"
	"project/elevio"
	"project/fsm"
	"project/request"
	"project/timer"
)

func main() {

	numFloors := constant.NumFloors
	//numElevs := constant.NumElevators

	elevio.Init("localhost:15657", numFloors)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	time_timeout := make(chan bool)

	var e elevator.Elevator
	fsm.OnInitBetweenFloors(&e)

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
			fsm.OnRequestButtonPress(&e, a.Floor, a.Button)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)

			fsm.OnFloorArrival(&e, a)

		case a := <-time_timeout:
			fmt.Printf("%+v\n", a)
			timer.Stop()
			fsm.OnDoorTimeout(&e)

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
			} else {
				pair := request.ChooseDirection(e)
				e.Dirn = pair.Dirn
				e.Behaviour = pair.Behaviour
				elevio.SetMotorDirection(e.Dirn)
			}

		}
	}
}
