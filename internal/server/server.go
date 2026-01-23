// Package server provides HTTP server
package server

import (
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	isClosed atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		return nil, err
	}

	server := &Server{listener: listener}
	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.isClosed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for !s.isClosed.Load() {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Fatalf("Error: could not accept connection: %v", err)
		}
		log.Print("Connection accepted")

		go s.handle(conn)
	}

	log.Printf("Server was closed: stopped listening")
}

func (s *Server) handle(conn net.Conn) {
	msg := `HTTP/1.1 200 OK
Content-Type: text/plain
Content-Length: 13

Hello World!`

	_, err := conn.Write([]byte(msg))
	if err != nil {
		log.Printf("Error: could not write to conn: %v", err)
	}

	if err = conn.Close(); err != nil {
		log.Printf("Error: could not close conn: %v", err)
	}

	log.Print("Successfuly wrote response and closed connection")
}
