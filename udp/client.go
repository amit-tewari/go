package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

func checkError(err error) {
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
	localAddr, err := net.ResolveUDPAddr("udp", os.Getenv("DNS_LOCAL")+":11001")
	checkError(err)

	Conn, err := net.ListenUDP("udp", localAddr)
	Conn.SetReadBuffer(1000000000000)
	checkError(err)
	go dnsReader(Conn)
	go dnsWriter(Conn)

	defer Conn.Close()
	time.Sleep(time.Second * 20)
}
func dnsReader(conn *net.UDPConn) {
	for {
		buf := make([]byte, 4000)
		//on MAC, sleep is needed, since ReadFromUDP is non-blocking
		//time.Sleep(time.Millisecond * 450)
		//n, addr, err := conn.ReadFromUDP(buf)
		_, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("From Reader", err)
		}
		//fmt.Println("Received ", n, "bytes from ", addr)
		var msg dnsmessage.Message
		err = msg.Unpack(buf)
		checkError(err)
		fmt.Printf("\n"+time.Now().Format("15:04:05.000000000")+" : ")
		if msg.Header.RCode == dnsmessage.RCodeSuccess {
			//fmt.Printf("%+s %+v", msg.Answers[0].Header.Name, msg.Answers[0].Body)
			fmt.Printf("%+v", msg)
			continue
		} else if msg.Header.RCode == dnsmessage.RCodeNameError {

			fmt.Printf("%+s : RCodeNameError", msg.Questions[0].Name)
			continue
		} else {
			fmt.Printf("%+s : RCode = %d", msg.Questions[0].Name, msg.Header.RCode)
			continue
		}
		fmt.Printf("\n\nDNS Reader : %v \n%T\n%+v \n\n%s\n", msg, msg.Answers[0].Body, msg.Answers[0].Body, msg.Answers[0].Header.Name)
		soa := msg.Answers[0].Body
		//soa{NS, MBox, Serial, Refresh, Retry, Expire, MinTTL} = &dnsmessage.SOAResource{msg.Answers[0].Body}
		fmt.Printf("soa %%T = %T\nsoa %%v = %+v\nsoa %%d = %d", soa, soa, soa)
		return

		var p dnsmessage.Parser
		if _, err := p.Start(buf); err != nil {
			checkError(err)
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
		//fmt.Println("Received ", string(buf[0:n]), " from ", addr, gotIPs)
	}
}
func dnsWriter(conn *net.UDPConn) {
	msg := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID: 100, Response: false, OpCode: 0, RecursionDesired: true},
		Questions: []dnsmessage.Question{
			{
				Type:  dnsmessage.TypeSOA,
				Class: dnsmessage.ClassINET,
			}}}
	msg.Questions[0].Name = mustNewName("google.com.")
	scanner := bufio.NewScanner(os.Stdin)
	var i uint16 = 0
	for scanner.Scan() {
		time.Sleep(time.Millisecond)
		// `Text` returns the current token, here the next line,
		// from the input.
		i++
		domain := scanner.Text()
		//fmt.Printf("%s", domain)
		msg.Header.ID = i % 65000
		msg.Questions[0].Name = mustNewName(domain + ".")
		buf, err := msg.Pack()
		//fmt.Printf("DNS writer : %v %d %s\n", msg, len(buf), msg.Questions[0].Name)
		checkError(err)
		DNSServerAddr, err := net.ResolveUDPAddr("udp",  os.Getenv("DNS_REMOTE")+":53")
		checkError(err)
		_, err = conn.WriteToUDP(buf, DNSServerAddr)
		if err != nil {
			fmt.Println(msg, err)
			time.Sleep(time.Second * 1)
		}
	}
	fmt.Print("All Read : ")
	fmt.Printf(time.Now().Format("15:04:05.000000000")+"\n")
}
