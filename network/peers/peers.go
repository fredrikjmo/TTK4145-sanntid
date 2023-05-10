package peers

import (
	"fmt"
	"net"
	dt "project/dataTypes"
	"project/network/conn"
	"sort"
	"time"
)

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}

const interval = 15 * time.Millisecond
const timeout = 500 * time.Millisecond

func PeerListHandler(localIP string,
	peerUpdate_OrderAssCh chan<- PeerUpdate,
	peerUpdate_DataDistributor chan<- PeerUpdate,
	peerUpdate_OrderHandler chan<- PeerUpdate,
	isAliveCh <-chan bool,
) {
	var (
		peerUpdateCh = make(chan PeerUpdate)
		peerList     = PeerUpdate{}

		timerDataDistributor = time.NewTimer(time.Hour)
		timerOrderHandler    = time.NewTimer(time.Hour)
		timerOrderAssigner   = time.NewTimer(time.Hour)
	)
	timerDataDistributor.Stop()
	timerOrderHandler.Stop()
	timerOrderAssigner.Stop()

	go Transmitter(dt.PEER_LIST_PORT, localIP, isAliveCh)
	go Receiver(dt.PEER_LIST_PORT, peerUpdateCh)

	for {
		select {
		case peerList = <-peerUpdateCh:
			timerDataDistributor.Reset(1)
			timerOrderHandler.Reset(1)
			timerOrderAssigner.Reset(1)
		case <-timerDataDistributor.C:
			select {
			case peerUpdate_DataDistributor <- peerList:
			default:
				timerDataDistributor.Reset(1)
			}
		case <-timerOrderHandler.C:
			select {
			case peerUpdate_OrderHandler <- peerList:
			default:
				timerOrderHandler.Reset(1)
			}
		case <-timerOrderAssigner.C:
			select {
			case peerUpdate_OrderAssCh <- peerList:
			default:
				timerOrderAssigner.Reset(1)
			}
		}
	}
}

func Transmitter(port int, id string, transmitEnable <-chan bool) {

	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	enable := true
	for {
		select {
		case enable = <-transmitEnable:
		case <-time.After(interval):
		}
		if enable {
			conn.WriteTo([]byte(id), addr)
		}
	}
}

func Receiver(port int, peerUpdateCh chan<- PeerUpdate) {

	var buf [1024]byte
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	conn := conn.DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		id := string(buf[:n])

		// Adding new connection
		p.New = ""
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				p.New = id
				updated = true
			}

			lastSeen[id] = time.Now()
		}

		// Removing dead connection
		p.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Now().Sub(v) > timeout {
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}

		// Sending update
		if updated {
			p.Peers = make([]string, 0, len(lastSeen))

			for k, _ := range lastSeen {
				p.Peers = append(p.Peers, k)
			}

			sort.Strings(p.Peers)
			sort.Strings(p.Lost)
			peerUpdateCh <- p
		}
	}
}
