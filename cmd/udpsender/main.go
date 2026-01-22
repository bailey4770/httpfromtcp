package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("Error: could not resolve udp addr: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("Error: could not dial udp: %v", err)
	}
	defer func() { _ = conn.Close() }()

	input := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")

		line, err := input.ReadString('\n')
		if err != nil {
			log.Printf("Error: reading input: %v", err)
		}

		if _, err = conn.Write([]byte(line)); err != nil {
			log.Printf("Error: could not write to UDP conn: %v", err)
		}
	}
}
