package toplayer

import (
	. "chansnstructs"
	. "network"
	. "encoding/json"
	"fmt"
	. "net"
	. "optimal"
	"os"
	"os/exec"
	. "statemachine"
	"os/signal"
	"sync"
	"time"
	"syscall"

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
	externalList, err := Read_from_file() 
	if err != nil {
		fmt.Println(err.Error(), "failed to read from file")
	}
	//externalListRFF := Read_from_file()
	internal_logic_channels_init()
	//fmt.Println("start slave")

	templocalAddr := Get_my_ip()
	localAddr, _ := ResolveUDPAddr("udp", templocalAddr)
	fmt.Println("local address ",localAddr)
	
	go Receive()
	go Select_receive()
	go Select_send_slave()
	go Write_to_network()
	go Select_send_master()
	Interuption_killer(externalList)

	netInitTrigger := IpOrderList{}
	ExSlaveChans.ToCommNetworkInitChan <- netInitTrigger
	ipVarList := make([]*UDPAddr, N_ELEV)
	var ipOrd IpOrderList
	tick1 := time.After(3000 * time.Millisecond)
	check1 := true
	tick2 := time.After(3000* time.Millisecond)
	check2 := true
	go func(){
		for check2{
			select{
			case<- tick2:
				check2 = false

			default:
				ExSlaveChans.ToCommNetworkInitChan <- netInitTrigger
				time.Sleep(time.Millisecond*25)
			}
		}
	}()

	ipAddressiter := 0
	for check1 {
	
		select {
		case ipOrd = <-ExCommChans.ToSlaveNetworkInitRespChan:
			isInList := false
			for i := 0; i < N_ELEV; i++ {
				if ipVarList[i] != nil {
					if ipVarList[i].String()[:15] == ipOrd.Ip.String()[:15] {
						isInList = true
					}
				}
			}
			if !isInList {
						ipVarList[ipAddressiter] = ipOrd.Ip
						ipAddressiter++
					}

		case <-tick1:
			check1 = false
			break
		}
	}

	nilCounter := 0
	var mongofaen[N_ELEV] int
	tempBestMaster := ipVarList[0]
	for i := 0; i < len(ipVarList); i++ {
		if ipVarList[i] == nil{
			nilCounter++
			mongofaen[i] = 1
		}
		if ipVarList[i] != nil && tempBestMaster.String()[:15] > ipVarList[i].String()[:15] {
			tempBestMaster = ipVarList[i]
		}
		//fmt.Println("		i = ",i)
	}
	var tempSortArray []*UDPAddr = make([]*UDPAddr,N_ELEV)
	for i:= 0; i < (len(ipVarList) - nilCounter); i++{
		tempbestIP := ipVarList[0]
		minval := 0		
		for j:= 0; j < len(ipVarList); j++ {
			fmt.Println(ipVarList[j])
			if mongofaen[j] != 1{
				if tempbestIP.String()[:15] > ipVarList[j].String()[:15]{
					tempbestIP = ipVarList[j]
					minval = j
				}
			}
		}
		tempSortArray[i] = ipVarList[i]
		tempSortArray[i] = ipVarList[minval]

		mongofaen[minval] = 1
		}
	ipVarList = &tempSortArray

	if templocalAddr == tempBestMaster.String()[:15] {
		m := Master{}
		fmt.Println("adds master!!!")

		//m.ExternalList = externalListRFF
		//tick3 := time.After(40* time.Millisecond)
		//check3 := true
		/*go func(){
			for check3{
				select{
				case<- tick3:
					check3 = false

				default:
					ExMasterChans.ToCommSendIpListChan <- ipVarList

				}
			}
		}()*/

		m.SlaveIp = ipVarList
		m.ExternalList = make([][N_FLOORS][2]bool, N_ELEV+1)
		m.Statelist = make([]IpState, N_ELEV)
		fmt.Println("wqerqwrqwrqwrqwerqwerqwerqwer", m.Statelist)

		go Activate_master(&m) //&
		//fmt.Println("master= ", m)
		//go Check_slaves(&m) //&
		go Optimalization_init(&m)

	}
	s := Slave{}
	
	for i := 0; i < len(ipVarList) - nilCounter; i++ {
		if ipVarList[i].String()[:15] == templocalAddr {
			s.Nr = i
		}
	}
	go Check_master(&s)
	go Slave_Order_Outgoing()
	go Slave_order_arrays_incoming(&s)
	go Slave_state_updated()


	go Elevator_manager()
	fmt.Println("go elevator manager")
	blockingChan := make(chan bool)
	<-blockingChan
	//conn.Close()
	fmt.Println("end of activeslave, after block")
}

func Activate_master(m *Master) { //*
	//fmt.Println("Starting Master")
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
		fmt.Println("start master updated state incoming")
		updatedState := <-ExCommChans.ToMasterUpdateStateChan
		//fmt.Println("State update received  YOLOMAX: ", updatedState)

		InLogicChans.ToStateUpdater <- updatedState

		ExMasterChans.ToCommUpdateStateReceivedChan <- updatedState
		//fmt.Println("finished loop of updated state")
	}
}

func Master_top_logic(m *Master) {
	fmt.Println("master top logic")
	for {
		select {
		case ipOrder := <-InLogicChans.ToTopLogicOrderChan:
			if !ipOrder.Ord.TurnOn { // If order executed, just update the internal arrays and the updater will notify when updated. It will use IP smartly
				for i := 0; i < N_ELEV; i++ {
					if m.SlaveIp[i] != nil && m.SlaveIp[i].String()[:15] == ipOrder.Ip.String()[:15] {
						m.ExternalList[i][ipOrder.Ord.Floor][ipOrder.Ord.ButtonType] = ipOrder.Ord.TurnOn
						m.ExternalList[N_ELEV][ipOrder.Ord.Floor][ipOrder.Ord.ButtonType] = ipOrder.Ord.TurnOn
						if ipOrder.Ord.TurnOn {
						//	fmt.Println("ERROR MASTER TOP LOGIC")
						}
					}

				}
				InLogicChans.ExternalListIsUpdated <- true 
				//ToStateMachineArrayChan <- LocalMaster.AllArrays[LocalMaster.myIP]	

			} else { //else its a button pressed and we need the optimization module decide who gets it
			//	fmt.Println("               This is else in toplogic")
				ExOptimalChans.OptimizationTriggerChan <- ipOrder

			}
		case ipState := <-InLogicChans.ToStateUpdater:
			//fmt.Println(" START UPDATE STATE")
			for i := 0; i < N_ELEV; i++ {
				if m.SlaveIp[i] == ipState.Ip {
					m.Statelist[i] = ipState
				}
			}
			//fmt.Println(" END UPDATE STATE")
		case ipOrder := <-ExOptimalChans.OptimizationReturnChan:
			//fmt.Println("optimization is finished")
			for i := 0; i < N_ELEV; i++ {
				if m.SlaveIp[i] == ipOrder.Ip {
					m.ExternalList[i][ipOrder.Ord.Floor][ipOrder.Ord.ButtonType] = ipOrder.Ord.TurnOn
					m.ExternalList[N_ELEV][ipOrder.Ord.Floor][ipOrder.Ord.ButtonType] = ipOrder.Ord.TurnOn
					if !ipOrder.Ord.TurnOn {
						//fmt.Println("MADDAFAKKINGS ERROR MASTER TOP LOGIC RETURN OPTIMIZATION")
					}
				}

			}
			InLogicChans.ExternalListIsUpdated <- true
		}
	}
}


func Master_incoming_order_executed() { 
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
			//fmt.Println("countfdowntimer")
			timerMap[orderReceived] = time.Now()
		}
	}()
	for {
		select {
		//Updates the queue, if the same kind of messages are sent simultaneously
		case orderExe = <-ExCommChans.ToMasterOrderExecutedChan: //This is on IP-message-form
	
			fmt.Println(SyncOrderMap.m)
			InLogicChans.ToTopLogicOrderChan <- orderExe /*The code that receives isnt made yet. Should handle optimization module there.*/
		

			SyncOrderMap.RLock()
			inQueue := SyncOrderMap.m[orderExe.Ord] //It will be nil if its not in the map
		
			SyncOrderMap.RUnlock()
			if inQueue == nil { //If its not in queue we should
				//InLogicChans.ToTopLogicOrderChan <- orderExe
				SyncOrderMap.Lock()
				SyncOrderMap.m[orderExe.Ord] = orderExe.Ip
				fmt.Println(SyncOrderMap.m)
				SyncOrderMap.Unlock()
			}
			countdownChan <- orderExe
			ExMasterChans.ToCommOrderExecutedConfirmedChan <- orderExe
		default:
			for ipOrder, timestamp := range timerMap {
				if time.Since(timestamp) > time.Millisecond*500 { //RLock before and after this if?
					delete(timerMap, ipOrder)
					SyncOrderMap.Lock()
					delete(SyncOrderMap.m, ipOrder.Ord)
					SyncOrderMap.Unlock()
					fmt.Println("...")
				}
			}
			time.Sleep(25 * time.Millisecond) // Change to optimize
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
	//localSlaveSlice := m.SlaveIp
	//timerMap := make(map[*UDPAddr]time.Time)               //timers for each IP
	/*
	startCountdownChan := make(chan bool)
	countdownFinishedChan := make(chan bool)
	allSlavesAnsweredChan := make(chan bool)
	startReceivingChan := make(chan [][N_FLOORS][2]bool, N_ELEV+1)
	countingSlaveSlice := make(chan [][N_FLOORS][2]bool, N_ELEV+1)
	var hasSentAlert bool
	timer := make(<-chan time.Time)
	*/

	//Get the total shit if it has been updated, and the slaves need to know
	go func() {
		for {
			select {
			case <-InLogicChans.ExternalListIsUpdated: //COMING FROM LOCAL STATEMACHINE -  THIS NEEDS TO BE MADE - FEEL FREE TO CHANGE NAME
			//	fmt.Println("I 			GET 			MADDAFAKKINGS 			HERE")

				//localSlaveMap.ExternalList = m.Get_external_list()
				ExMasterChans.ToCommOrderListChan <- IpOrderList{nil, m.ExternalList} //TO THE COMMUNICATION MODULE --  THIS NEEDS TO BE MADE - FEEL FREE TO CHANGE NAME
				//fmt.Println("external list in master updated outgoing",m.ExternalList)
				//startReceivingChan <- localSlaveSlice.ExternalList
				//startCountdownChan <- true
				//case <-countdownFinishedChan:
				//ExMasterChans.ToCommOrderListChan <- localSlaveSlice //TO THE COMMUNICATION MODULE --  THIS NEEDS TO BE MADE - FEEL FREE TO CHANGE NAME
			}

		}
	}()
	/*
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
					checkingSlaveIps =  [N_ELEV] bool
					receivedOrder := ipOrdList.ExternalList
					allListsMatch = true


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


					if len(countingSlaveSlice) == 1 && !hasSentAlert { //Only LightArray Remains
						allSlavesAnsweredChan <- true
						hasSentAlert = true
					}
				}
				time.Sleep(25 * time.Millisecond)
			}
		}()*/
}

func Slave_order_arrays_incoming(s *Slave) {
	//var NewExternalList map[*UDPAddr]*[N_FLOORS][2]bool //Think sending the master could be good, but master isnt a good name
	go func() {
		for {

			NewExternalList := <-ExCommChans.ToSlaveOrderListChan
			s.ExternalList = NewExternalList.ExternalList
			ExStateMChans.SingleExternalListChan <-NewExternalList.ExternalList[s.Nr]
			//ExStateMChans.SingleExternalListChan <- *NewExternalList.ExternalList[s.Get_ip()]

			//This access the last array containing ligths
			ExStateMChans.LightChan <- NewExternalList.ExternalList[N_ELEV]
			//fmt.Println("HAS SENT LIGHTARRAY DOWN TO STATEMACHINE")
			//Here we need to save all the information about the other slaves, and send our own to the statemachine

			//We dont want to confirm this
			//ExSlaveChans.ToCommOrderListReceivedChan <- NewExternalList
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
			//fmt.Println("				ORDER BEFORE DELETION", orderOut)
			//fmt.Println("Before deletion: ", SyncOrderMap.m)
			SyncOrderMap.Lock()
			delete(SyncOrderMap.m, orderOut)
			SyncOrderMap.Unlock()
			//fmt.Println("After deletion: ", SyncOrderMap.m)
		}
	}()

}

//Sends state if timer expires or state changes.
func Slave_state_updated() {
	sendAgainTimer := make(<-chan time.Time)
	var localCurrentState State
	for {
		select {
		case localCurrentState = <-ExStateMChans.CurrentStateChan:
			//fmt.Println("slave state updated inside case!!!!")
			ExSlaveChans.ToCommUpdatedStateChan <- localCurrentState
			//fmt.Println("after tocommupdatedstatechan")
			sendAgainTimer = time.After(50 * time.Millisecond)
		case currentStateReceived := <-ExCommChans.ToSlaveUpdateStateReceivedChan:
			if currentStateReceived.Sta == localCurrentState {
				sendAgainTimer = nil //Not sure if this is legal, will this send to channel if its set to nil??
				//fmt.Println("slave state updated")
			}
		case <-sendAgainTimer: //This will be sent when time runs out, I think.
			//fmt.Println("send again timer")

			ExSlaveChans.ToCommUpdatedStateChan <- localCurrentState
			//	fmt.Println("after send again timer")
			sendAgainTimer = time.After(500 * time.Millisecond)

		}
	}
}

//checks if some of the slaves sends Im Slave signal

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
W
func Interuption_killer(externalList [][N_FLOORS][2]bool) {
	InteruptChan := make(chan os.Signal) 
    signal.Notify(InteruptChan, os.Interrupt)
    signal.Notify(InteruptChan, syscall.SIGTERM)
    <-InteruptChan
    //what should be done when ctrl-c is pressed?????
    //<- goes here.
    //SystemInit()
    Write_to_file(externalList)

    fmt.Println("Got ctrl-c signal")
    cmd := exec.Command("mate-terminal", "-x", "./main")
    cmd.Start()
    os.Exit(0)
}
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
            Restart_system(m.ExternalList)

        }

        for _, value := range timer {
            if time.Since(value) > MAXWAIT_IM_HERE {
                ExMasterChans.RestartSystemChan <- m.ExternalList
            }
        }
    }
}


func Restart_system(externalList [][N_FLOORS][2]bool) {
    for {
        <-ExternalMasterChans.RestartSystemChan
        fmt.Println("ToCommRestartSystemChan <- true")
        Write_to_file(externalList)
        cmd := exec.Command("mate-terminal", "-x", EXE_FILE)
        defer os.Exit(666)
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
    defer file.Close()
    if err != nil {
        fmt.Println(err.Error(), " Error writing file")
    }

}
func Read_from_file() ([][N_FLOORS][2]bool, error) {
    file, err := os.Open("exList.txt")
    if err != nil {
        fmt.Println(err.Error(), " Error opening file")
    }
    defer file.Close()
    bArr := make([]byte, 256)
    n, err := file.Read(bArr)
    if err != nil {
        fmt.Println(err.Error(), " Error opening file")
    }
    var externalList [][N_FLOORS][2]bool
    err = Unmarshal(bArr[:n], &externalList)
    fmt.Println("in read from file ", externalList)
    if err != nil {
        fmt.Println(err.Error(), " Error Unmasrhaling file")
    }
    return externalList, err
}
func Get_my_ip() string {
    con1, _ := Dial("udp", "google.com:http")
    myIP := con1.LocalAddr().String()
    str := strings.Split(myIP, ":")
    myIP = str[0]
    con1.Close()
    return myIP
}
