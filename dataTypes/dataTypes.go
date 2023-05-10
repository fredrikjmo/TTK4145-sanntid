package dt

import (
	"time"
)

// ****** ELEVATOR CONSTANTS ********
const (
	N_FLOORS  = 4
	N_BUTTONS = 3
	DOOR_OPEN_TIME = 3 * time.Second
	WD_TIMEOUT = 5 * time.Second
)

// ****** NETWORK CONSTANTS ********
const (
	MS_PORT                             = 15480
	PEER_LIST_PORT                      = 15489
	ORDERSTATE_PORT                     = 15488
	DATA_DISTRIBUTOR_PORT               = 15487
	BROADCAST_PERIOD      time.Duration = 100 * time.Millisecond
)

// ****** DATA TYPES ********
type ElevData struct {
	Behavior    string         `json:"behaviour"`
	Floor       int            `json:"floor"`
	Direction   string         `json:"direction"`
	CabRequests [N_FLOORS]bool `json:"cabRequests"`
}

type NodeInfo struct {
	IP   string   `json:"ip"`
	Data ElevData `json:"data"`
}

type AllNodeInfoWithSenderIP struct {
	SenderIP    string     `json:"senderIp"`
	AllNodeInfo []NodeInfo `json:"allNodeInfo"`
}

type CostFuncInput struct {
	HallRequests [N_FLOORS][2]bool   `json:"hallRequests"`
	States       map[string]ElevData `json:"states"`
}

type CostFuncInputSlice struct {
	HallRequests [N_FLOORS][2]bool `json:"hallRequests"`
	States       []NodeInfo        `json:"states"`
}

// ****** DATA TYPE CONVERTING FUNCTIONS ********
func NodeInfoMapToSlice(nodesInfoMap map[string]ElevData) []NodeInfo {
	nodesInfoSlice := []NodeInfo{}
	for ip, elevData := range nodesInfoMap {
		nodesInfoSlice = append(nodesInfoSlice, NodeInfo{IP: ip, Data: elevData})
	}
	return nodesInfoSlice
}

func CostFuncInputSliceToMap(costFuncInputSlice CostFuncInputSlice) CostFuncInput {
	allNodeInfo := map[string]ElevData{}
	for _, nodeInfo := range costFuncInputSlice.States {
		allNodeInfo[nodeInfo.IP] = nodeInfo.Data
	}
	costFuncInput := CostFuncInput{HallRequests: costFuncInputSlice.HallRequests, States: allNodeInfo}
	return costFuncInput
}
