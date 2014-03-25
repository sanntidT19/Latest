package network

import (
	. "chansnstructs"
	. "encoding/json"
	"fmt"
	. "net"
	//"time"
)

var InCommChans InternalCommunicationChannels

type InternalCommunicationChannels struct {
	newExternalList              chan IpOrderList
	slaveToStateExMasterChanshan chan int //send input to statemachine
}

func Select_send_master() {

	for {
		select {
		//Master

		case externalOrderList := <-ExMasterChans.ToCommOrderListChan:
			Send_order(externalOrderList)
		case Ord := <-ExMasterChans.ToCommOrderExecutedConfirmedChan:
			Send_order_executed_confirmation(Ord)
		case <-ExMasterChans.ToCommImMasterChan:
			Send_im_master()
		case state := <-ExMasterChans.ToCommUpdateStateReceivedChan:
			Send_update_state_received(state)
		}
	}
}
func Select_send_slave() {
	for {
		select {
		//Slave
		case ipOrdList := <-ExSlaveChans.ToCommNetworkInitChan:
			Send_network_init(ipOrdList)
		case ipOrdList := <-ExSlaveChans.ToCommNetworkInitRespChan:
			Send_network_init_response(ipOrdList)
		case externalOrderList := <-ExSlaveChans.ToCommOrderListReceivedChan:
			Send_order_received(externalOrderList)
		case Ord := <-ExSlaveChans.ToCommOrderExecutedChan:
			Send_order_executed(Ord)
		case Ord := <-ExSlaveChans.ToCommOrderExecutedReConfirmedChan: //	STRANGE NAME ???!!!???
			Send_order_executed_reconfirmed(Ord)
		case Ord := <-ExSlaveChans.ToCommExternalButtonPushedChan:
			Send_ex_button_push(Ord)
		case <-ExSlaveChans.ToCommImSlaveChan:
			Send_im_slave()
		case state := <-ExSlaveChans.ToCommUpdatedStateChan:
			Send_update_state(state)

		}
	}
}

func internal_comm_chans_init() {
	InCommChans.newExternalList = make(chan IpOrderList)
	InCommChans.slaveToStateExMasterChanshan = make(chan int) //send input to statemachine
}

func Send_network_init(ordList IpOrderList) {
	byteOrder, _ := Marshal(ordList.ExternalList)
	prefix, _ := Marshal("ini")
	ExNetChans.ToNetwork <- append(prefix, byteOrder...)
}

//From slave
func Send_network_init_response(ordList IpOrderList) {
	byteOrder, _ := Marshal(ordList)
	prefix, _ := Marshal("inr")
	ExNetChans.ToNetwork <- append(prefix, byteOrder...)
}

//to slave
func Send_order(ordList IpOrderList) { //send exectuionOrderList
	byteOrder, _ := Marshal(ordList.ExternalList)
	prefix, _ := Marshal("ord")
	ExNetChans.ToNetwork <- append(prefix, byteOrder...)
}

//To master
func Send_order_received(ordList IpOrderList) {
	byteMessage, _ := Marshal(ordList)
	prefix, _ := Marshal("ore")
	ExNetChans.ToNetwork <- append(prefix, byteMessage...)
}

//To master
func Send_order_executed(ord Order) {
	byteMessage, _ := Marshal(ord)
	prefix, _ := Marshal("oex")
	ExNetChans.ToNetwork <- append(prefix, byteMessage...)
}

//To slave
func Send_order_executed_confirmation(ord IpOrderMessage) {
	byteMessage, _ := Marshal(ord.Ord)
	prefix, _ := Marshal("eco")
	ExNetChans.ToNetwork <- append(prefix, byteMessage...)
}

//to master
func Send_order_executed_reconfirmed(ord Order) {
	byteMessage, _ := Marshal(ord)
	prefix, _ := Marshal("oce")
	ExNetChans.ToNetwork <- append(prefix, byteMessage...)
}

//to slave
func Send_im_master() {
	byteMessage, _ := Marshal("i am master")
	prefix, _ := Marshal("iam")
	ExNetChans.ToNetwork <- append(prefix, byteMessage...)

}

// to master
func Send_im_slave() {
	byteMessage, _ := Marshal("i am slave")
	prefix, _ := Marshal("ias")
	ExNetChans.ToNetwork <- append(prefix, byteMessage...)
}

//to master
func Send_update_state(sta IpState) {
	byteIpOrder, _ := Marshal(sta)
	prefix, _ := Marshal("ust")
	ExNetChans.ToNetwork <- append(prefix, byteIpOrder...)
}

//to slave
func Send_update_state_received(state IpState) {
	byteIpOrder, _ := Marshal(state)
	prefix, _ := Marshal("sus")
	ExNetChans.ToNetwork <- append(prefix, byteIpOrder...)
}

//To master
func Send_ex_button_push(ord Order) {
	byteMessage, _ := Marshal(ord)
	prefix, _ := Marshal("ebp")
	ExNetChans.ToNetwork <- append(prefix, byteMessage...)
}

func Send_button_pressed(ord IpOrderMessage) {
	byteIpOrder, _ := Marshal(ord.Ord)
	prefix, _ := Marshal("bpc")
	ExNetChans.ToNetwork <- append(prefix, byteIpOrder...)
}
func Send_restart_system() {
	trigger, _ := Marshal(true)
	prefix, _ := Marshal("ree")
	ExNetChans.ToNetwork <- append(prefix, trigger...)
}
func Select_receive() {

	for {

		ipByteArr := <-ExNetChans.ToComm
		byteArr := ipByteArr.Barr
		ip := ipByteArr.Ip

		Decrypt_message(byteArr, ip)
	}
}

func Decrypt_message(message []byte, addr *UDPAddr) {

	switch {
	case string(message[1:4]) == "ini":

		noPrefix := message[5:]
		ordList := IpOrderList{}
		_ = Unmarshal(noPrefix, &ordList)
		ordList.Ip = addr
		fmt.Println("prefix = ini")
		ExCommChans.ToSlaveNetworkInitRespChan <- ordList

	case string(message[1:4]) == "inr":
		noPrefix := message[5:]
		ordList := IpOrderList{}
		_ = Unmarshal(noPrefix, &ordList)
		ordList.Ip = addr
		fmt.Println("prefix = inr")
		ExCommChans.ToSlaveNetworkInitChan <- ordList

	case string(message[1:4]) == "ord":
		noPrefix := message[5:]
		ordList := IpOrderList{}
		_ = Unmarshal(noPrefix, &ordList)
		ordList.Ip = addr
		fmt.Println("prefix = ord")
		ExCommChans.ToSlaveOrderListChan <- ordList

	case string(message[1:4]) == "ore":
		noPrefix := message[5:]
		ordList := IpOrderList{}
		_ = Unmarshal(noPrefix, &ordList)
		ordList.Ip = addr
		fmt.Println("prefix = ore")
		ExCommChans.ToMasterOrderListReceivedChan <- IpOrderList{addr, ordList.ExternalList}

	case string(message[1:4]) == "oex":
		noPrefix := message[5:]
		ord := Order{}
		_ = Unmarshal(noPrefix, &ord)
		fmt.Println("prefix = oex")
		ExCommChans.ToMasterOrderExecutedChan <- IpOrderMessage{addr, ord}

	case string(message[1:4]) == "eco":
		noPrefix := message[5:]
		ord := IpOrderMessage{}
		_ = Unmarshal(noPrefix, &ord)
		ord.Ip = addr
		fmt.Println("prefix = eco")
		ExCommChans.ToSlaveOrderExecutedConfirmedChan <- IpOrderMessage{addr, ord.Ord}

	case string(message[1:4]) == "ebp":
		noPrefix := message[5:]
		ord := Order{}
		_ = Unmarshal(noPrefix, &ord)
		fmt.Println("prefix = ebp")
		ExCommChans.ToMasterExternalButtonPushedChan <- IpOrderMessage{addr, ord}

	case string(message[1:4]) == "oce":
		noPrefix := message[5:]
		ord := Order{}
		_ = Unmarshal(noPrefix, &ord)
		fmt.Println("prefix = oce")
		ExCommChans.ToMasterOrderExecutedReConfirmedChan <- IpOrderMessage{addr, ord}

	case string(message[1:4]) == "ust":
		noPrefix := message[5:]
		ipSta := IpState{}
		_ = Unmarshal(noPrefix, &ipSta)
		ipSta.Ip = addr
		fmt.Println("prefix = ust")
		ExCommChans.ToMasterUpdateState <- ipSta

	case string(message[1:4]) == "sus":
		noPrefix := message[5:]
		ipSta := IpState{}
		_ = Unmarshal(noPrefix, &ipSta)
		ipSta.Ip = addr
		fmt.Println("prefix = ust")
		ExCommChans.ToSlaveUpdateStateReceivedChan <- ipSta

	case string(message[1:4]) == "iam":
		noPrefix := message[5:]
		stringMessage := string(noPrefix)
		fmt.Println("prefix = iam")
		ExCommChans.ToSlaveImMasterChan <- stringMessage

	case string(message[1:4]) == "ias":
		noPrefix := message[5:]
		ipOrd := IpOrderMessage{}
		_ = Unmarshal(noPrefix, &ipOrd)
		ipOrd.Ip = addr
		fmt.Println("prefix = ias")
		ExCommChans.ToMasterImSlaveChan <- ipOrd

	case string(message[1:4]) == "bpc":
		noPrefix := message[5:]
		ipOrd := IpOrderMessage{}
		_ = Unmarshal(noPrefix, &ipOrd)
		ipOrd.Ip = addr
		ExCommChans.ToSlaveButtonPressedConfirmedChan <- ipOrd

	case string(message[1:4]) == "ree":
		noPrefix := message[5:]
		ipOrd := IpOrderList{}
		_ = Unmarshal(noPrefix, &ipOrd)
		ipOrd.Ip = addr
		ExCommChans.ToSlaveRestartSystemTriggerChan <- ipOrd

	default:

		fmt.Println("ingen caser utløst; prefix er: " /*string(message[1:4])()*/)
	}

}
