package orderStateHandler

import (
	dt "project/dataTypes"
	elevio "project/localElevator/driver"
	"project/network/bcast"
	"project/network/peers"
	"reflect"
	"time"
)

type NodeOrderStates struct {
	IP          string                     `json:"ip"`
	OrderStates [dt.N_FLOORS][2]OrderState `json:"orderStates"`
}

type OrderState string

const (
	STATE_NONE      OrderState = "none"
	STATE_NEW       OrderState = "new"
	STATE_CONFIRMED OrderState = "confirmed"
)

func OrderStateHandler(localIP string,
	peerUpdateCh <-chan peers.PeerUpdate,
	hallButtonEventCh <-chan elevio.ButtonEvent,
	executedHallOrderCh <-chan elevio.ButtonEvent,
	orderStatesToBoolCh chan<- [dt.N_FLOORS][2]bool,
) {
	var (
		peerList           = peers.PeerUpdate{}
		AllNodeOrderStates = map[string][dt.N_FLOORS][2]OrderState{}

		broadCastTimer    = time.NewTimer(time.Hour)
		statesToBoolTimer = time.NewTimer(time.Hour)

		receiveCh  = make(chan NodeOrderStates)
		transmitCh = make(chan NodeOrderStates)
	)
	broadCastTimer.Stop()
	statesToBoolTimer.Stop()

	go bcast.Receiver(dt.ORDERSTATE_PORT, receiveCh)
	go bcast.Transmitter(dt.ORDERSTATE_PORT, transmitCh)

	for {
		select {
		case peerList = <-peerUpdateCh:
			newNodeOrderStates := removeDeadNodes(peerList, AllNodeOrderStates, localIP)
			newNodeOrderStates = addAliveNodes(peerList, newNodeOrderStates)
			newNodeOrderStates = withdrawOrderConfirmations(peerList, newNodeOrderStates)
			AllNodeOrderStates = newNodeOrderStates
			broadCastTimer.Reset(dt.BROADCAST_PERIOD)
			statesToBoolTimer.Reset(1)

		case receivedData := <-receiveCh:
			receivedStates := receivedData.OrderStates
			senderIP := receivedData.IP

			if senderIP == localIP || reflect.DeepEqual(receivedStates, AllNodeOrderStates[senderIP]) {
				break
			}
			AllNodeOrderStates[senderIP] = receivedStates

			LocalStates := AllNodeOrderStates[localIP]
			for floor := range receivedStates {
				for btn, inputBtnState := range receivedStates[floor] {
					LocalStates[floor][btn] = getNewState(inputBtnState, LocalStates[floor][btn])
					switch LocalStates[floor][btn] {
					case STATE_NEW:
					case STATE_CONFIRMED:
						elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
					case STATE_NONE:
						elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
					}
				}
			}
			AllNodeOrderStates[localIP] = LocalStates
			statesToBoolTimer.Reset(1)

		case BtnPress := <-hallButtonEventCh:
			LocalStates := AllNodeOrderStates[localIP]
			if LocalStates[BtnPress.Floor][BtnPress.Button] == STATE_NONE {
				LocalStates[BtnPress.Floor][BtnPress.Button] = STATE_NEW
				AllNodeOrderStates[localIP] = LocalStates
				statesToBoolTimer.Reset(1)
			}
		case executedOrder := <-executedHallOrderCh:
			LocalStates := AllNodeOrderStates[localIP]
			if LocalStates[executedOrder.Floor][executedOrder.Button] == STATE_CONFIRMED {
				LocalStates[executedOrder.Floor][executedOrder.Button] = STATE_NONE
				AllNodeOrderStates[localIP] = LocalStates
				statesToBoolTimer.Reset(1)
				elevio.SetButtonLamp(executedOrder.Button, executedOrder.Floor, false)
			}
		case <-broadCastTimer.C:
			transmitCh <- NodeOrderStates{IP: localIP, OrderStates: AllNodeOrderStates[localIP]}
			broadCastTimer.Reset(dt.BROADCAST_PERIOD)

		case <-statesToBoolTimer.C:
			select {
			case orderStatesToBoolCh <- orderStatesToBool(AllNodeOrderStates[localIP]):
			default:
				statesToBoolTimer.Reset(1)
			}
		}
		// Confirm Local State based on AllNodeOrderStates
		LocalStates := AllNodeOrderStates[localIP]
		for floor := 0; floor < dt.N_FLOORS; floor++ {
			for btn := 0; btn < dt.N_BUTTONS-1; btn++ {
				if orderCanBeConfirmed(floor, btn, AllNodeOrderStates) {
					LocalStates[floor][btn] = STATE_CONFIRMED
					elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
				}
			}
		}
		if AllNodeOrderStates[localIP] != LocalStates {
			AllNodeOrderStates[localIP] = LocalStates
			statesToBoolTimer.Reset(1)
		}
	}
}
