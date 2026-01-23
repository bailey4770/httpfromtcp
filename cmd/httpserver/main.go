package main

import (
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

func testHandler(w *response.Writer, req *request.Request) {
	headers := response.GetDefaultHeaders()
	headers.Override("Content-Type", "text/html")

	if req.RequestLine.RequestTarget == "/yourproblem" {
		msg := `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
		response.Write(w, response.StatusBadRequest, headers, msg)
		return
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		msg := `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
		response.Write(w, response.StatusInternalServerError, headers, msg)
		return
	}

	msg := `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

	response.Write(w, response.StatusOK, headers, msg)
}
