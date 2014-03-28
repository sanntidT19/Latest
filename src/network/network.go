package network

import (
	. "chansnstructs"
	"fmt"
	"math/rand"
	. "net"
	"time"
)

func Write_to_network() {

	localIp := "129.241.187.255" + PORT
	addr, err := ResolveUDPAddr("udp", localIp)
	c, err := DialUDP("udp", nil, addr)
	defer c.Close()
	if err != nil {
		//fmt.Println("Error in write to network: ", err.Error())
	}
	//fmt.Println("to writing", string(to_writing))

	//fmt.Println("before taking off send to network chan")
	for {
		to_writing := <-ExNetChans.ToNetwork
		
		//fmt.Println("in for write")

		err = c.SetWriteDeadline(time.Now().Add(20 * time.Millisecond))
		_, err = c.Write(to_writing)
		//fmt.Println("after write", string(to_writing))
		if err != nil {
			//fmt.Println("gogo write with error")
			fmt.Println(err.Error())
		} else {
			//fmt.Println("no errror in write")
		}
		//time.Sleep(10 * time.Millisecond)
	}
	//}
	//fmt.Println("exit write")
	//fmt.Println("after taking off send to network chan")
}

func Receive() { //will error trigger if just read fails? or will it only go on deadline?

	buf := make([]byte, 1024)
	addr, _ := ResolveUDPAddr("udp", PORT)

	c, err := ListenUDP("udp", addr)
	for {
		if err != nil {
			fmt.Println("error in receive: ", err.Error())
		}
		//fmt.Println("receive")

		//c.SetReadDeadline(time.Now().Add(60 * time.Millisecond)) //returns error if deadline is reached
		n, sendersAddr, err := c.ReadFromUDP(buf) //n contanis numbers of used bytes, fills buf with content on the connection
		if err == nil {
			//fmt.Println("original buffer", buf)

			ipByte := IpByteArr{sendersAddr, buf[:n]}

			ExNetChans.ToComm <- ipByte
			//fmt.Println("network Sends to Communication")

		} else {
			fmt.Println("receive666")
			//fmt.Println("Error: " + err.Error())
			//ExSlaveChans.ToSlaveImMasterChan <- false
		}
		//fmt.Println("receive3")
		//fmt.Println("Local addr in receive", sendersAddr)
		time.Sleep(25 * time.Millisecond)

	}
}

func Random_init(min int, max int) int { //gives a random int for waiting
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}
