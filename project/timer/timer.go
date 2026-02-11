package timer

import "time"

var (
	timerEndTime time.Time
	timerActive  bool
)

func GetWallTime() time.Time {
	return time.Now()
}

func Start(duration float64) {
	timerEndTime = GetWallTime().Add(
		time.Duration(duration * float64(time.Second)),
	)
	timerActive = true
}

func Stop() {
	timerActive = false
}

func TimedOut() bool {
	return timerActive && GetWallTime().After(timerEndTime)
}

//test
