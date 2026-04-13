package main

import (
	"bytes"
	"fmt"
	"net"
)

func HandleQuery(conn *net.UDPConn) error {
	buf := make([]byte, 512)
	n, src, err := conn.ReadFromUDP(buf)
	if err != nil {
		return err
	}

	request, err := ParsePacket(buf[:n])
	if err != nil {
		return err
	}

	response := Packet{
		Header: Header{
			ID:                 request.Header.ID,
			RecursionDesired:   true,
			RecursionAvailable: true,
			Response:           true,
		},
	}

	if len(request.Questions) == 0 {
		response.Header.Rescode = FORMERR
	} else {
		q := request.Questions[0]
		fmt.Printf("Received query: %+v\n", q)

		res, rerr := Lookup(q.Name, q.Type)
		if rerr != nil {
			response.Header.Rescode = SERVFAIL
		} else {
			response.Questions = append(response.Questions, q)
			response.Header.Rescode = res.Header.Rescode
			for _, rec := range res.Answers {
				fmt.Printf("Answer: %+v\n", rec)
				response.Answers = append(response.Answers, rec)
			}
			for _, rec := range res.Authorities {
				fmt.Printf("Authority: %+v\n", rec)
				response.Authorities = append(response.Authorities, rec)
			}
			for _, rec := range res.Resources {
				fmt.Printf("Resource: %+v\n", rec)
				response.Resources = append(response.Resources, rec)
			}
		}
	}

	resBuf := &bytes.Buffer{}
	if err = response.Write(resBuf); err != nil {
		return err
	}

	_, err = conn.WriteToUDP(resBuf.Bytes(), src)
	return err
}
