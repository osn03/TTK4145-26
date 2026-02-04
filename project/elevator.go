package elevator

import{
	"fmt"
}

type Dirn int;

const{
	
}


func ElevatorFloorSensor() int {
	// Calls the hardware to get the current floor
	return elevatorHardwareGetFloorSensorSignal()
}

func ElevatorRequestButton(floor int, b Button) int {
	// Returns 1 if button pressed, 0 otherwise
	return elevatorHardwareGetButtonSignal(int(b), floor)
}

func ElevatorStopButton() int {
	return elevatorHardwareGetStopSignal()
}

func ElevatorObstruction() int {
	return elevatorHardwareGetObstructionSignal()
}

// ---- Output / actuators ----

func ElevatorFloorIndicator(floor int) {
	elevatorHardwareSetFloorIndicator(floor)
}

func ElevatorRequestButtonLight(floor int, b Button, on bool) {
	val := 0
	if on {
		val = 1
	}
	elevatorHardwareSetButtonLamp(int(b), floor, val)
}

func ElevatorDoorLight(on bool) {
	val := 0
	if on {
		val = 1
	}
	elevatorHardwareSetDoorOpenLamp(val)
}

func ElevatorStopButtonLight(on bool) {
	val := 0
	if on {
		val = 1
	}
	elevatorHardwareSetStopLamp(val)
}

func ElevatorMotorDirection(d Dirn) {
	elevatorHardwareSetMotorDirection(int(d))
}