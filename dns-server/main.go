package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:2053")
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	defer conn.Close()

	fmt.Println("Listening on :2053")
	for {
		if err := HandleQuery(conn); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}
}
