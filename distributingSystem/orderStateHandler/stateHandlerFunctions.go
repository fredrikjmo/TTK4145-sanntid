package orderStateHandler

import (
	dt "project/dataTypes"
	"project/network/peers"
)

func orderCanBeConfirmed(floor, btn int,
	AllNodeOrderStates map[string][dt.N_FLOORS][2]OrderState) bool {
	for _, nodeStates := range AllNodeOrderStates {
		if nodeStates[floor][btn] != STATE_NEW {
			return false
		}
	}
	return true
}

func getNewState(inputState, currentState OrderState) OrderState {
	switch inputState {
	case STATE_NONE:
		if currentState == STATE_CONFIRMED {
			return STATE_NONE
		}
	case STATE_NEW:
		if currentState == STATE_NONE {
			return STATE_NEW

		}
	case STATE_CONFIRMED:
		if currentState == STATE_NEW {
			return STATE_CONFIRMED
		}
	}
	return currentState
}

func orderStatesToBool(orderStates [dt.N_FLOORS][2]OrderState) [dt.N_FLOORS][2]bool {
	hallOrders := [dt.N_FLOORS][2]bool{}
	for floor := range orderStates {
		for btn, state := range orderStates[floor] {
			if state == STATE_CONFIRMED {
				hallOrders[floor][btn] = true
			}
		}
	}
	return hallOrders
}

func addAliveNodes(peerList peers.PeerUpdate,
	AllNodeOrderStates map[string][dt.N_FLOORS][2]OrderState,
) map[string][dt.N_FLOORS][2]OrderState {
	outputMap := make(map[string][dt.N_FLOORS][2]OrderState)
	for key, value := range AllNodeOrderStates {
		outputMap[key] = value
	}
	for _, nodeIP := range peerList.Peers {
		if _, nodeOrdersSaved := outputMap[nodeIP]; !nodeOrdersSaved {
			outputMap[nodeIP] = [dt.N_FLOORS][2]OrderState{
				{STATE_NONE, STATE_NONE},
				{STATE_NONE, STATE_NONE},
				{STATE_NONE, STATE_NONE},
				{STATE_NONE, STATE_NONE}}
		}
	}
	return outputMap
}

func removeDeadNodes(peerList peers.PeerUpdate,
	AllNodeOrderStates map[string][dt.N_FLOORS][2]OrderState,
	localIP string,
) map[string][dt.N_FLOORS][2]OrderState {
	outputMap := make(map[string][dt.N_FLOORS][2]OrderState)
	for IP := range AllNodeOrderStates {
		if contains(peerList.Peers, IP) || IP == localIP {
			outputMap[IP] = AllNodeOrderStates[IP]
		}
	}
	return outputMap
}

func withdrawOrderConfirmations(peerList peers.PeerUpdate,
	AllNodeOrderStates map[string][dt.N_FLOORS][2]OrderState,
) map[string][dt.N_FLOORS][2]OrderState {
	outputMap := make(map[string][dt.N_FLOORS][2]OrderState)
	for key, value := range AllNodeOrderStates {
		outputMap[key] = value
	}
	for _, nodeIP := range peerList.Peers {
		orderStates, nodeOrdersSaved := outputMap[nodeIP]
		if nodeOrdersSaved {
			for floor := range orderStates {
				for btn, state := range orderStates[floor] {
					switch state {
					case STATE_NONE:
					case STATE_NEW:
					case STATE_CONFIRMED:
						orderStates[floor][btn] = STATE_NEW
						outputMap[nodeIP] = orderStates
					}
				}
			}
		}
	}
	return outputMap
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
