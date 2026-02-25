package network

import (
	"project/constant"
	"project/network/TransformElevator"
	"project/network/peers"
	"project/types"
)

type Msg struct {
	Id        string
	Status    bool
	Floor     int
	Dirn      types.MotorDirection
	Requests  [constant.NumFloors][constant.NumButtons]types.ReqState
	Behaviour types.ElevatorBehavior
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

func NetworkCum(in <-chan types.ExternalElevator, outMsg chan<- Msg, outNoder chan<- peers.PeerUpdate, localid string) {
	a, b := TransformElevator.Set_up1(<-in, localid)
	for {
		select {
		case msg := <-a:
			outMsg <- TranslateToMsg(msg)
		case noder := <-b:
			outNoder <- noder
		}
	}
}
