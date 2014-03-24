package main

import (
	"fmt"
	. "net"
	"os/exec"
	"strconv"
	"time"

	//. "network"
)

const (
	PORT = ":20019"
)

/*
func test() {
	buf := make([]byte, 1024)
	var i int
	//go les(buf)
	//go skriv(i)

	/*
		if err != nil {
			fmt.Println(err.Error())
		}


	fmt.Println("hei")
	//defer c1.Close()
	//defer c2.Close()
	time.Sleep(10 * time.Second)
}
*/
func main() {
	nr := backup()

	cmd := exec.Command("mate-terminal", "-x", "go", "run", "commtest.go")
	cmd.Start()

	go primary(nr)

}

func backup() int {
	//fmt.Println("i have backup")
	var nr int
	buf := make([]byte, 1024)
	address, err := ResolveUDPAddr("udp", ":20019") //leser bare fra porten generellt
	conn, err := ListenUDP("udp", address)

	if err != nil {
		fmt.Println("backup conn create error:", err)
	}

	fmt.Println("backup")
	for {

		conn.SetReadDeadline(time.Now().Add(1000 * time.Millisecond))
		n, _, err := conn.ReadFromUDP(buf) //n contanis numbers of used bytes
		if err == nil {
			nr, _ = strconv.Atoi(string(buf[0:n]))
			fmt.Println("nr: ", nr)
		} else {
			break
		}
	}
	conn.Close()
	return nr
}

func primary(nr int) {
	address, _ := ResolveUDPAddr("udp", "129.241.187.255:20019")
	conn, _ := DialUDP("udp", nil, address)

	//fmt.Println(a)
	fmt.Println("primary")
	for {

		nr++
		fmt.Println(nr)

		_, err := conn.Write([]byte(strconv.Itoa(nr)))
		if err != nil {
			fmt.Println("fail")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

/*
func les(buf []byte) {
	for {
		fmt.Println("test")
		addr2, _ := ResolveUDPAddr("udp", PORT)
		c2, _ := ListenUDP("udp", addr2)

		//c2.SetReadDeadline(time.Now().Add(1000 * time.Millisecond)) //returns error if deadline is reached
		n, _, err := c2.ReadFromUDP(buf)
		if err == nil {
			nr, _ := strconv.Atoi(string(buf[0:n]))
			fmt.Println("nr: ", nr)
		} else {
			fmt.Println("fail")

		}

		fmt.Println(string(buf[0:n]))
		fmt.Println("ja")
	}
}
func skriv(i int) {
	for {
		addr, _ := ResolveUDPAddr("udp", "129.241.187.255"+PORT) //leser bare fra porten generellt
		//adddr, _ := ResolveUDPAddr("udp", "192.241.187.157")
		c1, _ := DialUDP("udp", nil, addr)
		c1.WriteToUDP([]byte("heiiiiii"), addr)
		c1.WriteToUDP([]byte(strconv.Itoa(i)), addr)
		i++
		time.Sleep(10 * time.Millisecond)

	}
}
*/
