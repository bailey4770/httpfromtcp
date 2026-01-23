package main

import (
	"fmt"
	"log"
	"net"

	"github.com/bailey4770/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("Error: could not create listener: %v", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error: could not accept connection: %v", err)
		}
		log.Print("Connection accepted")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("Error: could not get HTTP request from conn: %v", err)
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %v\n", req.RequestLine.Method)
		fmt.Printf("- Target: %v\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %v\n", req.RequestLine.HTTPVersion)

		fmt.Println("Headers:")
		for k, v := range req.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}

		fmt.Println("Body:")
		fmt.Println(string(req.Body))
	}
}
