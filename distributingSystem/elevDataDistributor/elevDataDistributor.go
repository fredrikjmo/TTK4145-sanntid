package elevDataDistributor

import (
	dt "project/dataTypes"
	"project/network/bcast"
	"project/network/peers"
	"reflect"
	"time"
)

func DataDistributor(localIP string,
	peerUpdateCh <-chan peers.PeerUpdate,
	initCabRequestsCh chan<- [dt.N_FLOORS]bool,
	localElevDataCh <-chan dt.ElevData,
	orderStatesToBoolCh <-chan [dt.N_FLOORS][2]bool,
	costFuncInputCh chan<- dt.CostFuncInputSlice,
) {

	var (
		receiveCh   = make(chan dt.AllNodeInfoWithSenderIP)
		transmittCh = make(chan dt.AllNodeInfoWithSenderIP)

		initTimer      = time.NewTimer(time.Hour)
		costFuncTimer  = time.NewTimer(time.Hour)
		broadCastTimer = time.NewTimer(time.Hour)

		peerList           = peers.PeerUpdate{}
		allElevData        = map[string]dt.ElevData{}
		costFuncInputSlice = dt.CostFuncInputSlice{}
	)

	initTimer.Stop()
	costFuncTimer.Stop()
	broadCastTimer.Stop()

	go bcast.Receiver(dt.DATA_DISTRIBUTOR_PORT, receiveCh)
	go bcast.Transmitter(dt.DATA_DISTRIBUTOR_PORT, transmittCh)

	initTimer.Reset(time.Second * 3)
	initCabRequests := [dt.N_FLOORS]bool{}
initialization:
	for {
		select {
		case receivedData := <-receiveCh:
			senderIP := receivedData.SenderIP
			allNodesInfo := receivedData.AllNodeInfo

			for _, receivedNodeInfo := range allNodesInfo {
				if receivedElevData := receivedNodeInfo.Data; receivedNodeInfo.IP == localIP {
					for floor, receivedOrder := range receivedElevData.CabRequests {
						initCabRequests[floor] = initCabRequests[floor] || receivedOrder
					}
				} else if receivedNodeInfo.IP == senderIP {
					allElevData[senderIP] = receivedElevData
				}
			}
		case <-initTimer.C:
			initCabRequestsCh <- initCabRequests
			allElevData[localIP] = <-localElevDataCh
			broadCastTimer.Reset(1)
			break initialization
		}
	}

	for {
		select {
		case peerList = <-peerUpdateCh:
		case allElevData[localIP] = <-localElevDataCh:
		case hallRequests := <-orderStatesToBoolCh:
			aliveNodesData := []dt.NodeInfo{{IP: localIP, Data: allElevData[localIP]}}
			for _, nodeIP := range peerList.Peers {
				nodeData, nodeDataExists := allElevData[nodeIP]
				if nodeDataExists {
					aliveNodesData = append(aliveNodesData, dt.NodeInfo{IP: nodeIP, Data: nodeData})
				}
			}
			costFuncInputSlice = dt.CostFuncInputSlice{
				HallRequests: hallRequests,
				States:       aliveNodesData,
			}
			costFuncTimer.Reset(1)
		case receivedData := <-receiveCh:
			senderIP := receivedData.SenderIP
			allNodesInfo := receivedData.AllNodeInfo

			if senderIP == localIP {
				break
			}
			for _, receivedNodeInfo := range allNodesInfo {
				if receivedNodeInfo.IP == senderIP && !reflect.DeepEqual(allElevData[senderIP], receivedNodeInfo.Data) {
					allElevData[senderIP] = receivedNodeInfo.Data
				}
			}
		case <-broadCastTimer.C:
			transmittCh <- dt.AllNodeInfoWithSenderIP{SenderIP: localIP, AllNodeInfo: dt.NodeInfoMapToSlice(allElevData)}
			broadCastTimer.Reset(dt.BROADCAST_PERIOD)

		case <-costFuncTimer.C:
			select {
			case costFuncInputCh <- costFuncInputSlice:
			default:
				costFuncTimer.Reset(1)
			}
		}
	}
}
