package oassign

import (
	"encoding/json"
	"fmt"
	"os/exec"
	dt "project/dataTypes"
	"project/network/bcast"
	peers "project/network/peers"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type AssignerBehaviour string

const (
	MASTER AssignerBehaviour = "master"
	SLAVE  AssignerBehaviour = "slave"
)

func OrderAssigner(localIP string,
	peerUpdateCh <-chan peers.PeerUpdate,
	costFuncInputCh <-chan dt.CostFuncInputSlice,
	assignedOrdersCh chan<- [dt.N_FLOORS][2]bool) {

	var (
		receiveOrdersCh   = make(chan map[string][dt.N_FLOORS][2]bool)
		transmittOrdersCh = make(chan map[string][dt.N_FLOORS][2]bool)

		hraExecutable     = ""
		assignerBehaviour = MASTER

		localHallOrders       = [dt.N_FLOORS][2]bool{}
		ordersToExternalNodes = map[string][dt.N_FLOORS][2]bool{}

		broadCastTimer   = time.NewTimer(1)
		localOrdersTimer = time.NewTimer(time.Hour)
	)
	broadCastTimer.Stop()
	localOrdersTimer.Stop()

	go bcast.Receiver(dt.MS_PORT, receiveOrdersCh)
	go bcast.Transmitter(dt.MS_PORT, transmittOrdersCh)

	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	for {
		select {
		case peerUpdate := <-peerUpdateCh:
			assignerBehaviour = getAssignerBehaviour(localIP, peerUpdate.Peers)
		case costFuncInput := <-costFuncInputCh:
			input := dt.CostFuncInputSliceToMap(costFuncInput)
			switch assignerBehaviour {
			case SLAVE:
			case MASTER:
				jsonBytes, err := json.Marshal(input)
				if err != nil {
					fmt.Println("json.Marshal error: ", err)
					break
				}
				ret, err := exec.Command(hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
				if err != nil {
					fmt.Println("exec.Command error: ", err)
					fmt.Println(string(ret))
					break
				}
				output := map[string][dt.N_FLOORS][2]bool{}
				err = json.Unmarshal(ret, &output)
				if err != nil {
					fmt.Println("json.Unmarshal error: ", err)
					break
				}
				if newElevHallOrders, ok := output[localIP]; ok {
					localHallOrders = newElevHallOrders
					localOrdersTimer.Reset(1)
				}
				ordersToExternalNodes = output
				broadCastTimer.Reset(1)
			}
		case newOrders := <-receiveOrdersCh:

			if !reflect.DeepEqual(newOrders[localIP], localHallOrders) {
				switch assignerBehaviour {
				case MASTER:
				case SLAVE:
					localHallOrders = newOrders[localIP]
					localOrdersTimer.Reset(1)
				}
			}
		case <-broadCastTimer.C:

			broadCastTimer.Reset(dt.BROADCAST_PERIOD)
			switch assignerBehaviour {
			case SLAVE:
			case MASTER:
				transmittOrdersCh <- ordersToExternalNodes
			}
		case <-localOrdersTimer.C:
			select {
			case assignedOrdersCh <- localHallOrders:
			default:
				localOrdersTimer.Reset(1)
			}
		}
	}
}

func getAssignerBehaviour(localIP string, peers []string) AssignerBehaviour {
	if len(peers) == 0 {
		return MASTER
	}

	localIPArr := strings.Split(localIP, ".")
	LocalLastByte, _ := strconv.Atoi(localIPArr[len(localIPArr)-1])

	maxIP := LocalLastByte
	for _, externalIP := range peers {
		externalIPArr := strings.Split(externalIP, ".")
		externalLastByte, _ := strconv.Atoi(externalIPArr[len(externalIPArr)-1])
		if externalLastByte > maxIP {
			maxIP = externalLastByte
		}
	}

	if maxIP <= LocalLastByte {
		return MASTER
	}
	return SLAVE
}
