package main

import (
	. "chansnstructs"
	"fmt"
	//"net"
	. "toplayer"
)

func main() {

	//	var externalList map[*net.UDPAddr]*[N_FLOORS][2]bool

	Init_elevator(true)
	blockingChan := make(chan bool)
	<-blockingChan

}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

func Init_elevator(firstRun bool) {
	fmt.Println("start init elev")
	Channels_init()

	Activate_slave()
	fmt.Println("running")
	fmt.Println("finish init elev")
}
