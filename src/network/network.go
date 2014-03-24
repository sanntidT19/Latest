package network

import (
	. "chansnstructs"
	. "encoding/json"
	"fmt"
	"math/rand"
	. "net"
	"time"
)

func SlaveAction(ExternalList map[*UDPAddr]*[N_FLOORS][2]bool) map[*UDPAddr]*[N_FLOORS][2]bool {
	localAddr := getLocalIp()
	//localIp, err := LookupHost(PORT)
	//localAddr, err := ResolveUDPAddr("udp", localIp)

	s := Slave{}
	s.IP = localAddr
	//buf := make([]byte, 1024)
	address, err := ResolveUDPAddr("udp", ":20019") //leser bare fra porten generellt
	conn, err := ListenUDP("udp", address)

	if err != nil {
		fmt.Println("backup conn create error:", err)
	}
	//localAddr is updated in receive
	go Receive()
	go Select_receive()
	go Select_send_slave()
	go Write_to_network(conn)

	ExSlaveChans.ToCommNetworkInitChan <- IpOrderList{}
	var ipOrdList IpOrderList
	ipVarList := make([]*UDPAddr, N_ELEV)
	var i int

	go func() {
		select {
		case ipOrdList = <-ExCommChans.ToSlaveNetworkInitRespChan:
			ipVarList[i] = ipOrdList.Ip
			i++
		}
	}()

	//VELGER MASTER
	tempBestMaster := ipVarList[0]
	for i := 1; i < len(ipVarList); i++ {
		if IpSum(tempBestMaster) < IpSum(ipVarList[i]) {
			tempBestMaster = ipVarList[i]
		}
	}
	if localAddr == tempBestMaster {
		go MasterAction(ExternalList)
	}

	conn.Close()
	return ExternalList
}

func MasterAction(ExternalList map[*UDPAddr]*[N_FLOORS][2]bool) {
	address, _ := ResolveUDPAddr("udp", "129.241.187.255"+PORT)
	conn, _ := DialUDP("udp", nil, address)

	fmt.Println("Master")
	go Write_to_network(conn)
	go Select_receive()
}

func IpSum(addr *UDPAddr) (sum byte) {
	bArr, _ := Marshal(*addr)
	for _, value := range bArr {
		sum += value
	}
	return sum
}
func getLocalIp() *UDPAddr {
	addr, _ := ResolveUDPAddr("udp", ":20019")
	return addr
}
func network_detector() {
	for {

	}
}

func Write_to_network(c Conn) {
	to_writing := <-ExNetChans.ToNetwork
	fmt.Println("to writing", string(to_writing))

	for {
		err := c.SetWriteDeadline(time.Now().Add(20 * time.Millisecond))
		_, err = c.Write(to_writing)
		if err != nil {
			fmt.Println(err.Error())
		} else {

		}
		time.Sleep(15 * time.Millisecond)
	}
}

func Receive() { //will error trigger if just read fails? or will it only go on deadline?
	buf := make([]byte, 1024)
	addr, _ := ResolveUDPAddr("udp", PORT)
	c, err := ListenUDP("udp", addr)
	fmt.Println("receive")
	defer c.Close()
	//this will also check if the master is still there.
	c.SetReadDeadline(time.Now().Add(60 * time.Millisecond)) //returns error if deadline is reached
	sendersAddr, localAddr, err := c.ReadFromUDP(buf)        //n contanis numbers of used bytes, fills buf with content on the connection
	fmt.Println("Local addr", localAddr)
	fmt.Println("Senders addr: ", sendersAddr)
	if err == nil { //if error is nil, read from buffer
		ExNetChans.ToComm <- buf

	} else {
		fmt.Println("Error: " + err.Error())
		//ExSlaveChans.ToSlaveImMasterChan <- false
	}

}

func Random_init(min int, max int) int { //gives a random int for waiting
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}
