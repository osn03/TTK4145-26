package cost_fns

import (
	"project/constant"
	"project/elevator"
	"project/elevio"
	"project/request"
	"sort"
)

//
// ==========================
// INTERNAL TYPES
// ==========================
//

type Req struct {
	Active     bool
	AssignedTo string
}

type State struct {
	ID    string
	State elevator.Elevator
	Time  int64
}

//
// ==========================
// ENTRY POINT
// ==========================
//

func OptimalHallRequests(
	hallReqs [][]bool,
	elevatorStates map[string]elevator.Elevator,
	includeCab bool,
) map[string][][]bool {

	numFloors := len(hallReqs)

	if len(elevatorStates) == 0 {
		panic("No elevator states provided")
	}

	for _, s := range elevatorStates {
		if len(s.Requests) != numFloors {
			panic("Hall and cab requests do not have same length")
		}
		if s.Floor < 0 || s.Floor >= numFloors {
			panic("Elevator at invalid floor")
		}
	}

	reqs := toReq(hallReqs)
	states := initialStates(elevatorStates)

	for i := range states {
		performInitialMove(&states[i], reqs)
	}

	for {

		sort.Slice(states, func(i, j int) bool {
			if states[i].Time != states[j].Time {
				return states[i].Time < states[j].Time
			}
			return states[i].ID < states[j].ID
		})


		done := true

		if anyUnassigned(reqs) {
			done = false
		}

		if unvisitedAreImmediatelyAssignable(reqs, states) {
			assignImmediate(reqs, states)
			done = true
		}

		if done {
			break
		}

		performSingleMove(&states[0], reqs)
	}

	result := make(map[string][][]bool)

	for id := range elevatorStates {

		buttons := 2
		if includeCab {
			buttons = 3
		}

		result[id] = make([][]bool, numFloors)

		for f := 0; f < numFloors; f++ {
			result[id][f] = make([]bool, buttons)

			if includeCab {
				result[id][f][2] = elevator.ReqIsActive(elevatorStates[id].Requests[f][elevio.BT_Cab])
			}
		}
	}

	for f := 0; f < numFloors; f++ {
		for c := 0; c < 2; c++ {
			if reqs[f][c].Active {
				id := reqs[f][c].AssignedTo	
				if id != "" {
					result[id][f][c] = true
				}
			}
		}
	}

	return result
}

//
// ==========================
// HELPERS
// ==========================
//

func toReq(hallReqs [][]bool) [][]Req {
	numFloors := len(hallReqs)
	r := make([][]Req, numFloors)

	for f := 0; f < numFloors; f++ {
		r[f] = make([]Req, 2)
		for b := 0; b < 2; b++ {
			r[f][b] = Req{
				Active:     hallReqs[f][b],
				AssignedTo: "",
			}
		}
	}
	return r
}

func initialStates(states map[string]elevator.Elevator) []State {
	ids := make([]string, 0, len(states))
	for id := range states {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	result := make([]State, len(ids))

	for i, id := range ids {
		result[i] = State{
			ID:    id,
			State: copyState(states[id]),
			Time:  int64(i), // tilsvarer usecs i D
		}
	}
	return result
}

func copyState(s elevator.Elevator) elevator.Elevator {
	cab := make([]int, len(s.Requests))
	for f := 0; f < len(s.Requests); f++ {
		cab[f] = s.Requests[f][elevio.BT_Cab]
	}

	return elevator.Elevator{
		Behaviour: s.Behaviour,
		Floor:     s.Floor,
		Dirn:      s.Dirn,
		Requests:  s.Requests,
	}
}

//
// ==========================
// REQUEST LOGIC
// ==========================
//

func anyUnassigned(reqs [][]Req) bool {
	for _, floor := range reqs {
		for _, r := range floor {
			if r.Active && r.AssignedTo == "" {
				return true
			}
		}
	}
	return false
}

//
// ==========================
// INITIAL MOVE
// ==========================
//

func performInitialMove(s *State, reqs [][]Req) {

	switch s.State.Behaviour {

	case elevator.EB_DoorOpen:
		s.Time += constant.DoorOpenDurationMS / 2
		fallthrough

	case elevator.EB_Idle:
		for btn := 0; btn < 2; btn++ {
			if reqs[s.State.Floor][btn].Active  && reqs[s.State.Floor][btn].AssignedTo == "" {
				reqs[s.State.Floor][btn].AssignedTo = s.ID
				s.Time += constant.DoorOpenDurationMS
			}
		}

	case elevator.EB_Moving:
		s.State.Floor += int(s.State.Dirn)
		s.Time += constant.TravelDurationMS / 2
	}
}

//
// ==========================
// SINGLE MOVE
// ==========================
//

func performSingleMove(s *State, reqs [][]Req) {

	e := withUnassignedRequests(*s, reqs)

	onClearRequest := func(btn elevio.ButtonType) {
		switch btn {
		case elevio.BT_HallUp, elevio.BT_HallDown:
			reqs[s.State.Floor][btn].AssignedTo = s.ID
		case elevio.BT_Cab:
			s.State.Requests[s.State.Floor][elevio.BT_Cab] = 0
		}
	}
	e = request.ClearAtCurrentFloorWithCallback(e, onClearRequest)

	switch s.State.Behaviour {

	case elevator.EB_Moving:
		if request.ShouldStop(e) {
			s.State.Behaviour = elevator.EB_DoorOpen
			s.Time += constant.DoorOpenDurationMS
			request.ClearAtCurrentFloorWithCallback(e, onClearRequest)
		} else {
			s.State.Floor += int(s.State.Dirn)
			s.Time += constant.TravelDurationMS
		}

	case elevator.EB_Idle, elevator.EB_DoorOpen:
		s.State.Dirn = request.ChooseDirection(e).Dirn

		if s.State.Dirn == elevio.MD_Stop {

			if request.Here(e) {
				request.ClearAtCurrentFloorWithCallback(e, onClearRequest)
				s.Time += constant.DoorOpenDurationMS
				s.State.Behaviour = elevator.EB_DoorOpen
			} else {
				s.State.Behaviour = elevator.EB_Idle
			}

		} else {
			s.State.Behaviour = elevator.EB_Moving
			s.State.Floor += int(s.State.Dirn)
			s.Time += constant.TravelDurationMS
		}
	}
}

//
// ==========================
// IMMEDIATE ASSIGN
// ==========================
//

func elevatorHasAnyCab(e elevator.Elevator) bool {
	for f := 0; f < constant.NumFloors; f++ {
		if e.Requests[f][elevio.BT_Cab] {
			return true
		}
	}
	return false
}

func unvisitedAreImmediatelyAssignable(reqs [][]Req, states []State) bool {
	// 1) Hvis noen heis har noen cab-request -> false
	for _, s := range states {
		if elevatorHasAnyCab(s.State) {
			return false
		}
	}

	// 2) Ingen etasje kan ha både hallUp og hallDown aktive samtidig
	for _, floorReqs := range reqs {
		activeCount := 0
		for _, r := range floorReqs {
			if r.Active {
				activeCount++
			}
		}
		if activeCount == 2 {
			return false
		}
	}

	// 3) Alle unassigned hall-requests må være på en etasje hvor det står en heis (som også har 0 cab)
	for f, floorReqs := range reqs {
		for _, r := range floorReqs {
			if r.Active && r.AssignedTo == "" {
				found := false
				for _, s := range states {
					if s.State.Floor == f && !elevatorHasAnyCab(s.State) {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		}
	}

	return true
}



func assignImmediate(reqs [][]Req, states []State) {
    for f := range reqs {
        for c := range reqs[f] {
            if reqs[f][c].Active && reqs[f][c].AssignedTo == "" {
                // assign to first elevator on floor with NO cab requests
                for i := range states {
                    if states[i].State.Floor == f {
                        // check for no cab requests
                        hasCab := false
                        for _, cab := range states[i].State.Requests[f] {
                            if cab { hasCab = true; break }
                        }
                        if !hasCab {
                            reqs[f][c].AssignedTo = states[i].ID
                            states[i].Time += constant.DoorOpenDurationMS
                            break
                        }
                    }
                }
            }
        }
    }
}


//
// ==========================
// ELEVATOR WRAPPER (STUB)
// ==========================
//




func withUnassignedRequests(s State, reqs [][]Req) elevator.Elevator {
	var e elevator.Elevator
	e.Floor = s.State.Floor
	e.Dirn = s.State.Dirn
	e.Behaviour = s.State.Behaviour

	// Cab
	for f := 0; f < constant.NumFloors; f++ {
		e.Requests[f][elevio.BT_Cab] = s.State.Requests[f][elevio.BT_Cab]
	}

	// Hall (kun 0..1)
	for f := 0; f < constant.NumFloors; f++ {
		for btn := elevio.ButtonType(0); btn <= elevio.BT_HallDown; btn++ {
			r := reqs[f][int(btn)]
			if r.Active && (r.AssignedTo == "" || r.AssignedTo == s.ID) {
				e.Requests[f][btn] = true
			}
		}
	}
	return e
}
