package main

import (
	. "chansnstructs"
	"fmt"
	//"net"
	//"os"
	. "os/exec"
	//"path/filepath"
	//"strings"
	//. "time"
	. "toplayer"
)

func main() {

	//var externalList map[*net.UDPAddr]*[N_FLOORS][2]bool

	//	Init_elevator(true)
	//	blockingChan := make(chan bool)
	//	<-blockingChan
	dist_slaves()

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
func dist_slaves() {
	//cmd2 := Command("mate-terminal").Run()
	//fmt.Println(cmd2.Output())

	//cmd = Command(EXE_FILE)

	/*
		//checkError(err)
		fmt.Println("2", cmd)
		//fmt.Println("cmd", cmd)
		//fmt.Println(
		Sleep(5 * Millisecond)
		cmd = Command("scp", "-r", "student@", LOCALHOST, ":", "main", "main")
		fmt.Println("3")
	*/
	//	cmd.Run()
	fmt.Println("finito")
	//fmt.Println(cmd.Output())
	//cmd = Command(EXE_FILE)
	//cmd.Run()
	/*
		cmd2 := Command("ssh", IP2)
		cmd2 = Command("scp", "-r", "student@", LOCALHOST, ":", EXE_FILE, EXE_FILE)
		cmd2 = Command(EXE_FILE)
		cmd2.Start()
}
