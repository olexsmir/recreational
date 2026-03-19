package main

import (
	_ "embed"
	"fmt"
)

//go:embed response_packet.txt
var respPack []byte

func main() {
	p, err := ParsePacket(respPack)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", p.Header)
	for _, q := range p.Questions {
		fmt.Printf("%+v\n", q)
	}
	for _, r := range p.Answers {
		fmt.Printf("%+v\n", r)
	}
	for _, r := range p.Authorities {
		fmt.Printf("%+v\n", r)
	}
	for _, r := range p.Resources {
		fmt.Printf("%+v\n", r)
	}
}
