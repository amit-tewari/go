package main

import (
	"fmt"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"time"
)

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func mustNewName(name string) dnsmessage.Name {
	n, err := dnsmessage.NewName(name)
	if err != nil {
		panic(err)
	}
	return n
}

func main() {
	ServerAddr, err := net.ResolveUDPAddr("udp", "LOCAL_IP:11001")
	CheckError(err)

	Conn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	go dns_reader(Conn)
	go dns_writer(Conn)

	defer Conn.Close()
	time.Sleep(time.Second * 1)
}
func dns_reader(conn *net.UDPConn) {
	for {
		buf := make([]byte, 1024)
		//on MAC, sleep is needed, since ReadFromUDP is non-blocking
		//time.Sleep(time.Millisecond * 450)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("From Reader", err)
		}
		var msg dnsmessage.Message
		err = msg.Unpack(buf)
		fmt.Printf("DNS Reader : %v\n", msg)
		//return
		var p dnsmessage.Parser
		if _, err := p.Start(buf); err != nil {
			CheckError(err)
			//panic(err)
		}
		for {
			q, err := p.Question()
			if err == dnsmessage.ErrSectionDone {
				break
			}
			if err != nil {
				panic(err)
			}

			fmt.Println("Found question for name", q.Name.String())
			if err := p.SkipAllQuestions(); err != nil {
				panic(err)
			}
			break
		}

		var gotIPs []net.IP
		for {
			h, err := p.AnswerHeader()
			if err == dnsmessage.ErrSectionDone {
				break
			}
			if err != nil {
				panic(err)
			}

			if (h.Type != dnsmessage.TypeA && h.Type != dnsmessage.TypeAAAA) || h.Class != dnsmessage.ClassINET {
				continue
			}

			/*if !strings.EqualFold(h.Name.String(), wantName) {
				if err := p.SkipAnswer(); err != nil {
					panic(err)
				}
				continue
			}*/

			switch h.Type {
			case dnsmessage.TypeA:
				r, err := p.AResource()
				if err != nil {
					panic(err)
				}
				gotIPs = append(gotIPs, r.A[:])
			case dnsmessage.TypeAAAA:
				r, err := p.AAAAResource()
				if err != nil {
					panic(err)
				}
				gotIPs = append(gotIPs, r.AAAA[:])
			}
		}

		fmt.Printf("Found A/AAAA records for name %v\n", gotIPs)
		fmt.Println("Received ", string(buf[0:n]), " from ", addr, gotIPs)
	}
}
func dns_writer(conn *net.UDPConn) {
	msg := dnsmessage.Message{
		Header: dnsmessage.Header{ID: 100, Response: false, OpCode: 0, Authoritative: true, RecursionDesired: true},
		Questions: []dnsmessage.Question{
			{
				Name:  mustNewName("servernet.se."),
				Type:  dnsmessage.TypeA,
				Class: dnsmessage.ClassINET,
				//},
				//{
				//	Name:  mustNewName("hotmail.com."),
				//	Type:  dnsmessage.TypeA,
				//	Class: dnsmessage.ClassINET,
			}}}
	buf, err := msg.Pack()
	fmt.Printf("DNS writer : %v\n", msg)
	CheckError(err)
	LocalAddr, err := net.ResolveUDPAddr("udp", "8.8.8.8:53")
	CheckError(err)
	_, err = conn.WriteToUDP(buf, LocalAddr)
	if err != nil {
		fmt.Println(msg, err)
	}
}
