package esm

import (
	"project/constant"
	"project/elevator"
	"project/elevio"
	"time"
)

const numFloors int = constant.NumFloors
const numButtons int = constant.NumButtons

type ExternalElevator struct {
	status   bool
	timeout  *time.Timer
	elevator elevator.Elevator
}

type WorldView struct {
	Elevators       map[int]ExternalElevator
	OnlineElevators int
	local		 elevator.Elevator
}

func UpdateOrders(worldview *WorldView) {
	for buttonType := elevio.ButtonType(0); buttonType < constant.NumButtons; buttonType++ {
		for floor := 0; floor < constant.NumFloors; floor++ {

			allUpdatet := 0

			for id, elev := range worldview.Elevators {

				if elev.status == true {

					localrequest := worldview.local.Requests[floor][buttonType]
					externalrequest := elev.elevator.Requests[floor][buttonType]

					if localrequest < externalrequest {
						worldview.local.Requests[floor][buttonType] = externalrequest

					} else if localrequest == externalrequest && localrequest > 0 {
						allUpdatet += 1
					}
				}
			}
			if allUpdatet == worldview.OnlineElevators {
				worldview.local.Requests[floor][buttonType] += 1
			}
		}
	}
}

func UpdateWorldView(worldview *WorldView, extelevator ExternalElevator, id int, out chan<- int) {

	if existing, ok := worldview.Elevators[id]; ok {

		existing.elevator = extelevator.elevator
		existing.status = extelevator.status

		worldview.Elevators[id] = existing
		return
	}

	AddElevator(worldview, id, extelevator.elevator, out)

	//Id must be int, Status must be bool, Elevator must be elevator.Elevator
}

func AddElevator(worldview *WorldView, id int, elevator elevator.Elevator, out chan<- int) {

	timeout := time.AfterFunc(constant.NetworkTimeout*time.Millisecond, func() {
		out <- id
	})

	worldview.Elevators[id] = ExternalElevator{
		status:   true,
		elevator: elevator,
		timeout:  timeout,
	}

	worldview.OnlineElevators += 1

	//denne funksjonen fungerer ikke, men mer eller mindre en plassholder for å legge til en ny elevator.
}

func HandleTimeout(id int, worldview *WorldView) {
	worldview.Elevators[id] = ExternalElevator{status: false}
	worldview.OnlineElevators -= 1
}

func ResetTimeout(id int, worldview *WorldView) {
	if existing, ok := worldview.Elevators[id]; ok {
		existing.timeout.Reset(constant.NetworkTimeout * time.Millisecond)
	}
}

func UpdateLocal(){

}


func RunESM(hardware chan elevator.Elevator) {
	//Denne funksjonen skal kjøres i egen gorouting, håndterer worldview, timouts og oppdatering av ordre

	timers := make(chan int)
	heartbeat := make(chan int)

	var worldview WorldView
	
	for {
		select {
		case id := <-timers:
			HandleTimeout(id, &worldview)

		case message := <-msg:

			UpdateWorldView(&worldview, message.elevator, message.id, timers)
			UpdateOrders(&worldview) 

		case id := <-heartbeat:
			ResetTimeout(id, &worldview)

		case local := <-hardware:
			worldview.local = local
		}
	}
}
