package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type ResultCode = uint8

const (
	NOERROR ResultCode = iota
	FORMERR
	SERVFAIL
	NXDOMAIN
	NOTIMP
	REFUSED
)

type Header struct {
	ID uint16

	RecursionDesired    bool
	TruncatedMessage    bool
	AuthoritativeAnswer bool
	OPCode              uint8
	Response            bool
	Rescode             ResultCode
	CheckingDisabled    bool
	AuthedData          bool
	Z                   bool
	RecursionAvailable  bool

	Questions            uint16
	Answers              uint16
	AuthoritativeEntries uint16
	ResourceEntries      uint16
}

func ReadHeader(r *bytes.Reader) (Header, error) {
	var h Header
	if err := binary.Read(r, binary.BigEndian, &h.ID); err != nil {
		return h, fmt.Errorf("reading ID: %w", err)
	}

	var flags uint16
	if err := binary.Read(r, binary.BigEndian, &flags); err != nil {
		return h, fmt.Errorf("reading flags: %w", err)
	}
	h.unpackFlags(flags)

	if err := binary.Read(r, binary.BigEndian, &h.Questions); err != nil {
		return h, fmt.Errorf("reading questions: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &h.Answers); err != nil {
		return h, fmt.Errorf("reading answers: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &h.AuthoritativeEntries); err != nil {
		return h, fmt.Errorf("reading auth entries: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &h.ResourceEntries); err != nil {
		return h, fmt.Errorf("reading resource entries: %w", err)
	}

	return h, nil
}

func (h *Header) Write(b *bytes.Buffer) error {
	_ = binary.Write(b, binary.BigEndian, h.ID)

	_ = b.WriteByte((b2u8(h.RecursionDesired)) |
		(b2u8(h.TruncatedMessage) << 1) |
		(b2u8(h.AuthoritativeAnswer) << 2) |
		(h.OPCode << 3) |
		(b2u8(h.Response) << 7))

	_ = b.WriteByte(h.Rescode |
		(b2u8(h.CheckingDisabled) << 4) |
		(b2u8(h.AuthedData) << 5) |
		(b2u8(h.Z) << 6) |
		(b2u8(h.RecursionAvailable) << 7))

	_ = binary.Write(b, binary.BigEndian, h.Questions)
	_ = binary.Write(b, binary.BigEndian, h.Answers)
	_ = binary.Write(b, binary.BigEndian, h.AuthoritativeEntries)
	_ = binary.Write(b, binary.BigEndian, h.ResourceEntries)
	return nil
}

func (h *Header) unpackFlags(flags uint16) {
	h.RecursionDesired = flags&(1<<8) != 0
	h.TruncatedMessage = flags&(1<<9) != 0
	h.AuthoritativeAnswer = flags&(1<<10) != 0
	h.OPCode = uint8((flags >> 11) & 0xF)
	h.Response = flags&(1<<15) != 0
	h.Rescode = ResultCode(flags & 0xF)
	h.CheckingDisabled = flags&(1<<4) != 0
	h.AuthedData = flags&(1<<5) != 0
	h.Z = flags&(1<<6) != 0
	h.RecursionAvailable = flags&(1<<7) != 0
}

func (h Header) packFlags() uint16 {
	var flags uint16
	if h.RecursionDesired {
		flags |= 1 << 8
	}
	if h.TruncatedMessage {
		flags |= 1 << 9
	}
	if h.AuthoritativeAnswer {
		flags |= 1 << 10
	}
	flags |= uint16(h.OPCode) << 11
	if h.Response {
		flags |= 1 << 15
	}
	flags |= uint16(h.Rescode)
	if h.CheckingDisabled {
		flags |= 1 << 4
	}
	if h.AuthedData {
		flags |= 1 << 5
	}
	if h.Z {
		flags |= 1 << 6
	}
	if h.RecursionAvailable {
		flags |= 1 << 7
	}
	return flags
}

func b2u8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
