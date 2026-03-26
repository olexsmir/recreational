package main

import (
	"bytes"
	"encoding/binary"
)

type Question struct {
	Name  string
	Type  QueryType
	Class uint16
}

func ReadQuestion(r *bytes.Reader, packet []byte) (Question, error) {
	name, err := readName(r, packet)
	if err != nil {
		return Question{}, err
	}

	var qtype, qclass uint16
	_ = binary.Read(r, binary.BigEndian, &qtype)
	_ = binary.Read(r, binary.BigEndian, &qclass)

	return Question{
		Name:  name,
		Type:  QueryType(qtype),
		Class: qclass,
	}, nil
}

func (q Question) Write(b *bytes.Buffer) error {
	_ = writeName(b, q.Name)
	_ = binary.Write(b, binary.BigEndian, q.Type)
	_ = binary.Write(b, binary.BigEndian, q.Class)
	return nil
}
