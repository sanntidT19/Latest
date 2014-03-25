package main

import (
	. "chansnstructs"
	"fmt"
	. "net"
	"time"
	//)	. "toplayer"
)

func main() {
	var externalList map[*UDPAddr]*[N_FLOORS][2]bool
	//Init_elevator(externalList, true)
	Channels_init()
	Active_slave(externalList)
	time.Sleep(100 * time.Secound)
	//blockingChan := make(chan bool)
	//<-blockingChan
}

func Init_elevator(externalList map[*UDPAddr]*[N_FLOORS][2]bool, firstRun bool) {

	Channels_init()
	go Active_slave(externalList)

	time.Sleep(100 * time.Secound)
	//blockingChan := make(chan bool)
	//<-blockingChan

}
