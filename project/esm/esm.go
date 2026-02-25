package esm

import (
	"project/network"
	"project/constant"
	"project/elevator"
	"project/elevio"
	"time"
	"project/cost"
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

	AssignedLocal [constant.NumFloors][2]bool // Vet ikke om vi må ha denne
}

func UpdateOrders(worldview *WorldView) {
	for buttonType := elevio.ButtonType(0); buttonType < constant.NumButtons; buttonType++ {
		for floor := 0; floor < constant.NumFloors; floor++ {

			allUpdatet := 0

			for _, elev := range worldview.Elevators {

				if elev.Status == true {

					localrequest := worldview.local.Requests[floor][buttonType]
					externalrequest := elev.Elevator.Requests[floor][buttonType]

					if localrequest < externalrequest && !(localrequest == elevator.ReqUnconfirmed && externalrequest == elevator.ReqDeleting) {
						worldview.local.Requests[floor][buttonType] = externalrequest

					} else if localrequest == externalrequest && (localrequest == elevator.ReqUnconfirmed || localrequest == elevator.ReqDeleting) {
						allUpdatet += 1
					}
				}
			}
			if allUpdatet == worldview.OnlineElevators {
				if worldview.local.Requests[floor][buttonType] == elevator.ReqDeleting {
					worldview.local.Requests[floor][buttonType] = elevator.ReqNone
				} else if worldview.local.Requests[floor][buttonType] == elevator.ReqUnconfirmed {
					worldview.local.Requests[floor][buttonType] = elevator.ReqConfirmed
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
		Status: true,
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
	worldview.local.Floor = local.Floor
	worldview.local.Dirn = local.Dirn
	worldview.local.Behaviour = local.Behaviour

	for floor := 0; floor < numFloors; floor++ {
		for button := 0; button < numButtons; button++ {
			hardwareRequest := local.Requests[floor][button]
			storedRequest := worldview.local.Requests[floor][button]

			switch storedRequest {
			case elevator.ReqNone:
				if hardwareRequest != elevator.ReqUnconfirmed {
					worldview.local.Requests[floor][button] = elevator.ReqConfirmed
				}

			case elevator.ReqUnconfirmed:
				//do nothing

			case elevator.ReqConfirmed:
				if hardwareRequest == elevator.ReqDeleting {
					worldview.local.Requests[floor][button] = elevator.ReqDeleting
				}

			case elevator.ReqDeleting:
				//do nothing
			}
		}
	}
}

func ShareLocalStates(out chan ExternalElevator, localstatus bool, local elevator.Elevator) {
	outmessage := ExternalElevator{Status: localstatus, Elevator: local}
	out <- outmessage
}

func ComputeAssignments(worldview *WorldView, localID string) map[string][][]bool {
    // Build hallReqs
    hallReqs := make([][]bool, constant.NumFloors)
    for f := 0; f < constant.NumFloors; f++ {
        hallReqs[f] = make([]bool, 2)
        for btn := elevio.ButtonType(0); btn <= elevio.BT_HallDown; btn++ {
            hallReqs[f][btn] = elevator.ReqIsActive(worldview.local.Requests[f][btn])
        }
    }

    // Build elevatorStates (online + local)
    states := make(map[string]elevator.Elevator)
    for id, ext := range worldview.Elevators {
        if ext.Status {
            states[id] = ext.Elevator
        }
    }
    states[localID] = worldview.local

    return cost.OptimalHallRequests(hallReqs, states, true)
}

func ApplyLocalAssignment(worldview *WorldView, localID string, assigned map[string][][]bool) bool {
    a, ok := assigned[localID]
    if !ok {
        // if local missing, treat as all false
        changed := false
        for f := 0; f < constant.NumFloors; f++ {
            for b := 0; b < 2; b++ {
                if worldview.AssignedLocal[f][b] {
                    worldview.AssignedLocal[f][b] = false
                    changed = true
                }
            }
        }
        return changed
    }

    changed := false
    for f := 0; f < constant.NumFloors; f++ {
        for b := 0; b < 2; b++ {
            v := a[f][b]
            if worldview.AssignedLocal[f][b] != v {
                worldview.AssignedLocal[f][b] = v
                changed = true
            }
        }
    }
    return changed
}
func BuildLocalExecutorElevator(worldview *WorldView) elevator.Elevator {
    e := worldview.local // copy

    // Overwrite hall requests with "assigned to me" (as confirmed)
    for f := 0; f < constant.NumFloors; f++ {
        for btn := elevio.ButtonType(0); btn <= elevio.BT_HallDown; btn++ {
            if worldview.AssignedLocal[f][btn] {
                e.Requests[f][btn] = elevator.ReqConfirmed
            } else {
                // IMPORTANT: do not let unassigned hall affect local motion
                e.Requests[f][btn] = elevator.ReqNone
            }
        }
    }
    // Keep cab requests as-is
    return e
}


func SetAllLights(e elevator.Elevator) {
	for floor := 0; floor < constant.NumFloors; floor++ {
		for button := 0; button < constant.NumButtons; button++ {
			elevio.SetButtonLamp(elevio.ButtonType(button), floor, e.Requests[floor][button] == elevator.ReqConfirmed)
		}
	}
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
			/* assigned := ComputeAssigments(&worldview, localid)
			changed := ApplyLocalAssignment(&worldview, localid, assigned)
			if changed {
				// Build executor view and notify FSM to re-evaluate if it is idle/doorOpen.
				execE := BuildLocalExecutorElevator(&worldview)
				select {
				case fsmKick <- execE:
				default:
					// non-blocking: it's fine to drop if FSM will get another soon
				}
			} */

			SetAllLights(worldview.local)

		case local := <-hardware:
			ResetLocalTimeout(timeout)
			UpdateLocal(&worldview, local)
			ShareLocalStates(out, LocalStatus, local)

			SetAllLights(worldview.local)
		}
	}
}
