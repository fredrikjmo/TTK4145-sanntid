package printing

import (
	"fmt"
	dt "project/dataTypes"
	orderStateHandler "project/distributingSystem/orderStateHandler"
	"sort"
)

// States for hall requests
const (
	STATE_NONE      orderStateHandler.OrderState = "none"
	STATE_NEW       orderStateHandler.OrderState = "new"
	STATE_CONFIRMED orderStateHandler.OrderState = "confirmed"
)

func ElevData_toString(WW map[string]dt.ElevData) string {
	// Create the separator row
	separatorRow := "-------------------------------------------------------------------------------------------------------\n"
	text := "####################################################################################################\n"
	text += "________________________________________ ElevData________________________________________________\n\n"

	// Find the maximum length of any ID.
	ColLen := 14

	// Print the header row.
	text = text + fmt.Sprintf("%-*s | Behavior       | Floor          | Direction      | Cab Requests\n", ColLen, "ID")
	text += separatorRow

	// Print each elevator's data.
	ids := []string{}
	for id := range WW {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		text = text + fmt.Sprintf("%-*s | %-*s | %-*d | %-*s | %v\n",
			ColLen, id,
			ColLen, WW[id].Behavior,
			ColLen, WW[id].Floor+1,
			ColLen, WW[id].Direction,
			WW[id].CabRequests)
	}
	text += separatorRow
	// text = text + fmt.Sprintf("\nHallrequest: %v \n", WW.HallRequests)
	// text += separatorRow
	text = text + "\n####################################################################################################\n"
	return text
}

func NOS_toString(RSM map[string][dt.N_FLOORS][2]orderStateHandler.OrderState) string {
	text := "####################################################################################################\n"
	text = text + "______________________________________________NOS________________________________________________\n\n"
	separatorRow := "-------------------------------------------------------------------------------------------------------\n"

	// Build the header
	header := []string{"ID", "|1_UP", "|1_DWN", "|2_UP", "|2_DWN", "|3_UP", "|3_DWN", "|4_UP", "|4_DWN"}
	var headerStr string
	for _, v := range header {
		headerStr += fmt.Sprintf("%-12s", v)
	}
	headerStr += "\n"
	text = text + headerStr
	text += separatorRow

	// Iterate over each elevator's data

	ids := []string{}
	for id := range RSM {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		text += fmt.Sprintf("%-12s", id)
		for _, state := range RSM[id] {
			switch state[0] {
			case "none":
				text += "|NONE       "
			case "new":
				text += "|NEW        "
			case "confirmed":
				text += "|CONF       "
			default:
				text += "|UNDF       "
			}
			switch state[1] {
			case "none":
				text += "|NONE       "
			case "new":
				text += "|NEW        "
			case "confirmed":
				text += "|CONF       "
			default:
				text += "|UNDF       "
			}
		}
		text += "\n"
	}
	text += separatorRow
	text += "#######################################################################################################\n"

	return text
}
