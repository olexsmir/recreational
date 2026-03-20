package main

import "bytes"

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
