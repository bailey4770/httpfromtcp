// Package response handles http response logic
package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/bailey4770/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case 200:
		_, err := w.Write([]byte("HTTP/1.1 200 OK\r\n"))
		if err != nil {
			return err
		}
	case 400:
		_, err := w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		if err != nil {
			return err
		}
	case 500:
		_, err := w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		if err != nil {
			return err
		}
	default:
		msg := fmt.Sprintf("HTTP/1.1 %d \r\n", statusCode)
		_, err := w.Write([]byte(msg))
		if err != nil {
			return err
		}
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers.Set("Content-Length", strconv.Itoa(contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")

	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		header := key + ": " + value + "\r\n"
		if _, err := w.Write([]byte(header)); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte("\r\n")); err != nil {
		return err
	}

	return nil
}
