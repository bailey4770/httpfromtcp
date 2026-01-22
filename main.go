package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		log.Fatalf("Error: could not open file: %v", err)
	}

	messages := getLinesChannel(file)
	for msg := range messages {
		fmt.Printf("read: %s\n", msg)
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
