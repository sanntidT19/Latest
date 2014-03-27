package main

import (
	. "chansnstructs"
	"fmt"

	. "toplayer"
)

func main() {

	fmt.Println(GetMyIP())
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
