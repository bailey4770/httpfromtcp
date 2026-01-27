package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bailey4770/httpfromtcp/internal/request"
	"github.com/bailey4770/httpfromtcp/internal/server"
)

const port = 8080

func main() {
	server, err := server.Serve(port, router)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer func() { _ = server.Close() }()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func router(req *request.Request) server.Handler {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		return yourProblemHandler
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		return myProblemHandler
	}

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		return chunkedHandler
	}

	if req.RequestLine.RequestTarget == "/video" {
		return videoHandler
	}

	return defaultHandler
}
