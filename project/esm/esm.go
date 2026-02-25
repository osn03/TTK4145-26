package esm

import (
	"fmt"
	"project/constant"
	"project/cost"
	"project/elevator"
	"project/elevio"
	"project/network"
	"project/network/peers"
	"project/types"
	"time"
)

const numFloors int = constant.NumFloors
const numButtons int = constant.NumButtons

func UpdateOrders(worldview *types.WorldView) {
	for buttonType := types.ButtonType(0); buttonType < constant.NumButtons; buttonType++ {
		for floor := 0; floor < constant.NumFloors; floor++ {

			allUpdatet := 1

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

func SetAllLights(e types.Elevator) {
	for floor := 0; floor < constant.NumFloors; floor++ {
		for button := 0; button < constant.NumButtons; button++ {
			elevio.SetButtonLamp(types.ButtonType(button), floor, e.Requests[floor][button] == types.ReqConfirmed)
		}
	}
}

func UpdatePeerStatus(worldview *types.WorldView, status peers.PeerUpdate) {
	for id, elev := range worldview.Elevators {
		for _, peer := range status.Lost {
			if peer == id {
				elev.Status = false
				worldview.OnlineElevators -= 1
			}
		}
	}
}

func RunESM(hardware chan types.Elevator, in chan network.Msg, out chan types.ExternalElevator, statusin chan peers.PeerUpdate, localid string, fsmKick chan [numFloors][numButtons]types.ReqState) {
	//Denne funksjonen skal kjøres i egen gorouting, håndterer worldview, timouts og oppdatering av ordre

	timer := make(chan bool)

	timeout := time.AfterFunc(constant.LocalTimoutDurationMS*time.Millisecond, func() {
		timer <- true
	})

	worldview := types.WorldView{
		Elevators:       make(map[string]types.ExternalElevator),
		OnlineElevators: 1,
		Local:           types.Elevator{},
	}

	LocalStatus := true

	for {
		select {
		case <-timer:
			fmt.Println("Local elevator timed out")
			HandleLocalTimeout(&LocalStatus)

		case message := <-in:
			fmt.Println("Received network update")

			UpdateWorldView(&worldview, message)
			UpdateOrders(&worldview)

			cost.AssignOrders(worldview, localid, fsmKick)

			SetAllLights(worldview.Local)

		case local := <-hardware:
			fmt.Println("Received local update")
			ResetLocalTimeout(timeout)
			UpdateLocal(&worldview, local)
			ShareLocalStates(out, LocalStatus, local)

			elevator.ElevatorPrint(worldview.Local)

			cost.AssignOrders(worldview, localid, fsmKick)

			SetAllLights(worldview.Local)

		case status := <-statusin:
			fmt.Println("Received peer status update")

			UpdatePeerStatus(&worldview, status)

			cost.AssignOrders(worldview, localid, fsmKick)
		}
	}
}
