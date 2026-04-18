package main

import (
	"bytes"
	"fmt"
	"net"
)

func Lookup(qname string, qtype QueryType, server string) (Packet, error) {
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

func RecursiveLookup(qname string, qtype QueryType) (Packet, error) {
	ns := "198.41.0.4" // For now we're always starting with *a.root-servers.net*.
	for {
		fmt.Printf("attempting lookup of %v %s with ns %s\n", qtype, qname, ns)

		response, err := Lookup(qname, qtype, ns+":53")
		if err != nil {
			return Packet{}, err
		}

		if len(response.Answers) > 0 && response.Header.Rescode == NOERROR {
			return response, nil
		}

		if response.Header.Rescode == NXDOMAIN {
			return response, nil
		}

		if newNS, ok := response.GetResolvedNS(qname); ok {
			ns = newNS.String()
			continue
		}

		newNSName, ok := response.GetUnresolvedNS(qname)
		if !ok {
			return response, nil
		}

		recursive, err := RecursiveLookup(newNSName, AType)
		if err != nil {
			return response, nil
		}

		if ip, ok := recursive.GetRandomA(); ok {
			ns = ip.String()
		} else {
			return response, nil
		}
	}
}
