package main

import (
	. "chansnstructs"
	"fmt"
	//"net"
	. "os/exec"
	. "toplayer"
)

func main() {

	//var externalList map[*net.UDPAddr]*[N_FLOORS][2]bool

	Init_elevator(true)
	blockingChan := make(chan bool)
	<-blockingChan
	//dist_slaves()
}

func Init_elevator(firstRun bool) {
	fmt.Println("start init elev")
	Channels_init()

	Active_slave()
	fmt.Println("running")
	fmt.Println("finish init elev")
}
func dist_slaves() {
	//cmd2 := Command("mate-terminal").Run()
	//fmt.Println(cmd2.Output())

	//cmd = Command(EXE_FILE)

	cmd := Command("ssh", IP1)
	cmd.Run()
	cmd = Command("student")
	cmd.Run()
	cmd = Command("scp", "-r", "student@", LOCALHOST, ":", EXE_FILE, EXE_FILE)
	cmd.Run()
	fmt.Println(cmd.Output())
	cmd = Command(EXE_FILE)
	cmd.Run()
	/*
		cmd2 := Command("ssh", IP2)
		cmd2 = Command("scp", "-r", "student@", LOCALHOST, ":", EXE_FILE, EXE_FILE)
		cmd2 = Command(EXE_FILE)
		cmd2.Start()
	*/
}
