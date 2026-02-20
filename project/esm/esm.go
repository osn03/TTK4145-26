package esm

import (
	"project/Network/network"
	"project/constant"
	"project/elevator"
	"project/elevio"
	"time"
)

const numFloors int = constant.NumFloors
const numButtons int = constant.NumButtons

type ExternalElevator struct {
	Status   bool
	Elevator elevator.Elevator
}

type WorldView struct {
	Elevators       map[int]ExternalElevator
	OnlineElevators int
	local           elevator.Elevator
}

func UpdateOrders(worldview *WorldView) {
	for buttonType := elevio.ButtonType(0); buttonType < constant.NumButtons; buttonType++ {
		for floor := 0; floor < constant.NumFloors; floor++ {

			allUpdatet := 0

			for id, elev := range worldview.Elevators {

				if elev.Status == true {

					localrequest := worldview.local.Requests[floor][buttonType]
					externalrequest := elev.Elevator.Requests[floor][buttonType]

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

		existing.Elevator = extelevator.Elevator
		existing.Status = extelevator.Status

		worldview.Elevators[id] = existing
		return
	}

	AddElevator(worldview, id, extelevator.Elevator, out)

	//Id must be int, Status must be bool, Elevator must be elevator.Elevator
}

func AddElevator(worldview *WorldView, id int, elevator elevator.Elevator, out chan<- int) {

	worldview.Elevators[id] = ExternalElevator{
		Status:   true,
		Elevator: elevator,
	}

	worldview.OnlineElevators += 1

	//denne funksjonen fungerer ikke, men mer eller mindre en plassholder for å legge til en ny elevator.
}

func HandleTimeout(status *bool) {
	*status = false
}

func ResetLocalTimeout(timer *time.Timer) {
	timer.Reset(constant.LocalTimoutDurationMS * time.Millisecond)
}

func UpdateLocal(worldview *WorldView, local elevator.Elevator) {
	worldview.local = local
}

func ShareLocalStates(out chan ExternalElevator, localstatus bool, local elevator.Elevator){
	outmessage := ExternalElevator{Status: localstatus, Elevator: local}
	out <- outmessage
}

func RunESM(hardware chan elevator.Elevator, in chan network.Msg, out chan ExternalElevator) {
	//Denne funksjonen skal kjøres i egen gorouting, håndterer worldview, timouts og oppdatering av ordre

	timer := make(chan bool)

	timeout := time.AfterFunc(constant.LocalTimoutDurationMS*time.Millisecond, func() {
		timer <- true
	})

	var worldview WorldView
	LocalStatus := true

	for {
		select {
		case <-timer:
			HandleTimeout(&LocalStatus)

		case message := <-in:

			UpdateWorldView(&worldview, message.ExternalElevator, message.id)
			UpdateOrders(&worldview)

		case local := <-hardware:
			ResetLocalTimeout(timeout)
			UpdateLocal(&worldview, local)
			ShareLocalStates(out, LocalStatus, local)
		}
	}
}
