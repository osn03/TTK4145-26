package esm

import (
	network "project/Network"
	"project/Network/network"
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
	Timeout  *time.Timer
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

			for _, elev := range worldview.Elevators {

				if elev.status == true {

					localrequest := worldview.local.Requests[floor][buttonType]
					externalrequest := elev.elevator.Requests[floor][buttonType]

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
				} else {
					worldview.local.Requests[floor][buttonType] += 1
				}
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

	//Id must be string, Status must be bool, Elevator must be elevator.Elevator
}

func AddElevator(worldview *WorldView) {

	worldview.Elevators[id] = ExternalElevator{
		Status:   true,
		Elevator: elevator,
	}

	worldview.OnlineElevators += 1

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

func UpdateLocal(){

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

func RunESM(hardware chan elevator.Elevator, in chan network.Msg, out chan ExternalElevator, localid string) {
func SetAllLights(e elevator.Elevator) {
	for floor := 0; floor < constant.NumFloors; floor++ {
		for button := 0; button < constant.NumButtons; button++ {
			elevio.SetButtonLamp(elevio.ButtonType(button), floor, e.Requests[floor][button] == elevator.ReqConfirmed)
		}
	}
}


func RunESM(hardware chan elevator.Elevator) {
	//Denne funksjonen skal kjøres i egen gorouting, håndterer worldview, timouts og oppdatering av ordre

	timers := make(chan int)
	heartbeat := make(chan int)

	var worldview WorldView
	
	for {
		select {
		case <-timer:
			HandleTimeout(&LocalStatus)

		case message := <-msg:

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
