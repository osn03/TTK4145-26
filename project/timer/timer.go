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
		time.Duration(duration * float64(time.Millisecond)),
	)
	timerActive = true
}

func Stop() {
	timerActive = false
}

func TimedOut(receiver chan<- bool) {
	for {
		if timerActive && GetWallTime().After(timerEndTime) {
				receiver <- true
			}
		time.Sleep(10 * time.Millisecond)
	}
}


//test
