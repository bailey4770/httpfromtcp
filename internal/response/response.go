// Package response handles http response logic
package response

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/bailey4770/httpfromtcp/internal/headers"
)

type Writer struct {
	Conn net.Conn
}

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func (w *Writer) writeStatusLine(statusCode StatusCode) error {
	switch statusCode {
	case 200:
		_, err := w.Conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
		return err
	case 400:
		_, err := w.Conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		return err
	case 500:
		_, err := w.Conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		return err
	default:
		msg := fmt.Sprintf("HTTP/1.1 %d \r\n", statusCode)
		_, err := w.Conn.Write([]byte(msg))
		return err
	}
}

func (w *Writer) writeHeaders(headers headers.Headers) error {
	for key, value := range headers {
		header := key + ": " + value + "\r\n"
		if _, err := w.Conn.Write([]byte(header)); err != nil {
			return err
		}
	}

	_, err := w.Conn.Write([]byte("\r\n"))
	return err
}

func (w *Writer) writeBody(body string) error {
	_, err := w.Conn.Write([]byte(body))
	return err
}

func GetDefaultHeaders() headers.Headers {
	headers := headers.NewHeaders()

	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")

	return headers
}

func Write(w *Writer, statusCode StatusCode, headers headers.Headers, body string) {
	headers.Set("Content-Length", strconv.Itoa(len(body)))

	if err := w.writeStatusLine(statusCode); err != nil {
		log.Printf("Error: could not write error status line to writer: %v", err)
	}
	if err := w.writeHeaders(headers); err != nil {
		log.Printf("Error: could not write error headers to writer: %v", err)
	}
	if err := w.writeBody(body); err != nil {
		log.Printf("Error: could not write error body to writer: %v", err)
	}
}
