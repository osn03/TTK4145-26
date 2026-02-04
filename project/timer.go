package timer

import "time"

var (
	timerEndTime time.Time
	timerActive  bool
)

// tilsvarer get_wall_time
func getWallTime() time.Time {
	return time.Now()
}

// starter timeren med varighet i sekunder
func TimerStart(duration float64) {
	timerEndTime = getWallTime().Add(
		time.Duration(duration * float64(time.Second)),
	)
	timerActive = true
}

// stopper timeren
func TimerStop() {
	timerActive = false
}

// returnerer true hvis timeren er aktiv og tiden er ute
func TimerTimedOut() bool {
	return timerActive && getWallTime().After(timerEndTime)
}

//test
