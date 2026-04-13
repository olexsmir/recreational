package main

import (
	"bytes"
	"fmt"
	"net"
)

const server = "8.8.8.8:53"

func Lookup(qname string, qtype QueryType) (Packet, error) {
	conn, _ := net.Dial("udp", server)
	defer conn.Close()

	p := Packet{}
	p.Header.ID = 6666
	p.Header.RecursionDesired = true
	p.Questions = append(p.Questions, Question{
		Name:  qname,
		Type:  qtype,
		Class: 1, // IN
	})

	buf := &bytes.Buffer{}
	if err := p.Write(buf); err != nil {
		return Packet{}, fmt.Errorf("failed to write packet: %v", err)
	}

	if _, err := conn.Write(buf.Bytes()); err != nil {
		fmt.Printf("failed to write to connection: %v", err)
	}

	res := make([]byte, 512)
	n, err := conn.Read(res)
	if err != nil {
		fmt.Printf("failed to read from connection: %v", err)
	}

	return ParsePacket(res[:n])
}
