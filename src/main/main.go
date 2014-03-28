package main

import (
	. "chansnstructs"
	"fmt"

	. "toplayer"
)

func main() {

	fmt.Println(Get_my_ip())
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
	Channels_init()

	Activate_slave()
}
