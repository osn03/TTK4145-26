package esm

import (
	network "project/Network"
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
	Elevators       map[string]ExternalElevator
	OnlineElevators int
	local           elevator.Elevator
}

func UpdateOrders(worldview *WorldView) {
	for buttonType := elevio.ButtonType(0); buttonType < constant.NumButtons; buttonType++ {
		for floor := 0; floor < constant.NumFloors; floor++ {

			allUpdatet := 0

			for _, elev := range worldview.Elevators {

				if elev.Status == true {

					localrequest := worldview.local.Requests[floor][buttonType]
					externalrequest := elev.Elevator.Requests[floor][buttonType]

					if localrequest < externalrequest {
						worldview.local.Requests[floor][buttonType] = externalrequest

					} else if localrequest == externalrequest && (localrequest == elevator.ReqUnconfirmed || localrequest == elevator.ReqDeleting) {
						allUpdatet += 1
					}
				}
			}
			if allUpdatet == worldview.OnlineElevators {
				if worldview.local.Requests[floor][buttonType] == elevator.ReqDeleting {
					worldview.local.Requests[floor][buttonType] = elevator.ReqNone
				}else {
					worldview.local.Requests[floor][buttonType] += 1
				}
			}
		}
	}
}

func UpdateWorldView(worldview *WorldView, message network.Msg) {

	if existing, ok := worldview.Elevators[message.Id]; ok {

		existing.Elevator = elevator.Elevator{
			Floor:     message.Floor,
			Dirn:      message.Dirn,
			Requests:  message.Requests,
			Behaviour: message.Behaviour,
		}

		existing.Status = message.Status

		worldview.Elevators[message.Id] = existing
		return
	}

	AddElevator(worldview, message)

	//Id must be string, Status must be bool, Elevator must be elevator.Elevator
}

func AddElevator(worldview *WorldView, message network.Msg) {

	worldview.Elevators[message.Id] = ExternalElevator{
		Status:   true,
		Elevator: elevator.Elevator{
			Floor:     message.Floor,
			Dirn:      message.Dirn,
			Requests:  message.Requests,
			Behaviour: message.Behaviour,
		},
	}

	worldview.OnlineElevators += 1

}

func HandleLocalTimeout(status *bool) {
	*status = false
}

func ResetLocalTimeout(timer *time.Timer) {
	timer.Reset(constant.LocalTimoutDurationMS * time.Millisecond)
}

func UpdateLocal(worldview *WorldView, local elevator.Elevator) {
	worldview.local = local
}

func ShareLocalStates(out chan ExternalElevator, localstatus bool, local elevator.Elevator) {
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
			HandleLocalTimeout(&LocalStatus)

		case message := <-in:

			UpdateWorldView(&worldview, message)
			UpdateOrders(&worldview)
			//kjøre kostfunk for å finne ut hva min heis skal gjøre

		case local := <-hardware:
			ResetLocalTimeout(timeout)
			UpdateLocal(&worldview, local)
			ShareLocalStates(out, LocalStatus, local)
		}
	}
}
