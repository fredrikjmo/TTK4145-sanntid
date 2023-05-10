package elevfsm

import (
	dt "project/dataTypes"
	elevio "project/localElevator/driver"
)

func requests_mergeHallAndCab(hallRequests [dt.N_FLOORS][2]bool, cabRequests [dt.N_FLOORS]bool) [dt.N_FLOORS][dt.N_BUTTONS]bool {
	var requests [dt.N_FLOORS][dt.N_BUTTONS]bool
	for i := range requests {
		requests[i] = [dt.N_BUTTONS]bool{hallRequests[i][0], hallRequests[i][1], cabRequests[i]}
	}
	return requests
}

func requests_above(e Elevator) bool {
	requests := requests_mergeHallAndCab(e.HallRequests, e.CabRequests)
	for f := e.Floor + 1; f < dt.N_FLOORS; f++ {
		for btn := 0; btn < dt.N_BUTTONS; btn++ {
			if requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_below(e Elevator) bool {
	requests := requests_mergeHallAndCab(e.HallRequests, e.CabRequests)
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < dt.N_BUTTONS; btn++ {
			if requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_here(e Elevator) bool {
	requests := requests_mergeHallAndCab(e.HallRequests, e.CabRequests)
	for btn := 0; btn < dt.N_BUTTONS; btn++ {
		if requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

func requests_chooseDirection(e Elevator) (elevio.MotorDirection, ElevatorBehaviour) {
	switch e.Dirn {
	case elevio.MD_Up:
		if requests_above(e) {
			return elevio.MD_Up, EB_MOVING
		}
		if requests_here(e) {
			return elevio.MD_Down, EB_DOOR_OPEN
		}
		if requests_below(e) {
			return elevio.MD_Down, EB_MOVING
		}
		return elevio.MD_Stop, EB_IDLE

	case elevio.MD_Down:
		if requests_below(e) {
			return elevio.MD_Down, EB_MOVING
		}
		if requests_here(e) {
			return elevio.MD_Up, EB_DOOR_OPEN
		}
		if requests_above(e) {
			return elevio.MD_Up, EB_MOVING
		}
		return elevio.MD_Stop, EB_IDLE

	case elevio.MD_Stop: // there should only be one request in the Stop case. Checking up or down first is arbitrary.
		if requests_here(e) {
			return elevio.MD_Stop, EB_DOOR_OPEN
		}
		if requests_above(e) {
			return elevio.MD_Up, EB_MOVING
		}
		if requests_below(e) {
			return elevio.MD_Down, EB_MOVING
		}
		return elevio.MD_Stop, EB_IDLE

	default:
		return elevio.MD_Stop, EB_IDLE
	}
}

func requests_shouldStop(e Elevator) bool {
	requests := requests_mergeHallAndCab(e.HallRequests, e.CabRequests)
	switch e.Dirn {
	case elevio.MD_Down:
		return requests[e.Floor][elevio.BT_HallDown] ||
			requests[e.Floor][elevio.BT_Cab] ||
			!requests_below(e)
	case elevio.MD_Up:
		return requests[e.Floor][elevio.BT_HallUp] ||
			requests[e.Floor][elevio.BT_Cab] ||
			!requests_above(e)
	case elevio.MD_Stop:
		return true
	default:
		return true
	}
}

func requests_getExecutedHallOrder(e Elevator) elevio.ButtonEvent {
	requests := requests_mergeHallAndCab(e.HallRequests, e.CabRequests)
	switch e.Dirn {
	case elevio.MD_Stop:
		if requests[e.Floor][elevio.BT_HallUp] && !requests_above(e) {
			return elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp}
		} else if requests[e.Floor][elevio.BT_HallDown] && !requests_below(e) {
			return elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown}
		} else if requests[e.Floor][elevio.BT_HallDown] && requests_below(e) {
			return elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown}
		} else {
			return elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp}
		}
	case elevio.MD_Up:
		if !requests[e.Floor][elevio.BT_HallUp] && !requests_above(e) {
			return elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown}
		} else {
			return elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp}
		}
	case elevio.MD_Down:
		if !requests[e.Floor][elevio.BT_HallDown] && !requests_below(e) {
			return elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp}
		} else {
			return elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown}
		}
	default:
		panic("Elevator request error")
	}
}
