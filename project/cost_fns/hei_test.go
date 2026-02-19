package cost_fns

import (
	"reflect"
	"testing"

	"project/constant"
	"project/elevator"
	"project/elevio"
)

// ---- helpers ----

func hallReqs4(f0up, f0dn, f1up, f1dn, f2up, f2dn, f3up, f3dn bool) [][]bool {
	return [][]bool{
		{f0up, f0dn},
		{f1up, f1dn},
		{f2up, f2dn},
		{f3up, f3dn},
	}
}

// E builds an elevator state (snapshot) with cab requests set on given floors.
// Requests are now FSM states (0..3). We mark cab orders as "Confirmed" in tests.
func E(floor int, dir elevio.MotorDirection, beh elevator.ElevatorBehavior, cabFloors ...int) elevator.Elevator {
	var e elevator.Elevator
	e.Floor = floor
	e.Dirn = dir
	e.Behaviour = beh

	for _, f := range cabFloors {
		if f < 0 || f >= constant.NumFloors {
			panic("cab floor out of range in test helper")
		}
		e.Requests[f][elevio.BT_Cab] = elevator.ReqConfirmed
	}
	return e
}

func zerosAssignment() [][]bool {
	a := make([][]bool, constant.NumFloors)
	for f := 0; f < constant.NumFloors; f++ {
		a[f] = []bool{false, false}
	}
	return a
}

func setAssign(a [][]bool, floor int, btn elevio.ButtonType) {
	if btn != elevio.BT_HallUp && btn != elevio.BT_HallDown {
		panic("setAssign only supports hall buttons")
	}
	a[floor][int(btn)] = true
}

// compare only hall outputs [floor][2]
func assertHallResult(t *testing.T, got map[string][][]bool, want map[string][][]bool) {
	t.Helper()

	// Ensure all wanted IDs exist
	for id := range want {
		if _, ok := got[id]; !ok {
			t.Fatalf("missing elevator id %q in result", id)
		}
	}

	// Compare (and ignore any extra ids in got)
	for id, w := range want {
		g := got[id]

		// normalize: if includeCab==true in implementation, result might be [][3].
		// We'll compare only first 2 columns.
		if len(g) != len(w) {
			t.Fatalf("id %q: floor count mismatch got=%d want=%d", id, len(g), len(w))
		}

		for f := 0; f < len(w); f++ {
			if len(g[f]) < 2 || len(w[f]) < 2 {
				t.Fatalf("id %q floor %d: too few columns got=%d want=%d", id, f, len(g[f]), len(w[f]))
			}
			if g[f][0] != w[f][0] || g[f][1] != w[f][1] {
				t.Fatalf("id %q floor %d: got=%v want=%v", id, f, g[f][:2], w[f][:2])
			}
		}
	}
}

// ---- tests mirrored from D ----

func TestOptimalHallRequests_IdleElevatorWins(t *testing.T) {
	states := map[string]elevator.Elevator{
		"1": E(0, elevio.MD_Stop, elevator.EB_Idle /* no cab */),
		"2": E(3, elevio.MD_Down, elevator.EB_DoorOpen, 0 /* cab */),
		"3": E(2, elevio.MD_Up, elevator.EB_Moving, 0, 3 /* cab */),
	}

	hall := hallReqs4(
		false, false,
		true, false,
		false, false,
		false, false,
	)

	got := OptimalHallRequests(hall, states, false)

	want := map[string][][]bool{
		"1": func() [][]bool {
			a := zerosAssignment()
			setAssign(a, 1, elevio.BT_HallUp)
			return a
		}(),
		"2": zerosAssignment(),
		"3": zerosAssignment(),
	}

	assertHallResult(t, got, want)
}

func TestOptimalHallRequests_TwoIdleAtEndsClosestEvenWrongDir(t *testing.T) {
	states := map[string]elevator.Elevator{
		"1": E(0, elevio.MD_Stop, elevator.EB_Idle),
		"2": E(3, elevio.MD_Stop, elevator.EB_Idle),
	}

	hall := hallReqs4(
		false, false,
		false, true,
		true, false,
		false, false,
	)

	got := OptimalHallRequests(hall, states, false)

	want := map[string][][]bool{
		"1": func() [][]bool {
			a := zerosAssignment()
			setAssign(a, 1, elevio.BT_HallDown)
			return a
		}(),
		"2": func() [][]bool {
			a := zerosAssignment()
			setAssign(a, 2, elevio.BT_HallUp)
			return a
		}(),
	}

	assertHallResult(t, got, want)
}

func TestOptimalHallRequests_MovingCloserStillWins(t *testing.T) {
	states := map[string]elevator.Elevator{
		"1": E(0, elevio.MD_Up, elevator.EB_Moving),
		"2": E(3, elevio.MD_Stop, elevator.EB_Idle),
	}

	hall := hallReqs4(
		false, false,
		false, true,
		true, false,
		false, false,
	)

	got := OptimalHallRequests(hall, states, false)

	want := map[string][][]bool{
		"1": func() [][]bool {
			a := zerosAssignment()
			setAssign(a, 1, elevio.BT_HallDown)
			return a
		}(),
		"2": func() [][]bool {
			a := zerosAssignment()
			setAssign(a, 2, elevio.BT_HallUp)
			return a
		}(),
	}

	assertHallResult(t, got, want)
}

func TestOptimalHallRequests_CabAheadChangesDecision(t *testing.T) {
	states := map[string]elevator.Elevator{
		"1": E(0, elevio.MD_Stop, elevator.EB_Idle, 2),
		"2": E(3, elevio.MD_Stop, elevator.EB_Idle),
	}

	hall := hallReqs4(
		false, false,
		false, true,
		true, false,
		false, false,
	)

	got := OptimalHallRequests(hall, states, false)

	want := map[string][][]bool{
		"1": func() [][]bool {
			a := zerosAssignment()
			setAssign(a, 2, elevio.BT_HallUp)
			return a
		}(),
		"2": func() [][]bool {
			a := zerosAssignment()
			setAssign(a, 1, elevio.BT_HallDown)
			return a
		}(),
	}

	assertHallResult(t, got, want)
}

func TestOptimalHallRequests_MovingTowardWinsTie(t *testing.T) {
	states := map[string]elevator.Elevator{
		"27": E(1, elevio.MD_Down, elevator.EB_Moving),
		"20": E(1, elevio.MD_Down, elevator.EB_DoorOpen),
	}

	hall := hallReqs4(
		true, false,
		false, false,
		false, false,
		false, false,
	)

	got := OptimalHallRequests(hall, states, false)

	want := map[string][][]bool{
		"27": func() [][]bool {
			a := zerosAssignment()
			setAssign(a, 0, elevio.BT_HallUp)
			return a
		}(),
		"20": zerosAssignment(),
	}

	assertHallResult(t, got, want)
}

func TestOptimalHallRequests_LexicographicTieBreak(t *testing.T) {
	states := map[string]elevator.Elevator{
		"1": E(1, elevio.MD_Up, elevator.EB_Moving, 0),
		"2": E(1, elevio.MD_Stop, elevator.EB_Idle, 0),
		"3": E(1, elevio.MD_Stop, elevator.EB_Idle, 0),
	}

	hall := hallReqs4(
		true, false,
		false, false,
		false, false,
		false, true,
	)

	got := OptimalHallRequests(hall, states, false)

	want := map[string][][]bool{
		"1": func() [][]bool {
			a := zerosAssignment()
			setAssign(a, 3, elevio.BT_HallDown)
			return a
		}(),
		"2": func() [][]bool {
			a := zerosAssignment()
			setAssign(a, 0, elevio.BT_HallUp)
			return a
		}(),
		"3": zerosAssignment(),
	}

	assertHallResult(t, got, want)
}

func TestOptimalHallRequests_TwoHallSameFloorWithCabAhead_SplitBetweenElevators(t *testing.T) {
	states := map[string]elevator.Elevator{
		"1": E(3, elevio.MD_Down, elevator.EB_Moving, 0),
		"2": E(3, elevio.MD_Down, elevator.EB_Idle),
	}

	hall := hallReqs4(
		false, false,
		true, true,
		false, false,
		false, false,
	)

	got := OptimalHallRequests(hall, states, false)

	want := map[string][][]bool{
		"1": func() [][]bool {
			a := zerosAssignment()
			setAssign(a, 1, elevio.BT_HallDown)
			return a
		}(),
		"2": func() [][]bool {
			a := zerosAssignment()
			setAssign(a, 1, elevio.BT_HallUp)
			return a
		}(),
	}

	if !reflect.DeepEqual(len(got), len(want)) {
		// ignore
	}
	assertHallResult(t, got, want)
}
