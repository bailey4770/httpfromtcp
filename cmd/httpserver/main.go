package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bailey4770/httpfromtcp/internal/request"
	"github.com/bailey4770/httpfromtcp/internal/response"
	"github.com/bailey4770/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, testHandler)
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

func testHandler(w io.Writer, req *request.Request) *server.HandlerError {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		return &server.HandlerError{
			StatusCode: response.StatusBadRequest,
			ErrorBody:  "Your problem is not my problem\n",
		}
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		return &server.HandlerError{
			StatusCode: response.StatusInternalServerError,
			ErrorBody:  "Woopsie, my bad\n",
		}
	}

	successMsg := "All good, frfr\n"

	if _, err := w.Write([]byte(successMsg)); err != nil {
		return &server.HandlerError{
			StatusCode: response.StatusInternalServerError,
			ErrorBody:  "Could not write to writer\n",
		}
	}

	return nil
}
