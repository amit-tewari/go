package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func main() {
	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:11001")
	CheckError(err)

	Conn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	go reader(Conn)
	go writer(Conn, "10001")
	go writer(Conn, "10002")

	defer Conn.Close()
	for {
		time.Sleep(time.Second * 10)
	}
}
func reader(conn *net.UDPConn) {
	for {
		//on MAC, this sleep is needed, since ReadFromUDP is non-blocking
		//time.Sleep(time.Millisecond * 450)
		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("From Reader", err)
		}
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)
	}

}
func writer(conn *net.UDPConn, port string) {
	i := 0
	for {
		msg := strconv.Itoa(i)
		i++
		buf := []byte(msg)
		LocalAddr, err := net.ResolveUDPAddr("udp", "localhost:"+port)
		CheckError(err)
		_, err = conn.WriteToUDP(buf, LocalAddr)
		if err != nil {
			fmt.Println(msg, err)
		}
		time.Sleep(time.Second * 1)
	}
}
