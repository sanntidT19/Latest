package toplayer

import (
	. "chansnstructs"
	. "network"
	//. "sync"
	. "encoding/json"
	"fmt"
	. "net"
	"os"
	"os/exec"
	//. "statemachine"
	"sync"
	"time"
)

var InLogicChans InternalLogicChannels
var InteruptChan chan os.Signal
var RestartSystemChan chan map[*UDPAddr]*[N_FLOORS][2]bool

type InternalLogicChannels struct {
	ToStateUpdater          chan IpState
	ToTopLogicOrderChan     chan IpOrderMessage
	ToMasterUpdateStateChan chan IpState
	ExternalListIsUpdated   chan bool
	NetworkInitRespChan     chan *UDPAddr
}

func internal_logic_channels_init() {
	InLogicChans.ToStateUpdater = make(chan IpState)
	InLogicChans.ToTopLogicOrderChan = make(chan IpOrderMessage)
	InLogicChans.ToMasterUpdateStateChan = make(chan IpState)
	InLogicChans.ExternalListIsUpdated = make(chan bool)
	InLogicChans.NetworkInitRespChan = make(chan *UDPAddr)
}

func Active_slave() {
	fmt.Println("start slave")
	var localAddr *UDPAddr
	//s := Slave{}

	//s.Ip = localAddr
	//buf := make([]byte, 1024)

	fmt.Println("connection init")

	//localAddr is updated in receive
	go Receive()
	go Select_receive()
	go Select_send_slave()
	go Write_to_network()

	go func() {
		fmt.Println("to channel")
		ipOrd := IpOrderList{}
		ExSlaveChans.ToCommNetworkInitChan <- ipOrd
	}()
	time.Sleep(time.Second)
	ipVarList := make([]*UDPAddr, N_ELEV)
	//ipVarList[0] = localAddr

	var ipOrd IpOrderList
	tick := time.After(5000 * time.Millisecond)
	for {
		fmt.Println("forloop")

		select {
		case ipOrd = <-ExCommChans.ToSlaveNetworkInitRespChan:
			fmt.Println("case 2")
			isInList := false
			for i := 0; i < N_ELEV; i++ {
				fmt.Println("test1")
				if ipVarList[i] == ipOrd.Ip {
					isInList = true
				}
				fmt.Println("test1")
				if isInList {
					ipVarList[i] = ipOrd.Ip
				}

			}

		case <-tick:
			fmt.Println("case 3")
			break
		}
		fmt.Println("ipOrderList", ipOrd)
		//break /////////////////////???????????? it dont trigger the first break to break forloop
	}

	fmt.Println("velger master")
	//VELGER MASTER
	fmt.Println("ipVarList", ipVarList)
	tempBestMaster := ipVarList[0]
	for i := 1; i < len(ipVarList); i++ {
		if IpSum(tempBestMaster) > IpSum(ipVarList[i]) {
			tempBestMaster = ipVarList[i]
		}
	}
	//fmt.Println("localAddr", localAddr)
	fmt.Println("tempBestMaster", tempBestMaster)
	if localAddr == tempBestMaster {
		fmt.Println("this is master")
		go Active_master()
		go Check_slaves()

	} else {
		go Check_master()
		fmt.Println("this is slave")

	}
	//go Elevator_manager()
	fmt.Println("end of activeslave")
	blockingChan := make(chan bool)
	<-blockingChan
	//conn.Close()
	fmt.Println("end of activeslave, after block")
}

func Active_master() {
	//address, _ := ResolveUDPAddr("udp", "129.241.187.255"+PORT)
	//conn, _ := DialUDP("udp", nil, address)

	fmt.Println("Master")
	go Write_to_network()
	go Select_receive()

}

func IpSum(addr *UDPAddr) (sum byte) {
	bArr, _ := Marshal(addr)
	for _, value := range bArr {
		sum += value
	}
	return sum
}

func Master_top_logic(m Master) {

	for {
		select {
		case ipOrder := <-InLogicChans.ToTopLogicOrderChan:
			//HANDLE INTERNAL BUTTONS -> BUTTON TYPE
			if ipOrder.Ord.TurnOn { // If order executed, just update the internal arrays and the updater will notify when updated. It will use IP smartly
				m.Set_external_list_order(ipOrder.Ip, ipOrder.Ord.Floor, ipOrder.Ord.ButtonType, ipOrder)
				m.Set_external_list_order(nil, ipOrder.Ord.Floor, ipOrder.Ord.ButtonType, ipOrder)
				//m.ExternalList[nil][ipOrder.Order.Floor][ipOrder.Order.ButtonType] = ipOrder.Order.TurnOn

				InLogicChans.ExternalListIsUpdated <- true //THIS IS THE MAP WE MUST CHANGE. WE SHOULD DO THIS TOGETHER SINCE A LOT OF FUNCTIONALITY USES IT

				//ToStateMachineArrayChan <- LocalMaster.AllArrays[LocalMaster.myIP]	WE NEED TO LOOK AT THIS FFS!!!!!!

			} else { //else its a button pressed and we need the optimization module decide who gets it
				ExOptimalChans.OptimizationTriggerChan <- ipOrder

			}
		case ipState := <-InLogicChans.ToStateUpdater:
			m.Statelist[ipState.Ip] = ipState
		case ipOrder := <-ExOptimalChans.OptimizationReturnChan:
			m.Set_external_list_order(ipOrder.Ip, ipOrder.Ord.Floor, ipOrder.Ord.ButtonType, ipOrder)
			m.Set_external_list_order(nil, ipOrder.Ord.Floor, ipOrder.Ord.ButtonType, ipOrder)
			InLogicChans.ExternalListIsUpdated <- true //We now have two channels writing to one channel, but the goroutine should empty the buffer quite nicely

		}
	}
}

func Master_updated_state_incoming() {
	for {
		updatedState := <-InLogicChans.ToMasterUpdateStateChan
		InLogicChans.ToStateUpdater <- updatedState
		ExSlaveChans.ToCommUpdatedStateChan <- updatedState

	}
}

func Master_updated_externalList_outgoing(m Master) {
	localSlaveMap := IpOrderList{}
	//timerMap := make(map[*UDPAddr]time.Time)               //timers for each IP
	startCountdownChan := make(chan bool)
	countdownFinishedChan := make(chan bool)
	allSlavesAnsweredChan := make(chan bool)
	startReceivingChan := make(chan map[*UDPAddr]*[N_FLOORS][2]bool)
	countingSlaveMap := make(map[*UDPAddr]*[N_FLOORS][2]bool)
	var hasSentAlert bool
	timer := make(<-chan time.Time)

	//Get the total shit if it has been updated, and the slaves need to know
	go func() {
		for {
			select {
			case <-InLogicChans.ExternalListIsUpdated: //COMING FROM LOCAL STATEMACHINE -  THIS NEEDS TO BE MADE - FEEL FREE TO CHANGE NAME
				localSlaveMap.ExternalList = m.Get_external_list()
				ExMasterChans.ToCommOrderListChan <- localSlaveMap //TO THE COMMUNICATION MODULE --  THIS NEEDS TO BE MADE - FEEL FREE TO CHANGE NAME
				startReceivingChan <- localSlaveMap.ExternalList
				startCountdownChan <- true
			case <-countdownFinishedChan:
				ExMasterChans.ToCommOrderListChan <- localSlaveMap //TO THE COMMUNICATION MODULE --  THIS NEEDS TO BE MADE - FEEL FREE TO CHANGE NAME
			}

		}
	}()
	//Timer goroutine, receives answer and removes from
	//Also sends again if no answer
	go func() {
		for {
			select {
			case <-startCountdownChan:
				timer = time.After(500 * time.Millisecond) // If all answers, find a way to stop timer
			case <-timer:
				countdownFinishedChan <- true
			case <-allSlavesAnsweredChan:
				timer = nil
			}
		}
	}()
	go func() {
		for {
			select {
			case <-startReceivingChan:
				countingSlaveMap = m.Get_external_list()
				hasSentAlert = false
			case receivedOrder := <-ExCommChans.ToMasterOrderListReceivedChan: //(type: )  FROM THE COMMUNICATION MODULE --  THIS NEEDS TO BE MADE - FEEL FREE TO CHANGE NAME
				allListsMatch := true
				for key, _ := range receivedOrder.ExternalList {
					if receivedOrder.ExternalList[key] != m.Get_external_list()[key] {
						allListsMatch = false

					}
				}
				if allListsMatch {
					delete(countingSlaveMap, receivedOrder.Ip)
				}

				if len(countingSlaveMap) == 1 && !hasSentAlert { //Only LightArray Remains
					allSlavesAnsweredChan <- true
					hasSentAlert = true
				}
			}
			time.Sleep(25 * time.Millisecond)
		}

	}()
}

//Either copy-paste this or send it to optimization-module in the code where it is handled. May just have a goroutine in this function as well
func Master_incoming_order_executed() { //RENAME THIS MOTHERFUCKER TO TAKE CARE OF ALL THE SHITS//Generalize this for all orders, either ordered or executed.
	countdownChan := make(chan IpOrderMessage)
	timerMap := make(map[IpOrderMessage]time.Time)
	var SyncOrderMap struct {
		sync.RWMutex
		m map[Order]*UDPAddr
	}

	go func() {
		for {
			//Updates the queue, if the same kind of messages are sent simultaneously
			orderExe := <-ExCommChans.ToMasterOrderExecutedChan //This is on IP-message-form

			InLogicChans.ToTopLogicOrderChan <- orderExe /*The code that receives isnt made yet. Should handle optimization module there.*/

			SyncOrderMap.RLock()
			inQueue := SyncOrderMap.m[orderExe.Ord] //It will be nil if its not in the map
			SyncOrderMap.RUnlock()
			if inQueue == nil { //If its not in queue we should
				InLogicChans.ToTopLogicOrderChan <- orderExe
				SyncOrderMap.Lock()
				SyncOrderMap.m[orderExe.Ord] = orderExe.Ip
				SyncOrderMap.Unlock()
			}
			countdownChan <- orderExe
			ExMasterChans.ToCommOrderExecutedConfirmedChan <- orderExe
		}
	}()
	//Timer function, that concurrently deletes orders/renews the timer.
	go func() {
		for {
			select {
			case orderReceived := <-countdownChan:
				timerMap[orderReceived] = time.Now()
			default:
				for ipOrder, timestamp := range timerMap {
					if time.Since(timestamp) > time.Millisecond*500 { //RLock before and after this if?
						delete(timerMap, ipOrder)
						SyncOrderMap.Lock()
						delete(SyncOrderMap.m, ipOrder.Ord)
						SyncOrderMap.Unlock()
					}
				}
				time.Sleep(25 * time.Millisecond) // Change to optimize
			}
		}
	}()
}

/*
	Knows where this came from.
	Update the array that the optimization algorithm uses, but dont send it to the algorithm.
	I guess you need to send this to all the slaves and stop when they have confirmed that they have received it.
	They need to know this in case the master goes down. This will also trigger the lights of all elevators.
	Just keep sending this to all the slaves and mark the slaves as they confirm that they have received it.
	When all slaves are marked this will stop sending.

	If multiple different orders are incoming, how do we spawn to threads? Does goroutines work this way?
*/

//Slaves functions
/*
func Slave_top_Logic() {
	for {
		select {
		case order := <-ExternalButtonPressed:
			ToSlaveOrderOutChan <- order
		case allArrays := <-ToTopLogicChan:
			fmt.Println("UPDATE LOCAL SLAVE HERE, FIND OUT HOW")

		} //SHOULD WE JUST SEND STATE AND BUTTON PRESSED DIRECTLY???? DO WE NEED THIS GUY AT ALL?  WE NEED TO FIND THE FUCK OUT

	}

}*/

func Slave_Order_Outgoing() {
	//	countdownChan := make(chan IpOrderMessage)
	var SyncOrderMap struct {
		sync.RWMutex
		m map[IpOrderMessage]time.Time
	}

	go func() {
		for {
			//Updates the queue, if the same kind of messages are sent simultaneously
			orderOut := <-ExStateMChans.ButtonPressedChan
			IpOrderOut := IpOrderMessage{nil, orderOut}
			SyncOrderMap.Lock()
			SyncOrderMap.m[IpOrderOut] = time.Now()
			SyncOrderMap.Unlock()
			ExSlaveChans.ToCommExternalButtonPushedChan <- orderOut
		}
	}()
	//Timer function, that concurrently renews the timer and sends another message if the old one has timed out.
	go func() {
		for {
			for orderOut, timestamp := range SyncOrderMap.m {
				if time.Since(timestamp) > time.Millisecond*500 { //temp
					SyncOrderMap.Lock()
					SyncOrderMap.m[orderOut] = time.Now()
					SyncOrderMap.Unlock()
					ExSlaveChans.ToCommExternalButtonPushedChan <- orderOut.Ord

				}
				time.Sleep(25 * time.Millisecond) // Change to optimize
			}
		}
	}()
	//Waits for a response and removes the element from map if it has been confirmed by master.
	go func() {
		for {
			orderOut := <-ExCommChans.ToSlaveButtonPressedConfirmedChan
			SyncOrderMap.Lock()
			delete(SyncOrderMap.m, orderOut)
			SyncOrderMap.Unlock()
		}
	}()

}

//No queueing necessary
func Slave_order_arrays_incoming(s Slave) {
	//var NewExternalList map[*UDPAddr]*[N_FLOORS][2]bool //Think sending the master could be good, but master isnt a good name
	go func() {
		for {
			NewExternalList := <-ExCommChans.ToSlaveOrderListChan
			s.Overwrite_external_list(NewExternalList.ExternalList)
			ExStateMChans.SingleExternalListChan <- *NewExternalList.ExternalList[s.Get_ip()]
			ExStateMChans.LightChan <- *NewExternalList.ExternalList[nil]
			//Here we need to save all the information about the other slaves, and send our own to the statemachine
			ExSlaveChans.ToCommOrderListReceivedChan <- NewExternalList
		}
	}()
}

//Sends state if timer expires or state changes.
func Slave_state_updated() {
	sendAgainTimer := make(<-chan time.Time)
	var localCurrentState IpState
	for {
		select {
		case localCurrentState = <-ExStateMChans.CurrentStateChan:
			ExSlaveChans.ToCommUpdatedStateChan <- localCurrentState
			sendAgainTimer = time.After(50 * time.Millisecond)
		case currentStateReceived := <-ExCommChans.ToSlaveUpdateStateReceivedChan:
			if currentStateReceived == localCurrentState {
				sendAgainTimer = nil //Not sure if this is legal, will this send to channel if its set to nil??
			}
		case <-sendAgainTimer: //This will be sent when time runs out, I think.
			ExSlaveChans.ToCommUpdatedStateChan <- localCurrentState
			sendAgainTimer = time.After(50 * time.Millisecond)
		}
	}
}

//checks if some of the slaves sends Im Slave signal
func Check_slaves() {
	timer := make(map[*UDPAddr]time.Time)
	for {
		select {
		case slaveIpOrder := <-ExCommChans.ToMasterImSlaveChan:
			timer[slaveIpOrder.Ip] = time.Now()
		case <-ExCommChans.ToSlaveNetworkInitChan:
			fmt.Println("new slave on network - reset")
			Restart_system()

		}

		for _, value := range timer {
			if time.Since(value) > MAXWAIT_IM_HERE {
				//RestartSystemChan <- ExternalList
			}
		}
	}
}

//checks if master still is sending Im Master signal
func Check_master() {
	var timer time.Time
	for {
		select {
		case <-ExCommChans.ToSlaveImMasterChan:
			if time.Since(timer) > MAXWAIT_IM_HERE {
				//RestartSystemChan <- ExternalList
			}
		}
	}
}
func Restart_system() {
	for {
		<-RestartSystemChan
		ExSlaveChans.ToCommRestartSystemChan <- true
		defer os.Exit(666)
		cmd := exec.Command("mate-terminal", "-x", EXE_FILE)
		cmd.Run()
		//needs to use external list

	}
}
