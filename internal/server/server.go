// Package server provides HTTP server
package server

import (
	"log"
	"net"
	"sync/atomic"

	"github.com/bailey4770/httpfromtcp/internal/request"
	"github.com/bailey4770/httpfromtcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	listener net.Listener
	isClosed atomic.Bool
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		isClosed: atomic.Bool{},
		handler:  handler,
	}
	server.isClosed.Store(false)

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

	w := &response.Writer{Conn: conn}

	req, err := request.RequestFromReader(conn)
	if err != nil {
		headers := response.GetDefaultHeaders()
		response.Write(w, response.StatusBadRequest, headers, err.Error())
		return
	}

	s.handler(w, req)

	log.Print("Successfuly wrote response and closed connection")
}
