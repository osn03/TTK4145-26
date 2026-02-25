package esm

import (
	"project/constant"
	"project/cost"
	"project/elevator"
	"project/elevio"
	"project/network"
	"project/types"
	"time"
)

const numFloors int = constant.NumFloors
const numButtons int = constant.NumButtons

func UpdateOrders(worldview *types.WorldView) {
	for buttonType := elevio.ButtonType(0); buttonType < constant.NumButtons; buttonType++ {
		for floor := 0; floor < constant.NumFloors; floor++ {

			allUpdatet := 0

			for _, elev := range worldview.Elevators {

				if elev.Status == true {

					Localrequest := worldview.Local.Requests[floor][buttonType]
					externalrequest := elev.Elevator.Requests[floor][buttonType]

					if Localrequest < externalrequest && !(Localrequest == types.ReqUnconfirmed && externalrequest == types.ReqDeleting) {
						worldview.Local.Requests[floor][buttonType] = externalrequest

					} else if Localrequest == externalrequest && (Localrequest == types.ReqUnconfirmed || Localrequest == types.ReqDeleting) {
						allUpdatet += 1
					}
				}
			}
			if allUpdatet == worldview.OnlineElevators {
				if worldview.Local.Requests[floor][buttonType] == types.ReqDeleting {
					worldview.Local.Requests[floor][buttonType] = types.ReqNone
				} else if worldview.Local.Requests[floor][buttonType] == types.ReqUnconfirmed {
					worldview.Local.Requests[floor][buttonType] = types.ReqConfirmed
				}
			}
		}
	}
}

func UpdateWorldView(worldview *types.WorldView, message network.Msg) {

	if existing, ok := worldview.Elevators[message.Id]; ok {

		existing.Elevator = types.Elevator{
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

func AddElevator(worldview *types.WorldView, message network.Msg) {

	worldview.Elevators[message.Id] = types.ExternalElevator{
		Status: true,
		Elevator: types.Elevator{
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

func UpdateLocal(worldview *types.WorldView, Local types.Elevator) {
	worldview.Local.Floor = Local.Floor
	worldview.Local.Dirn = Local.Dirn
	worldview.Local.Behaviour = Local.Behaviour

	for floor := 0; floor < numFloors; floor++ {
		for button := 0; button < numButtons; button++ {
			hardwareRequest := Local.Requests[floor][button]
			storedRequest := worldview.Local.Requests[floor][button]

			switch storedRequest {
			case types.ReqNone:
				if hardwareRequest != types.ReqUnconfirmed {
					worldview.Local.Requests[floor][button] = types.ReqConfirmed
				}

			case types.ReqUnconfirmed:
				//do nothing

			case types.ReqConfirmed:
				if hardwareRequest == types.ReqDeleting {
					worldview.Local.Requests[floor][button] = types.ReqDeleting
				}

			case types.ReqDeleting:
				//do nothing
			}
		}
	}
}

func ShareLocalStates(out chan types.ExternalElevator, Localstatus bool, Local types.Elevator) {
	outmessage := types.ExternalElevator{Status: Localstatus, Elevator: Local}
	out <- outmessage
}

func SetAllLights(e elevator.Elevator) {
	for floor := 0; floor < constant.NumFloors; floor++ {
		for button := 0; button < constant.NumButtons; button++ {
			elevio.SetButtonLamp(elevio.ButtonType(button), floor, e.Requests[floor][button] == types.ReqConfirmed)
		}
	}
}

func RunESM(hardware chan elevator.Elevator, in chan network.Msg, out chan ExternalElevator, localid string, fsmKick chan [numFloors][numButtons]ReqState) {
	//Denne funksjonen skal kjøres i egen gorouting, håndterer worldview, timouts og oppdatering av ordre

	timer := make(chan bool)

	timeout := time.AfterFunc(constant.LocalTimoutDurationMS*time.Millisecond, func() {
		timer <- true
	})

	var worldview types.WorldView
	LocalStatus := true

	for {
		select {
		case <-timer:
			HandleLocalTimeout(&LocalStatus)

		case message := <-in:

			UpdateWorldView(&worldview, message)
			UpdateOrders(&worldview)
			cost.AssignOrders(&worldview, localid, fsmKick) 

			SetAllLights(worldview.Local)

		case Local := <-hardware:
			ResetLocalTimeout(timeout)
			UpdateLocal(&worldview, local)
			ShareLocalStates(out, LocalStatus, local)
			cost.AssignOrders(&worldview, localid, fsmKick) 

			SetAllLights(worldview.Local)
		}
	}
}
