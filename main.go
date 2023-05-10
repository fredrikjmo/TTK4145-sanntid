package main

import (
	dt "project/dataTypes"
	elevDataDistributor "project/distributingSystem/elevDataDistributor"
	orderStateHandler "project/distributingSystem/orderStateHandler"
	elevfsm "project/localElevator/FSM"
	btnEventSplitter "project/localElevator/btnEventSplitter"
	elevio "project/localElevator/driver"
	localip "project/network/localip"
	peers "project/network/peers"
	oassign "project/orderAssigner"
	"time"
)

var (
	floorCh           = make(chan int)
	obstrCh           = make(chan bool)
	localElevDataCh        = make(chan dt.ElevData)
	buttonEventCh     = make(chan elevio.ButtonEvent)
	cabButtonEventCh  = make(chan elevio.ButtonEvent)
	hallButtonEventCh = make(chan elevio.ButtonEvent)

	costFuncInputCh     = make(chan dt.CostFuncInputSlice)
	executedHallOrderCh = make(chan elevio.ButtonEvent)
	assignedOrdersCh    = make(chan [dt.N_FLOORS][2]bool)
	orderStatesToBoolCh = make(chan [dt.N_FLOORS][2]bool)
	initCabRequestsCh   = make(chan [dt.N_FLOORS]bool)

	peerUpdate_OrderAssCh        = make(chan peers.PeerUpdate)
	peerUpdate_DataDistributorCh = make(chan peers.PeerUpdate)
	peerUpdate_OrderHandlerCh    = make(chan peers.PeerUpdate)
	isAliveCh                    = make(chan bool)
)

func main() {
	localIP, _ := localip.LocalIP()
	elevio.Init("localhost:15657", dt.N_FLOORS)
	elevio.ClearAllLights()

	go elevio.PollFloorSensor(floorCh)
	go elevio.PollButtons(buttonEventCh)
	go elevio.PollObstructionSwitch(obstrCh)
	go btnEventSplitter.BtnEventSplitter(buttonEventCh, hallButtonEventCh, cabButtonEventCh)

	go peers.PeerListHandler(localIP,
		peerUpdate_OrderAssCh,
		peerUpdate_DataDistributorCh,
		peerUpdate_OrderHandlerCh,
		isAliveCh)

	go oassign.OrderAssigner(localIP,
		peerUpdate_OrderAssCh,
		costFuncInputCh,
		assignedOrdersCh)

	go elevDataDistributor.DataDistributor(localIP,
		peerUpdate_DataDistributorCh,
		initCabRequestsCh,
		localElevDataCh,
		orderStatesToBoolCh,
		costFuncInputCh)

	go orderStateHandler.OrderStateHandler(localIP,
		peerUpdate_OrderHandlerCh,
		hallButtonEventCh,
		executedHallOrderCh,
		orderStatesToBoolCh)

	time.Sleep(time.Millisecond * 40)

	go elevfsm.FSM(floorCh,
		obstrCh,
		cabButtonEventCh,
		initCabRequestsCh,
		assignedOrdersCh,
		executedHallOrderCh,
		localElevDataCh,
		isAliveCh)

	select {}
}
