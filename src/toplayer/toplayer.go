package toplayer

import (
	. "chansnstructs"
	. "network"
	//. "sync"
	. "encoding/json"
	"fmt"
	. "net"
	. "optimal"
	"os"
	"os/exec"
	. "statemachine"
	"strings"
	"sync"
	"time"
)

var InLogicChans InternalLogicChannels
var InteruptChan chan os.Signal
var RestartSystemChan chan [][N_FLOORS][2]bool

var standardArray [N_FLOORS][2]bool

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

func Activate_slave() {
	//externalListRFF := Read_from_file()
	internal_logic_channels_init()
	//fmt.Println("start slave")
	var localAddr *UDPAddr

	fmt.Println(localAddr)
	fmt.Println("&standardarray", standardArray)
	//standardArray = {{false, false}, {false ,false } ,{false ,false }, {false ,false }}
	//fmt.Println("connection init")
	go Receive()
	go Select_receive()
	go Select_send_slave()
	go Write_to_network()
	go Select_send_master()

	// This will not be blocking becauce of Select_send will take it
	netInitTrigger := IpOrderList{}
	ExSlaveChans.ToCommNetworkInitChan <- netInitTrigger
	ipVarList := make([]*UDPAddr, N_ELEV)
	var ipOrd IpOrderList
	tick := time.After(2000 * time.Millisecond)
	check := true

	ipAddressiter := 0
	for check {

		//fmt.Println("forloop")
		select {
		case ipOrd = <-ExCommChans.ToSlaveNetworkInitRespChan:
			//fmt.Println("case 2")
			isInList := false
			for i := 0; i < N_ELEV; i++ {
				if ipVarList[i] == ipOrd.Ip {
					isInList = true
				}
				//	fmt.Println("test1")
			}
			if !isInList {
				ipVarList[ipAddressiter] = ipOrd.Ip
				ipAddressiter++
			}

		case <-tick:

			//fmt.Println("case 3")
			check = false
			break
		}
		//fmt.Println("ipOrderList", ipOrd)

	}
	//VELGER MASTER
	//fmt.Println("ipVarList", ipVarList)

	tempBestMaster := ipVarList[0]
	for i := 1; i < len(ipVarList); i++ {
		if IpSum(tempBestMaster) > IpSum(ipVarList[i]) {
			tempBestMaster = ipVarList[i]
		}
	}
	fmt.Println("tempbestmaster: ", tempBestMaster)
	fmt.Println("localaddr : ", localAddr)
	localAddr = tempBestMaster // WE NEED TO FIND THE LOCAL IP SOMEHOW
	//fmt.Println("localAddr", localAddr)
	fmt.Println("EQUALITYYY", localAddr == tempBestMaster)
	//fmt.Println("tempBestMaster", tempBestMaster)
	if localAddr == tempBestMaster {
		m := Master{}

		//m.ExternalList = externalListRFF

		m.SlaveIp = ipVarList
		m.ExternalList = make([][N_FLOORS][2]bool, N_ELEV+1)
		m.Statelist = make([]IpState, N_ELEV)
		fmt.Println("wqerqwrqwrqwrqwerqwerqwerqwer", m.Statelist)

		go Activate_master(&m) //&
		fmt.Println("master= ", m)
		go Check_slaves(&m) //&
		go Optimalization_init(&m)

	}
	//no else becauce the master is also a "slave"
	s := Slave{}

	go Check_master(&s)
	//go Slave_top_Logic(s)
	go Slave_Order_Outgoing()
	go Slave_order_arrays_incoming(&s)
	go Slave_state_updated()

	//fmt.Println("this is slave")

	go Elevator_manager()

	//fmt.Println("end of activeslave")

	blockingChan := make(chan bool)
	<-blockingChan
	//conn.Close()
	fmt.Println("end of activeslave, after block")
}

func Activate_master(m *Master) { //*
	fmt.Println("Starting Master")
	//address, _ := ResolveUDPAddr("udp", "129.241.187.255"+PORT)
	//conn, _ := DialUDP("udp", nil, address)

	go Master_updated_state_incoming()
	go Write_to_network()
	go Select_receive()
	go Master_top_logic(m) //&
	go Master_incoming_order_executed()
	go Master_updated_externalList_outgoing(m) //&
}

func IpSum(addr *UDPAddr) (sum byte) {
	bArr, _ := Marshal(addr)
	for _, value := range bArr {
		sum += value
	}
	return sum
}

func Master_updated_state_incoming() {
	for {
		updatedState := <-ExCommChans.ToMasterUpdateStateChan
		fmt.Println("			State update received:  			", updatedState)

		InLogicChans.ToStateUpdater <- updatedState

		ExMasterChans.ToCommUpdateStateReceivedChan <- updatedState
		fmt.Println("finished loop of updated state")
	}
}

func Master_top_logic(m *Master) {
	for {
		select {
		case ipOrder := <-InLogicChans.ToTopLogicOrderChan:
			fmt.Println("we have received from secure sending grid")
			fmt.Println("ipOrder in master top logic: ", ipOrder)
			//m.ExternalList[ipOrder.Ip] = standardArray

			if !ipOrder.Ord.TurnOn { // If order executed, just update the internal arrays and the updater will notify when updated. It will use IP smartly
				for i := 0; i < N_ELEV; i++ {
					if m.SlaveIp[i] == ipOrder.Ip {
						m.ExternalList[i][ipOrder.Ord.Floor][ipOrder.Ord.ButtonType] = ipOrder.Ord.TurnOn
						m.ExternalList[N_ELEV][ipOrder.Ord.Floor][ipOrder.Ord.ButtonType] = ipOrder.Ord.TurnOn
						if ipOrder.Ord.TurnOn {
							fmt.Println("MADDAFAKKINGS ERROR MASTER TOP LOGIC")
						}
					}

				}
				InLogicChans.ExternalListIsUpdated <- true //THIS IS THE MAP WE MUST CHANGE. WE SHOULD DO THIS TOGETHER SINCE A LOT OF FUNCTIONALITY USES IT
				//ToStateMachineArrayChan <- LocalMaster.AllArrays[LocalMaster.myIP]	WE NEED TO LOOK AT THIS FFS!!!!!!

			} else { //else its a button pressed and we need the optimization module decide who gets it
				fmt.Println("               This is else in toplogic")
				ExOptimalChans.OptimizationTriggerChan <- ipOrder

			}
		case ipState := <-InLogicChans.ToStateUpdater:
			for i := 0; i < N_ELEV; i++ {
				if m.SlaveIp[i] == ipState.Ip {
					m.Statelist[i] = ipState
				}
			}
		case ipOrder := <-ExOptimalChans.OptimizationReturnChan:
			fmt.Println("optimization is finished")
			for i := 0; i < N_ELEV; i++ {
				if m.SlaveIp[i] == ipOrder.Ip {
					m.ExternalList[i][ipOrder.Ord.Floor][ipOrder.Ord.ButtonType] = ipOrder.Ord.TurnOn
					m.ExternalList[N_ELEV][ipOrder.Ord.Floor][ipOrder.Ord.ButtonType] = ipOrder.Ord.TurnOn
					if !ipOrder.Ord.TurnOn {
						fmt.Println("MADDAFAKKINGS ERROR MASTER TOP LOGIC RETURN OPTIMIZATION")
					}
				}

			}
			fmt.Println("ex list updated")
			InLogicChans.ExternalListIsUpdated <- true
			fmt.Println("ex list updated end")
			//sets order

			//sets lights

		}
	}
}

//Either copy-paste this or send it to optimization-module in the code where it is handled. May just have a goroutine in this function as well
func Master_incoming_order_executed() { //RENAME THIS MOTHERFUCKER TO TAKE CARE OF ALL THE SHITS//Generalize this for all orders, either ordered or executed.
	countdownChan := make(chan IpOrderMessage)
	timerMap := make(map[IpOrderMessage]time.Time)
	var orderExe IpOrderMessage

	var SyncOrderMap = struct {
		sync.RWMutex
		m map[Order]*UDPAddr
	}{m: make(map[Order]*UDPAddr)}

	go func() {
		for {
			orderReceived := <-countdownChan
			fmt.Println("countfdowntimer")
			timerMap[orderReceived] = time.Now()
		}
	}()

	for {

		select {
		//Updates the queue, if the same kind of messages are sent simultaneously
		case orderExe = <-ExCommChans.ToMasterOrderExecutedChan: //This is on IP-message-form
			fmt.Println("														received from masterORderExecuteChan")
			fmt.Println(SyncOrderMap.m)
			InLogicChans.ToTopLogicOrderChan <- orderExe /*The code that receives isnt made yet. Should handle optimization module there.*/
			fmt.Println("Sent to toplogicOrderchan")

			SyncOrderMap.RLock()
			inQueue := SyncOrderMap.m[orderExe.Ord] //It will be nil if its not in the map
			fmt.Println("inqueueueueueu:   ", inQueue)
			SyncOrderMap.RUnlock()
			if inQueue == nil { //If its not in queue we should
				//InLogicChans.ToTopLogicOrderChan <- orderExe
				SyncOrderMap.Lock()
				SyncOrderMap.m[orderExe.Ord] = orderExe.Ip
				fmt.Println(SyncOrderMap.m)
				SyncOrderMap.Unlock()
			}
			fmt.Println("masterincomming possible block toplayer")
			countdownChan <- orderExe
			fmt.Println("IS countdownchan the blocker?????????????")
			ExMasterChans.ToCommOrderExecutedConfirmedChan <- orderExe
			fmt.Println("Sent to ")

		default:
			fmt.Println("default")
			for ipOrder, timestamp := range timerMap {
				if time.Since(timestamp) > time.Millisecond*500 { //RLock before and after this if?
					delete(timerMap, ipOrder)
					SyncOrderMap.Lock()
					delete(SyncOrderMap.m, ipOrder.Ord)
					SyncOrderMap.Unlock()
					fmt.Println("...")
				}
				fmt.Println("IM IN THIS FOR FOREVER MADDAFAKKAS!!!111!!")
			}
			time.Sleep(2500 * time.Millisecond) // Change to optimize
		}
	}
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
func Master_updated_externalList_outgoing(m *Master) {
	localSlaveSlice := IpOrderList{}
	//timerMap := make(map[*UDPAddr]time.Time)               //timers for each IP
	startCountdownChan := make(chan bool)
	countdownFinishedChan := make(chan bool)
	allSlavesAnsweredChan := make(chan bool)
	startReceivingChan := make(chan [][N_FLOORS][2]bool, N_ELEV+1)
	countingSlaveSlice := make(chan [][N_FLOORS][2]bool, N_ELEV+1)
	var hasSentAlert bool
	timer := make(<-chan time.Time)

	//Get the total shit if it has been updated, and the slaves need to know
	go func() {
		for {
			select {
			case <-InLogicChans.ExternalListIsUpdated: //COMING FROM LOCAL STATEMACHINE -  THIS NEEDS TO BE MADE - FEEL FREE TO CHANGE NAME
				//localSlaveMap.ExternalList = m.Get_external_list()
				ExMasterChans.ToCommOrderListChan <- localSlaveSlice //TO THE COMMUNICATION MODULE --  THIS NEEDS TO BE MADE - FEEL FREE TO CHANGE NAME
				startReceivingChan <- localSlaveSlice.ExternalList
				startCountdownChan <- true
			case <-countdownFinishedChan:
				ExMasterChans.ToCommOrderListChan <- localSlaveSlice //TO THE COMMUNICATION MODULE --  THIS NEEDS TO BE MADE - FEEL FREE TO CHANGE NAME
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
	var allListsMatch bool
	go func() {
		for {
			select {
			case <-startReceivingChan:
				hasSentAlert = false
			case ipOrdList := <-ExCommChans.ToMasterOrderListReceivedChan:
				localSlaveSlice := m.Get_external_list()
				receivedOrder := ipOrdList.ExternalList
				allListsMatch = true
				for i := 0; i < N_ELEV; i++ {
					for j := 0; i < N_FLOORS; j++ {
						for k := 0; k < 2; k++ {
							if localSlaveSlice[i][j][k] && receivedOrder[i][j][k] {
								localSlaveSlice[i][j][k] = false
								hasSentAlert = true ////??????????+
							} else {
								allListsMatch = false
							}
						}
					}
				}
				/*****
				for key, _ := range receivedOrder.ExternalList {
					if receivedOrder.ExternalList[key] != m.Get_external_list()[key] {
						allListsMatch = false

					}
				}
				if allListsMatch {
					countingSlaveSlice[][][]
					delete(countingSlaveSlice, receivedOrder.Ip)
				}
				*/

				if len(countingSlaveSlice) == 1 && !hasSentAlert { //Only LightArray Remains
					allSlavesAnsweredChan <- true
					hasSentAlert = true
				}
			}
			time.Sleep(25 * time.Millisecond)
		}
	}()
}

func Slave_order_arrays_incoming(s *Slave) {
	//var NewExternalList map[*UDPAddr]*[N_FLOORS][2]bool //Think sending the master could be good, but master isnt a good name
	go func() {
		for {
			NewExternalList := <-ExCommChans.ToSlaveOrderListChan
			//s.Overwrite_external_list(NewExternalList.ExternalList)
			//ExStateMChans.SingleExternalListChan <- *NewExternalList.ExternalList[s.Get_ip()]

			//This access the last array containing ligths
			ExStateMChans.LightChan <- NewExternalList.ExternalList[N_FLOORS]
			//Here we need to save all the information about the other slaves, and send our own to the statemachine
			ExSlaveChans.ToCommOrderListReceivedChan <- NewExternalList
		}
	}()
}
func Slave_Order_Outgoing() {
	//fmt.Println("Slave_Order_Outgoing")
	//	countdownChan := make(chan IpOrderMessage)
	var SyncOrderMap = struct {
		sync.RWMutex
		m map[IpOrderMessage]time.Time
	}{m: make(map[IpOrderMessage]time.Time)}
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
			orderOut := <-ExCommChans.ToSlaveOrderExecutedConfirmedChan
			fmt.Println("				ORDER BEFORE DELETION", orderOut)
			fmt.Println("Before deletion: ", SyncOrderMap.m)
			SyncOrderMap.Lock()
			delete(SyncOrderMap.m, orderOut)
			SyncOrderMap.Unlock()
			fmt.Println("After deletion: ", SyncOrderMap.m)
		}
	}()

}

//Sends state if timer expires or state changes.
func Slave_state_updated() {
	//sendAgainTimer := make(<-chan time.Time)
	var localCurrentState State
	for {
		select {
		case localCurrentState = <-ExStateMChans.CurrentStateChan:
			fmt.Println("slave state updated inside case!!!!")
			ExSlaveChans.ToCommUpdatedStateChan <- localCurrentState
			fmt.Println("after tocommupdatedstatechan")
			//sendAgainTimer = time.After(50 * time.Millisecond)
		case currentStateReceived := <-ExCommChans.ToSlaveUpdateStateReceivedChan:
			if currentStateReceived.Sta == localCurrentState {
				//sendAgainTimer = nil //Not sure if this is legal, will this send to channel if its set to nil??
				fmt.Println("slave state updated")
			}
			/*case <-sendAgainTimer: //This will be sent when time runs out, I think.
			fmt.Println("send again timer")

			ExSlaveChans.ToCommUpdatedStateChan <- localCurrentState
			//	fmt.Println("after send again timer")
			sendAgainTimer = time.After(500 * time.Millisecond)
			*/
		}
	}
}

//checks if some of the slaves sends Im Slave signal
func Check_slaves(m *Master) {
	ipList := m.SlaveIp
	var timer = map[*UDPAddr]time.Time{
		ipList[0]: time.Now(),
		ipList[1]: time.Now(),
		ipList[2]: time.Now(),
	}
	for {
		select {
		case slaveIpOrder := <-ExCommChans.ToMasterImSlaveChan:
			timer[slaveIpOrder.Ip] = time.Now()
		case <-ExCommChans.ToSlaveNetworkInitChan:
			//fmt.Println("new slave on network - reset")
			Restart_system(m.Get_external_list())

		}

		for _, value := range timer {
			if time.Since(value) > MAXWAIT_IM_HERE {
				RestartSystemChan <- m.ExternalList
			}
		}
	}
}

//checks if master still is sending Im Master signal
func Check_master(s *Slave) {
	var timer time.Time
	for {
		select {
		case <-ExCommChans.ToSlaveImMasterChan:
			if time.Since(timer) > MAXWAIT_IM_HERE {
				RestartSystemChan <- s.ExternalList

			}
		}
	}
}
func Restart_system(externalList [][N_FLOORS][2]bool) {
	for {
		<-RestartSystemChan
		ExSlaveChans.ToCommRestartSystemChan <- true
		Write_to_file(externalList)
		defer os.Exit(666)
		cmd := exec.Command("mate-terminal", "-x", EXE_FILE)
		cmd.Run()
		//needs to use external list

	}
}
func Write_to_file(externalList [][N_FLOORS][2]bool) {
	bArr, err := Marshal(externalList)
	if err != nil {
		fmt.Println(err.Error(), " Error marshaling externalList")
	}

	os.Remove("exList.txt")
	file, err := os.Create("exList.txt")
	if err != nil {
		fmt.Println(err.Error(), " Error creating file")
	}

	_, err = file.Write(bArr)
	if err != nil {
		fmt.Println(err.Error(), " Error writing file")
	}

}
func Read_from_file() [][N_FLOORS][2]bool {
	file, err := os.Open("exList.txt")
	if err != nil {
		fmt.Println(err.Error(), " Error opening file")
	}
	bArr := make([]byte, 256)
	_, err = file.Read(bArr)
	if err != nil {
		fmt.Println(err.Error(), " Error opening file")
	}
	externalList := make([][N_FLOORS][2]bool, N_ELEV)
	err = Unmarshal(bArr, &externalList)
	if err != nil {
		fmt.Println(err.Error(), " Error unmarhaling file")
	}
	return externalList
}

func GetMyIP() string {
	con1, _ := Dial("udp", "google.com:http")
	myIP := con1.LocalAddr().String()
	str := strings.Split(myIP, ":")
	myIP = str[0]
	con1.Close()
	return myIP
}
