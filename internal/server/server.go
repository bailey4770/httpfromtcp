// Package server provides HTTP server
package server

import (
	"log"
	"net"
	"sync/atomic"

	"github.com/bailey4770/httpfromtcp/internal/response"
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
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.isClosed.Load() {
				return
			}
			log.Printf("Error: could not accept connection: %v", err)
			continue
		}
		log.Print("Connection accepted")

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	if err := response.WriteStatusLine(conn, response.StatusOK); err != nil {
		log.Printf("Error: could not write status line: %v", err)
		return
	}

	headers := response.GetDefaultHeaders(0)
	if err := response.WriteHeaders(conn, headers); err != nil {
		log.Printf("Error: could not write headers: %v", err)
		return
	}

	log.Print("Successfuly wrote response and closed connection")
}
