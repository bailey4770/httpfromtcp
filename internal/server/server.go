// Package server provides HTTP server
package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/bailey4770/httpfromtcp/internal/request"
	"github.com/bailey4770/httpfromtcp/internal/response"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

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

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("Error: could not parse request: %v", err)
		handlerErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			ErrorBody:  err.Error(),
		}
		handlerErr.write(conn)
		return
	}

	buf := bytes.NewBuffer([]byte{})
	handlerErr := s.handler(buf, req)
	if handlerErr != nil {
		handlerErr.write(conn)
		log.Print("Warning: handler errored. Error writen to connection")
		return
	}

	buf.Write([]byte("\r\n"))

	if err := response.WriteStatusLine(conn, response.StatusOK); err != nil {
		log.Printf("Error: could not write response status line: %v", err)
		handlerErr := &HandlerError{
			StatusCode: response.StatusInternalServerError,
			ErrorBody:  err.Error(),
		}
		handlerErr.write(conn)
		return
	}

	headers := response.GetDefaultHeaders(len(buf.Bytes()))
	if err := response.WriteHeaders(conn, headers); err != nil {
		log.Printf("Error: could not write response headers: %v", err)
		handlerErr := &HandlerError{
			StatusCode: response.StatusInternalServerError,
			ErrorBody:  err.Error(),
		}
		handlerErr.write(conn)
		return
	}

	if _, err := conn.Write(buf.Bytes()); err != nil {
		log.Printf("Error: could not write response body: %v", err)
		handlerErr := &HandlerError{
			StatusCode: response.StatusInternalServerError,
			ErrorBody:  err.Error(),
		}
		handlerErr.write(conn)
		return
	}

	log.Print("Successfuly wrote response and closed connection")
}

type HandlerError struct {
	StatusCode response.StatusCode
	ErrorBody  string
}

func (h *HandlerError) write(w io.Writer) {
	msg := fmt.Sprintf("HTTP/1.1 %d \r\n\r\n%s\r\n", h.StatusCode, h.ErrorBody)
	if _, err := w.Write([]byte(msg)); err != nil {
		log.Printf("Error: could not write error to connection: %v", err)
	}
}
