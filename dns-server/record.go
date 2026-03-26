package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

type QueryType uint16

const (
	AType     QueryType = 1
	NSType    QueryType = 2
	CNAMEType QueryType = 5
	MXType    QueryType = 15
	AAAAType  QueryType = 28
)

type Record struct {
	Name     string
	Type     QueryType
	Class    uint16
	TTL      uint32
	Data     string
	Priority uint16
}

func ReadRecord(r *bytes.Reader, packet []byte) (Record, error) {
	name, err := readName(r, packet)
	if err != nil {
		return Record{}, err
	}

	var rtype QueryType
	var class, rdlen uint16
	var ttl uint32
	_ = binary.Read(r, binary.BigEndian, &rtype)
	_ = binary.Read(r, binary.BigEndian, &class)
	_ = binary.Read(r, binary.BigEndian, &ttl)
	_ = binary.Read(r, binary.BigEndian, &rdlen)

	var data string
	var priority uint16
	switch rtype {
	case AType:
		var ip [4]byte
		_, _ = r.Read(ip[:])
		data = net.IP(ip[:]).String()

	case AAAAType:
		var ip [16]byte
		_, _ = r.Read(ip[:])
		data = net.IP(ip[:]).String()

	case NSType, CNAMEType:
		data, _ = readName(r, packet)

	case MXType:
		var pri uint16
		_ = binary.Read(r, binary.BigEndian, &pri)
		host, _ := readName(r, packet)
		priority = pri
		data = host

	default:
		buf := make([]byte, rdlen)
		_, _ = r.Read(buf)
		data = fmt.Sprintf("%x", buf)
	}

	return Record{
		Name:     name,
		Type:     rtype,
		Class:    class,
		TTL:      ttl,
		Data:     data,
		Priority: priority,
	}, nil
}

func (r Record) Write(b *bytes.Buffer) (int, error) {
	start := b.Len()
	switch r.Type {
	case AType:
		_ = r.writePremable(b)
		_ = binary.Write(b, binary.BigEndian, uint16(4))

		ip := net.ParseIP(r.Data).To4()
		if ip == nil {
			return 0, fmt.Errorf("invalid IPv4 address: %s", r.Data)
		}

		_, _ = b.Write(ip)

	case AAAAType:
		_ = r.writePremable(b)
		_ = binary.Write(b, binary.BigEndian, uint16(16))

		ip := net.ParseIP(r.Data).To16()
		if ip == nil {
			return 0, fmt.Errorf("invalid IPv6 address: %s", r.Data)
		}

		_, _ = b.Write(ip)

	case NSType, CNAMEType:
		_ = r.writePremable(b)

		encoded := encodeName(r.Data)
		_ = binary.Write(b, binary.BigEndian, uint16(len(encoded)))
		_, _ = b.Write(encoded)

	case MXType:
		_ = r.writePremable(b)

		encoded := encodeName(r.Data)
		_ = binary.Write(b, binary.BigEndian, uint16(2+len(encoded)))
		_ = binary.Write(b, binary.BigEndian, uint16(r.Priority))
		_, _ = b.Write(encoded)

	default:
		fmt.Printf("Skipping record: %+v\n", r)
	}

	return b.Len() - start, nil
}

func (r Record) writePremable(b *bytes.Buffer) error {
	// TODO: errors
	_ = writeName(b, r.Name)
	_ = binary.Write(b, binary.BigEndian, r.Type)
	_ = binary.Write(b, binary.BigEndian, r.Class)
	_ = binary.Write(b, binary.BigEndian, r.TTL)
	return nil
}

func readName(r *bytes.Reader, packet []byte) (string, error) {
	var labels []string
	for {
		length, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		if length == 0 {
			break
		}
		// pointer: top two bits set (0xC0)
		if length&0xC0 == 0xC0 {
			low, err := r.ReadByte()
			if err != nil {
				return "", err
			}
			offset := int(uint16(length&0x3F)<<8 | uint16(low))
			sub := bytes.NewReader(packet[offset:])
			name, err := readName(sub, packet)
			if err != nil {
				return "", err
			}
			labels = append(labels, name)
			break // pointer always ends the name
		}
		buf := make([]byte, length)
		if _, err := r.Read(buf); err != nil {
			return "", err
		}
		labels = append(labels, string(buf))
	}
	return strings.Join(labels, "."), nil
}

func encodeName(name string) []byte {
	var b bytes.Buffer
	for label := range strings.SplitSeq(name, ".") {
		_ = b.WriteByte(byte(len(label)))
		_, _ = b.WriteString(label)
	}
	_ = b.WriteByte(0)
	return b.Bytes()
}

// TODO: wrap the Buffer, to have the len == 512 guard
func writeName(w *bytes.Buffer, qname string) error {
	for label := range strings.SplitSeq(qname, ".") {
		llen := len(label)
		if llen > 0x3f {
			return fmt.Errorf("single label exceeds 63 characters of length")
		}
		_ = w.WriteByte(byte(llen))
		_, _ = w.Write([]byte(label))
	}
	_ = w.WriteByte(0)
	return nil
}
