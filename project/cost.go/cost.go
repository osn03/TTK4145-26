package cost

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

// OptimalHallRequests simulates all elevators forward in time and assigns each active hall request
// to the elevator that will serve it first (tie-broken by simulated time then lexicographic ID).
// If includeCab is true, the returned matrix includes a third column with the elevator's current cab requests.
func OptimalHallRequests(
	hallReqs [][]bool,
	elevatorStates map[string]elevator.Elevator,
	includeCab bool,
) map[string][][]bool {

	numFloors := len(hallReqs)

	if len(elevatorStates) == 0 {
		panic("No elevator states provided")
	}

	// Basic input validation (floor bounds and "moving out of bounds" sanity).
	for id, s := range elevatorStates {
		if len(s.Requests) != numFloors {
			panic("Hall and cab requests do not have same length")
		}
		if s.Floor < 0 || s.Floor >= numFloors {
			panic("Elevator at invalid floor")
		}
		if s.Behaviour == elevator.EB_Moving {
			next := s.Floor + int(s.Dirn)
			if next < 0 || next >= numFloors {
				panic("Elevator " + id + " is moving out of bounds")
			}
		}
	}

	// Convert hallReqs into internal Req structs and build a stable, ID-sorted list of simulation states.
	reqs := toReq(hallReqs)
	states := initialStates(elevatorStates)

	// Apply an initial half-step so elevators start the event simulation aligned with D's timing model.
	for i := range states {
		performInitialMove(&states[i], reqs)
	}

	// Event-driven simulation loop: always advance the elevator with the earliest simulated time.
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

		// Fast-path: if all remaining unassigned hall calls are immediately assignable (no cab calls, etc.),
		// assign them without further simulation.
		if unvisitedAreImmediatelyAssignable(reqs, states) {
			assignImmediate(reqs, states)
			done = true
		}

		if done {
			break
		}

		// Simulate exactly one "event step" for the earliest elevator in time.
		performSingleMove(&states[0], reqs)
	}

	// Build output map: per elevator ID, a per-floor [2] (hall) or [3] (hall+cab) boolean matrix.
	result := make(map[string][][]bool)

	for id := range elevatorStates {
		buttons := 2
		if includeCab {
			buttons = 3
		}

		result[id] = make([][]bool, numFloors)
		for f := 0; f < numFloors; f++ {
			result[id][f] = make([]bool, buttons)

			// Optional third column mirrors whether the cab request at that floor is currently active.
			if includeCab {
				result[id][f][2] = elevator.ReqIsActive(elevatorStates[id].Requests[f][elevio.BT_Cab])
			}
		}
	}

	// Write final hall assignments into the output matrix.
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

// toReq converts a hall request boolean matrix [floor][2] into internal Req structs with empty assignments.
func toReq(hallReqs [][]bool) [][]Req {
	numFloors := len(hallReqs)
	r := make([][]Req, numFloors)

	for f := 0; f < numFloors; f++ {
		r[f] = make([]Req, 2)
		for b := 0; b < 2; b++ {
			r[f][b] = Req{/ corresponds to the microsecond offset used for tie-breaking in the D version
				Active:     hallReqs[f][b],
				AssignedTo: "",
			}
		}
	}
	return r
}

// initialStates creates a deterministic slice of simulation states, sorted by elevator ID,
// and seeds each elevator with a small increasing initial Time to enforce lexicographic tie-breaking.
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
			Time:  int64(i), 
		}
	}
	return result
}


func copyState(s elevator.Elevator) elevator.Elevator {
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

// anyUnassigned returns true if there exists any active hall request that has not yet been assigned to an elevator.
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

// performInitialMove applies the initial half-step timing normalization:
// - doorOpen: consume half door time, then treat as idle
// - idle: immediately take any hall requests at current floor
// - moving: advance one floor (half travel time), representing "arriving" at the next floor in the simulation model.
func performInitialMove(s *State, reqs [][]Req) {
	numFloors := len(reqs)

	switch s.State.Behaviour {

	case elevator.EB_DoorOpen:
		s.Time += constant.DoorOpenDurationMS / 2
		fallthrough

	case elevator.EB_Idle:
		for btn := 0; btn < 2; btn++ {
			if reqs[s.State.Floor][btn].Active && reqs[s.State.Floor][btn].AssignedTo == "" {
				reqs[s.State.Floor][btn].AssignedTo = s.ID
				s.Time += constant.DoorOpenDurationMS
			}
		}

	case elevator.EB_Moving:
		next := s.State.Floor + int(s.State.Dirn)
		if next < 0 || next >= numFloors {
			s.State.Dirn = elevio.MD_Stop
			s.State.Behaviour = elevator.EB_Idle
			return
		}
		s.State.Floor = next
		s.Time += constant.TravelDurationMS / 2
	}
}

//
// ==========================
// SINGLE MOVE
// ==========================
//

// performSingleMove advances the simulation for a single elevator by one event:
// - moving: either stop (door cycle + clear requests) or continue (travel one floor)
// - idle/doorOpen: choose direction; either stop and clear at current floor or start moving one floor.
func performSingleMove(s *State, reqs [][]Req) {
	numFloors := len(reqs)

	
	e := withUnassignedRequests(*s, reqs)

	// Callback used by the request-clearing function to mark hall assignments and mutate cab state.
	onClearRequest := func(btn elevio.ButtonType) {
		switch btn {
		case elevio.BT_HallUp, elevio.BT_HallDown:
			reqs[s.State.Floor][int(btn)].AssignedTo = s.ID
		case elevio.BT_Cab:
			s.State.Requests[s.State.Floor][elevio.BT_Cab] = elevator.ReqDeleting
		}
	}

	switch s.State.Behaviour {

	case elevator.EB_Moving:
		if request.ShouldStop(e) {
			e = request.ClearAtCurrentFloorWithCallback(e, onClearRequest)
			s.State.Behaviour = elevator.EB_DoorOpen
			s.Time += constant.DoorOpenDurationMS
		} else {
			next := s.State.Floor + int(s.State.Dirn)
			if next < 0 || next >= numFloors {
				s.State.Dirn = elevio.MD_Stop
				s.State.Behaviour = elevator.EB_Idle
				return
			}
			s.State.Floor = next
			s.Time += constant.TravelDurationMS
		}

	case elevator.EB_Idle, elevator.EB_DoorOpen:
		pair := request.ChooseDirection(e)
		s.State.Dirn = pair.Dirn

		if pair.Dirn == elevio.MD_Stop {
			if request.Here(e) {
				e = request.ClearAtCurrentFloorWithCallback(e, onClearRequest)
				s.Time += constant.DoorOpenDurationMS
				s.State.Behaviour = elevator.EB_DoorOpen
			} else {
				s.State.Behaviour = elevator.EB_Idle
			}
		} else {
			next := s.State.Floor + int(s.State.Dirn)
			if next < 0 || next >= numFloors {
				s.State.Dirn = elevio.MD_Stop
				s.State.Behaviour = elevator.EB_Idle
				return
			}

			s.State.Behaviour = elevator.EB_Moving
			s.State.Floor = next
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
		if elevator.ReqIsActive(e.Requests[f][elevio.BT_Cab]) {
			return true
		}
	}
	return false
}

// unvisitedAreImmediatelyAssignable checks a special fast-path condition where remaining unassigned hall requests
// can be assigned without simulation (no cab calls anywhere, no floors with both hall buttons active, and each
// unassigned request is on a floor where an elevator is currently present).
func unvisitedAreImmediatelyAssignable(reqs [][]Req, states []State) bool {
	for _, s := range states {
		if elevatorHasAnyCab(s.State) {
			return false
		}
	}
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

// assignImmediate assigns any remaining unassigned hall requests directly to an elevator standing at the same floor,
// and advances that elevator's simulated time by one door cycle to reflect serving the request.
func assignImmediate(reqs [][]Req, states []State) {
	for f := range reqs {
		for c := range reqs[f] {
			if !reqs[f][c].Active || reqs[f][c].AssignedTo != "" {
				continue
			}

			for i := range states {
				if states[i].State.Floor != f {
					continue
				}
				if elevatorHasAnyCab(states[i].State) {
					continue
				}

				reqs[f][c].AssignedTo = states[i].ID
				states[i].Time += constant.DoorOpenDurationMS
				break
			}
		}
	}
}

//
// ==========================
// ELEVATOR WRAPPER
// ==========================
//

// withUnassignedRequests builds a temporary elevator snapshot for decision logic, containing:
// - all cab requests from the elevator state
// - hall requests that are either unassigned or already assigned to this elevator ID.
func withUnassignedRequests(s State, reqs [][]Req) elevator.Elevator {
	var e elevator.Elevator
	e.Floor = s.State.Floor
	e.Dirn = s.State.Dirn
	e.Behaviour = s.State.Behaviour

	// Copy cab requests as-is from the elevator snapshot.
	for f := 0; f < constant.NumFloors; f++ {
		e.Requests[f][elevio.BT_Cab] = s.State.Requests[f][elevio.BT_Cab]
	}

	// Include hall requests that are unassigned or assigned to this elevator.
	for f := 0; f < constant.NumFloors; f++ {
		for btn := elevio.ButtonType(0); btn <= elevio.BT_HallDown; btn++ {
			r := reqs[f][int(btn)]
			if r.Active && (r.AssignedTo == "" || r.AssignedTo == s.ID) {
				e.Requests[f][btn] = elevator.ReqConfirmed
			}
		}
	}

	return e
}