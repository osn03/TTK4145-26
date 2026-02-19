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
		s.State.Behaviour = elevator.EB_Idle

	case elevator.EB_Idle:
		for c := 0; c < 2; c++ {
			if reqs[s.State.Floor][c].Active {
				reqs[s.State.Floor][c].AssignedTo = s.ID
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

	onClearRequest := func(c elevio.ButtonType) {
		switch c {
		case elevio.BT_HallUp, elevio.BT_HallDown:
			reqs[s.State.Floor][c].AssignedTo = s.ID
		case elevio.BT_Cab:
			s.State.Requests[s.State.Floor][elevio.BT_Cab] = 0
		}
	}

	switch s.State.Behaviour {

	case elevator.EB_Moving:
		if request.ShouldStop(e) {
			s.State.Behaviour = elevator.EB_DoorOpen
			s.Time += constant.DoorOpenDurationMS
			request.ClearAtCurrentFloor(e)
		} else {
			s.State.Floor += int(s.State.Dirn)
			s.Time += constant.TravelDurationMS
		}

	case elevator.EB_Idle, elevator.EB_DoorOpen:
		s.State.Dirn = request.ChooseDirection(e).Dirn

		if s.State.Dirn == elevio.MD_Stop {

			if request.Here(e) {
				request.ClearAtCurrentFloor(e)
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

func unvisitedAreImmediatelyAssignable(reqs [][]Req, states []State) bool {

	for _, s := range states {
		for _, cab := range s.State.Requests[s.State.Floor] {
			if cab {
				return false
			}
		}
	}

	for f, floor := range reqs {

		activeCount := 0
		for _, r := range floor {
			if r.Active {
				activeCount++
			}
		}
		if activeCount == 2 {
			return false
		}

		for _, r := range floor {
			if r.Active && r.AssignedTo == "" {

				found := false
				for _, s := range states {
					if s.State.Floor == f {
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

				for i := range states {
					if states[i].State.Floor == f {
						reqs[f][c].AssignedTo = states[i].ID
						states[i].Time += constant.DoorOpenDurationMS
						break
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



func withUnassignedRequests(
    s State,
    reqs [][]Req,
) elevator.Elevator {

    var e elevator.Elevator

    e.Floor = s.State.Floor
    e.Dirn = s.State.Dirn
    e.Behaviour = s.State.Behaviour

    //  Cab calls
    for f := 0; f < constant.NumFloors; f++ {
        if s.State.Requests[f][elevio.BT_Cab] == 1 || s.State.Requests[f][elevio.BT_Cab] == 2{
            e.Requests[f][elevio.BT_Cab] = 2
        } else {
			e.Requests[f][elevio.BT_Cab] = 0
		}
    }

    //  Hall calls
    for f := 0; f < constant.NumFloors; f++ {
        for btn := elevio.ButtonType(0); btn <= elevio.BT_HallDown; btn++ {

            r := reqs[f][int(btn)]

            if r.Active && (r.AssignedTo == "" || r.AssignedTo == s.ID) {
				e.Requests[f][btn] = 2
			}
        }
    }

    return e
}


