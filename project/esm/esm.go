package esm

import (
	"project/constant"
	"project/elevator"
	"project/elevio"
)

const numFloors int = constant.NumFloors
const numButtons int = constant.NumButtons

type ExternalElevator struct {
	status   bool
	elevator elevator.Elevator
}

type WorldView struct {
	Elevators       map[int]ExternalElevator
	OnlineElevators int
}

func UpdateOrders(internal *elevator.Elevator, worldview WorldView) {
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

func UpdateWorldView(worldview *WorldView, message network.Msg) {
	worldview.Elevators[message.Id] = ExternalElevator{status: message.Status, elevator: message.Elevator}
	//Id must be int, Status must be bool, Elevator must be elevator.Elevator
}

func NetworkTimeout() {

}
