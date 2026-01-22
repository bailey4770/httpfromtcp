package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
		messages := getLinesChannel(conn)
		for msg := range messages {
			fmt.Println(msg)
		}
	}
}

func getLinesChannel(file io.ReadCloser) <-chan string {
	messages := make(chan string)

	go func() {
		var currentLine string

		for {
			b := make([]byte, 8)

			n, err := file.Read(b)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if currentLine != "" {
						messages <- currentLine
					}
					close(messages)
					_ = file.Close()
					return
				}
				log.Fatalf("Error: reading from file: %v", err)
			}

			parts := strings.Split(string(b[:n]), "\n")
			for _, part := range parts[:len(parts)-1] {
				currentLine += part
				messages <- currentLine
				currentLine = ""
			}
			currentLine += parts[len(parts)-1]
		}
	}()

	return messages
}
