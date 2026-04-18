package main

import (
	"bytes"
	"net"
	"strings"
)

type Packet struct {
	Header      Header
	Questions   []Question
	Answers     []Record
	Authorities []Record
	Resources   []Record
}

func ParsePacket(packet []byte) (Packet, error) {
	r := bytes.NewReader(packet)

	var err error
	var p Packet

	p.Header, err = ReadHeader(r)
	if err != nil {
		return Packet{}, err
	}

	for i := 0; i < int(p.Header.Questions); i++ {
		q, err := ReadQuestion(r, packet)
		if err != nil {
			return Packet{}, err
		}
		p.Questions = append(p.Questions, q)
	}

	for i := 0; i < int(p.Header.Answers); i++ {
		a, err := ReadRecord(r, packet)
		if err != nil {
			return Packet{}, err
		}
		p.Answers = append(p.Answers, a)
	}

	for i := 0; i < int(p.Header.AuthoritativeEntries); i++ {
		ae, err := ReadRecord(r, packet)
		if err != nil {
			return Packet{}, err
		}
		p.Authorities = append(p.Authorities, ae)
	}

	for i := 0; i < int(p.Header.ResourceEntries); i++ {
		re, err := ReadRecord(r, packet)
		if err != nil {
			return Packet{}, err
		}
		p.Resources = append(p.Resources, re)
	}

	return p, nil
}

func (p *Packet) Write(b *bytes.Buffer) error {
	p.Header.Questions = uint16(len(p.Questions))
	p.Header.Answers = uint16(len(p.Answers))
	p.Header.AuthoritativeEntries = uint16(len(p.Authorities))
	p.Header.ResourceEntries = uint16(len(p.Resources))
	_ = p.Header.Write(b)

	for i := range p.Questions {
		_ = p.Questions[i].Write(b)
	}

	for i := range p.Answers {
		_, _ = p.Answers[i].Write(b)
	}

	for i := range p.Authorities {
		_, _ = p.Authorities[i].Write(b)
	}

	for i := range p.Resources {
		_, _ = p.Resources[i].Write(b)
	}

	return nil
}

func (p Packet) GetRandomA() (net.IP, bool) {
	for _, r := range p.Answers {
		if r.Type == AType {
			return net.ParseIP(r.Data), true
		}
	}
	return nil, false
}

func (p Packet) GetResolvedNS(qname string) (net.IP, bool) {
	for _, ns := range p.getNS(qname) {
		host := ns[1]
		for _, r := range p.Resources {
			if r.Type == AType && r.Name == host {
				return net.ParseIP(r.Data), true
			}
		}
	}
	return nil, false
}

func (p Packet) GetUnresolvedNS(qname string) (string, bool) {
	ns := p.getNS(qname)
	if len(ns) == 0 {
		return "", false
	}
	return ns[0][1], true
}

func (p Packet) getNS(qname string) [][2]string {
	var res [][2]string
	for _, r := range p.Authorities {
		if r.Type == NSType && strings.HasPrefix(qname, r.Name) {
			res = append(res, [2]string{r.Name, r.Data})
		}
	}
	return res
}
