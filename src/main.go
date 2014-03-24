package main

import (
	//"chansnstructs"
	"fmt"
)

func main() {
	externalList :=
		init(externalList, true)
}
func Init_elvator() {
	Channels_init()
	go Active_slave()

	blockingChan := make(chan bool)
	<-blocingChan
}
