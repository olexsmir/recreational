package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"net"
)

func main() {
	qname := "olexsmir.xyz"
	server := "8.8.8.8:53"

	conn, _ := net.Dial("udp", server)
	defer conn.Close()

	p := Packet{}
	p.Header.ID = 6666
	p.Header.RecursionDesired = true
	p.Questions = append(p.Questions, Question{
		Name:  qname,
		Type:  AType,
		Class: 1, // IN
	})

	buf := &bytes.Buffer{}
	if err := p.Write(buf); err != nil {
		fmt.Printf("failed to write packet: %v", err)
	}

	if _, err := conn.Write(buf.Bytes()); err != nil {
		fmt.Printf("failed to write to connection: %v", err)
	}

	res := make([]byte, 512)
	n, err := conn.Read(res)
	if err != nil {
		fmt.Printf("failed to read from connection: %v", err)
	}

	resPack, err := ParsePacket(res[:n])
	if err != nil {
		fmt.Printf("failed to parse packet: %v", err)
	}

	fmt.Printf("%+v\n", resPack.Header)
	for _, q := range resPack.Questions {
		fmt.Printf("%+v\n", q)
	}
	for _, r := range resPack.Answers {
		fmt.Printf("%+v\n", r)
	}
	for _, r := range resPack.Authorities {
		fmt.Printf("%+v\n", r)
	}
	for _, r := range resPack.Resources {
		fmt.Printf("%+v\n", r)
	}
}
