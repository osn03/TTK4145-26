package network

import (
	"project/Network/TransformElevator"
	"project/Network/peers"
	"project/constant"
	"project/elevator"
	"project/esm"
	"project/elevio"
)

type Msg struct {
	Id        string
	Status    bool
	Floor     int
	Dirn      elevio.MotorDirection
	Requests  [constant.NumFloors][constant.NumButtons]elevator.ReqState
	Behaviour elevator.ElevatorBehavior
}

func TranslateToMsg(elMsg TransformElevator.ElMsg) Msg {
	return Msg{
		Id:        elMsg.Sender,
		Status:    elMsg.Status,
		Floor:     elMsg.Floor,
		Dirn:      elMsg.Dirn,
		Requests:  elMsg.Requests,
		Behaviour: elMsg.Behaviour,
	}
}

func NetworkCum(in <-chan esm.ExternalElevator, outMsg chan<- Msg, outNoder chan<- peers.PeerUpdate) {
	a, b := TransformElevator.Set_up1(<-in)
	go func() {
		for {
			select {
			case msg := <-a:
				outMsg <- TranslateToMsg(msg)
			case noder := <-b:
				outNoder <- noder
			}
		}
	}()
}
