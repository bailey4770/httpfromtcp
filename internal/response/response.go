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

func Write(w *Writer, statusCode StatusCode, headers headers.Headers, body []byte) {
	headers.Set("Content-Length", strconv.Itoa(len(body)))

	if err := w.writeStatusLine(statusCode); err != nil {
		log.Printf("Error: could not write error status line to writer: %v", err)
	}
	if err := w.writeHeaders(headers); err != nil {
		log.Printf("Error: could not write error headers to writer: %v", err)
	}
	if _, err := w.writeBody(body); err != nil {
		log.Printf("Error: could not write error body to writer: %v", err)
	}
}

func GetDefaultHeaders() headers.Headers {
	headers := headers.NewHeaders()

	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")

	return headers
}

func StartStream(w *Writer, statusCode StatusCode, headers headers.Headers) {
	if err := w.writeStatusLine(statusCode); err != nil {
		log.Printf("Error: could not write error status line to writer: %v", err)
	}
	if err := w.writeHeaders(headers); err != nil {
		log.Printf("Error: could not write error headers to writer: %v", err)
	}
}

func (w *Writer) WriteChunkedBody(chunk []byte) (int, error) {
	total := 0

	n, err := fmt.Fprintf(w.Conn, "%x\r\n", len(chunk))
	if err != nil {
		return 0, nil
	}
	total += n

	n, err = w.writeBody(chunk)
	if err != nil {
		return 0, err
	}
	total += n

	n, err = fmt.Fprint(w.Conn, "\r\n")
	if err != nil {
		return 0, nil
	}
	total += n

	return total, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	n, err := w.writeBody([]byte("0\r\n\r\n"))
	if err != nil {
		return 0, err
	}

	return n, nil
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

func (w *Writer) writeBody(body []byte) (int, error) {
	n, err := w.Conn.Write([]byte(body))
	return n, err
}
