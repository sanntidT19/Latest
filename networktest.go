package main

import (
	. "chansnstructs"
	"fmt"
	. "network"
	//. "net"
	//"os"
	//"os/signal"
	//"syscall"
	//"time"
	//"os/exec"
	. "encoding/json"
)

func main() {

	//dette fungerer
	message := "i am master"
	byteMessage, _ := Marshal(message)
	prefix, _ := Marshal("iam")

	byteMessage = append(prefix, byteMessage...)

	if string(byteMessage[1:4]) == "iam" {
		fmt.Println("check")
	}
	sendConnection := Network_init()
	Channals_init()
	fmt.Println("test")
	go Select_send_master()
	go Write_to_network(sendConnection)
	go Select_receive()
	go Receive()
	//s := Slave{ nil, nil, [true, false, true,false], 2, 3}
	ExMasterChans.ToCommImMasterChan <- "i am Master"

	temp := <-ExCommChans.ToSlaveImMasterChan
	fmt.Println("Received form net: ", temp)

	blockChan := make(chan bool)
	<-blockChan

}
