package main

import (
	"fmt"
	"project/constant"
	"project/elevator"
	"project/elevio"
	"project/fsm"
	"project/timer"
)

func main() {

	numFloors := constant.NumFloors
	//numElevs := constant.NumElevators

	elevio.Init("localhost:15657", numFloors)

	var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	time_timeout := make(chan bool) //?

	var elevator elevator.Elevator
	fsm.OnInitBetweenFloors(&elevator)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go timer.TimedOut(time_timeout) //?

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			fsm.OnRequestButtonPress(&elevator, a.Floor, a.Button)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			/*if a == numFloors-1 {
				d = elevio.MD_Down
			} else if a == 0 {
				d = elevio.MD_Stop
			}*/
			fsm.OnFloorArrival(&elevator, a)

		case a := <-time_timeout:
			fmt.Printf("%+v\n", a)
			timer.Stop()
			fsm.OnDoorTimeout(&elevator)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)

				}
			}

		}
	}
}
