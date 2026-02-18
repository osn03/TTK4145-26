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
	status   	bool
	timeout 	*time.Timer
	elevator 	elevator.Elevator
}

type WorldView struct {
	Elevators       	map[int]ExternalElevator
	OnlineElevators 	int
}

func UpdateOrders(internal *elevator.Elevator, worldview *WorldView) {
	for buttonType := elevio.ButtonType(0); buttonType < constant.NumButtons; buttonType++ {
		for floor := 0; floor < constant.NumFloors; floor++ {

			allUpdatet := 0

			for id, elev := range worldview.Elevators {

				if elev.status == true {

					localrequest := internal.Requests[floor][buttonType]
					externalrequest := elev.elevator.Requests[floor][buttonType]

					if localrequest < externalrequest {
						internal.Requests[floor][buttonType] = externalrequest

					} else if localrequest == externalrequest && localrequest > 0 {
						allUpdatet += 1
					}
				}
			}
			if allUpdatet == worldview.OnlineElevators {
				internal.Requests[floor][buttonType] += 1
			}
		}
	}
}

func UpdateWorldView(worldview *WorldView, extelevator ExternalElevator, id int) {

	if existing, ok := worldview.Elevators[id]; ok {
		existing.timeout.Reset(constant.NetworkTimeout * time.Millisecond)

		existing.elevator 	= extelevator.elevator
		existing.status 	= extelevator.status

		worldview.Elevators[id] = existing
		return
	}

	AddElevator(worldview, id, extelevator.elevator)

	
	
	//Id must be int, Status must be bool, Elevator must be elevator.Elevator
}

func AddElevator(worldview *WorldView, id int, elevator elevator.Elevator) {
	timeout:= time.AfterFunc(constant.NetworkTimeout * time.Millisecond, func () {
		detectTimout(make(chan int), id)
	})

	worldview.Elevators[id] = ExternalElevator{
			status: 	true, 
			elevator: 	elevator,
			timeout: 	timeout,
		}
	
	worldview.OnlineElevators += 1

	//denne funksjonen fungerer ikke, men mer eller mindre en plassholder for å legge til en ny elevator.
}

func detectTimout(out chan<- int) {
	out <- id

}

func HandleTimeout(id int, worldview *WorldView) {
	worldview.Elevators[id] = ExternalElevator{status: false}
	worldview.OnlineElevators -= 1
}


func RunESM(){
	//Denne funksjonen skal kjøres i egen gorouting, håndterer worldview, timouts og oppdatering av ordre

	timers := make(chan int)

	var worldview WorldView

	go UpdateOrders(elevator, &worldview) //elevator må fikses her
	

	select {
		case id := <- timers:
			HandleTimeout(id, &worldview)
		
		case extelevator := <- msg:

	}

}