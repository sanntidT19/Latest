package main

import (
	. "chansnstructs"
	"fmt"
	//. "net"
	. "toplayer"
)

func main() {
	//var externalList map[*UDPAddr]*[N_FLOORS][2]bool

	Init_elevator(true)
	blockingChan := make(chan bool)
	<-blockingChan
}

func Init_elevator(firstRun bool) {
	fmt.Println("start init elev")
	Channels_init()

	Active_slave()
	fmt.Println("running")
	fmt.Println("finish init elev")
}
