package btnEventSplitter

import (
	elevio "project/localElevator/driver"
)

func BtnEventSplitter(btnEvent chan elevio.ButtonEvent,
	hallEvent chan elevio.ButtonEvent,
	cabEvent chan elevio.ButtonEvent) {
	for {
		select {
		case event := <-btnEvent:
			if event.Button == elevio.BT_Cab {
				cabEvent <- event
			} else {
				hallEvent <- event
			}
		}
	}
}
