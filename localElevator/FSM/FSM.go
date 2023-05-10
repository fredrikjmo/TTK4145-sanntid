package elevfsm

import (
	dt "project/dataTypes"
	elevio "project/localElevator/driver"
	"time"
)

type ElevatorBehaviour string

const (
	EB_DOOR_OPEN ElevatorBehaviour = "doorOpen"
	EB_MOVING    ElevatorBehaviour = "moving"
	EB_IDLE      ElevatorBehaviour = "idle"
)

type Elevator struct {
	Floor            int
	Dirn             elevio.MotorDirection
	Behaviour        ElevatorBehaviour
	CabRequests      [dt.N_FLOORS]bool
	HallRequests     [dt.N_FLOORS][2]bool
	doorOpenDuration time.Duration
}
type WD_Role string

const (
	WD_ALIVE     WD_Role = "alive"
	WD_DEAD      WD_Role = "dead"
	watchDogTime         = dt.WD_TIMEOUT
)

func FSM(
	floorCh <-chan int,
	obstrCh <-chan bool,
	cabButtonEventCh <-chan elevio.ButtonEvent,
	initCabRequestsCh <-chan [dt.N_FLOORS]bool,
	assignedOrdersCh <-chan [dt.N_FLOORS][2]bool,
	executedHallOrderCh chan<- elevio.ButtonEvent,
	localElevDataCh chan<- dt.ElevData,
	isAliveCh chan<- bool) {

	var (
		watchDogTimer  = time.NewTimer(time.Hour)
		watchDogStatus = WD_ALIVE

		executedHallOrder      = elevio.ButtonEvent{}
		executedHallOrderTimer = time.NewTimer(time.Hour)

		obstr         = false
		doorOpenTimer = time.NewTimer(time.Hour)
		elevDataTimer = time.NewTimer(time.Hour)

		e = Elevator{
			Floor:       -1,
			Dirn:        elevio.MD_Stop,
			Behaviour:   EB_IDLE,
			CabRequests: [dt.N_FLOORS]bool{false, false, false, false},
			HallRequests: [dt.N_FLOORS][2]bool{
				{false, false},
				{false, false},
				{false, false},
				{false, false}},
			doorOpenDuration: dt.DOOR_OPEN_TIME,
		}
	)
	doorOpenTimer.Stop()
	elevDataTimer.Stop()
	watchDogTimer.Stop()
	executedHallOrderTimer.Stop()

initialization:
	for {
		select {
		case e.Floor = <-floorCh:
			elevio.SetMotorDirection(elevio.MD_Stop)

			e.CabRequests = <-initCabRequestsCh
			for floor, order := range e.CabRequests {
				elevio.SetButtonLamp(elevio.BT_Cab, floor, order)
			}
			e.Dirn, e.Behaviour = requests_chooseDirection(e)
			elevDataTimer.Reset(1)

			switch e.Behaviour {
			case EB_IDLE:
			case EB_DOOR_OPEN:
				elevio.SetDoorOpenLamp(true)
				doorOpenTimer.Reset(e.doorOpenDuration)

			case EB_MOVING:
				elevio.SetMotorDirection(e.Dirn)
			}
			break initialization
		default:
			elevio.SetMotorDirection(elevio.MD_Down)
			time.Sleep(10 * time.Millisecond)
		}
	}

	for {
		select {
		case e.HallRequests = <-assignedOrdersCh:
			switch e.Behaviour {
			case EB_DOOR_OPEN:
			case EB_MOVING:
			case EB_IDLE:
				e.Dirn, e.Behaviour = requests_chooseDirection(e)
				elevDataTimer.Reset(1)

				switch e.Behaviour {
				case EB_IDLE:
				case EB_DOOR_OPEN:
					elevio.SetDoorOpenLamp(true)
					doorOpenTimer.Reset(e.doorOpenDuration)

				case EB_MOVING:
					elevio.SetMotorDirection(e.Dirn)
				}
			}
		case cabButtonEvent := <-cabButtonEventCh:
			e.CabRequests[cabButtonEvent.Floor] = true
			elevio.SetButtonLamp(elevio.BT_Cab, cabButtonEvent.Floor, true)
			elevDataTimer.Reset(1)
			switch e.Behaviour {
			case EB_DOOR_OPEN:
			case EB_MOVING:
			case EB_IDLE:
				e.Dirn, e.Behaviour = requests_chooseDirection(e)

				switch e.Behaviour {
				case EB_IDLE:
				case EB_DOOR_OPEN:
					elevio.SetDoorOpenLamp(true)
					doorOpenTimer.Reset(e.doorOpenDuration)

				case EB_MOVING:
					elevio.SetMotorDirection(e.Dirn)
				}
			}
		case e.Floor = <-floorCh:
			elevio.SetFloorIndicator(e.Floor)
			switch e.Behaviour {
			case EB_IDLE:
			case EB_DOOR_OPEN:
			case EB_MOVING:
				elevDataTimer.Reset(1)
				if requests_shouldStop(e) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevio.SetDoorOpenLamp(true)
					e.Behaviour = EB_DOOR_OPEN
					if !obstr {
						doorOpenTimer.Reset(e.doorOpenDuration)
					}
				}
			}
		case <-doorOpenTimer.C:
			switch e.Behaviour {
			case EB_IDLE:
			case EB_MOVING:
			case EB_DOOR_OPEN:
				e.CabRequests[e.Floor] = false
				elevio.SetButtonLamp(elevio.BT_Cab, e.Floor, false)
				executedHallOrder = requests_getExecutedHallOrder(e)
				e.HallRequests[executedHallOrder.Floor][executedHallOrder.Button] = false
				executedHallOrderTimer.Reset(1)

				e.Dirn, e.Behaviour = requests_chooseDirection(e)
				elevDataTimer.Reset(1)
				switch e.Behaviour {
				case EB_DOOR_OPEN:
					doorOpenTimer.Reset(e.doorOpenDuration)
				case EB_MOVING, EB_IDLE:
					elevio.SetDoorOpenLamp(false)
					elevio.SetMotorDirection(e.Dirn)
				}
			}
		case obstr = <-obstrCh:
			if obstr {
				doorOpenTimer.Stop()
			}
			switch e.Behaviour {
			case EB_IDLE:
			case EB_MOVING:
			case EB_DOOR_OPEN:
				if !obstr {
					doorOpenTimer.Reset(e.doorOpenDuration)
				}
			}
		case <-elevDataTimer.C:
			select {
			case localElevDataCh <- getElevatorData(e):
				watchDogTimer.Reset(watchDogTime)
				switch watchDogStatus {
				case WD_ALIVE:
				case WD_DEAD:
					isAliveCh <- true
					watchDogStatus = WD_ALIVE
				}
			default:
				elevDataTimer.Reset(1)
			}
		case <-watchDogTimer.C:
			switch e.Behaviour {
			case EB_IDLE:
			case EB_MOVING, EB_DOOR_OPEN:
				isAliveCh <- false
				watchDogStatus = WD_DEAD
			}
		case <-executedHallOrderTimer.C:
			select {
			case executedHallOrderCh <- executedHallOrder:
			default:
				executedHallOrderTimer.Reset(1)
			}
		}
	}
}

func getElevatorData(e Elevator) dt.ElevData {
	dirnToString := map[elevio.MotorDirection]string{
		elevio.MD_Down: "down",
		elevio.MD_Up:   "up",
		elevio.MD_Stop: "stop"}

	return dt.ElevData{
		Behavior:    string(e.Behaviour),
		Floor:       e.Floor,
		Direction:   dirnToString[e.Dirn],
		CabRequests: e.CabRequests}
}
