// Package request handles parsing a HTTP request from an io.Reader into a Request Struct
package request

import (
	"errors"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return &Request{}, err
	}

	parts := strings.Split(string(data), "\r\n")

	requestLine, err := parseRequestLine(parts[0])
	if err != nil {
		return &Request{}, err
	}

	return &Request{RequestLine: requestLine}, nil
}

func parseRequestLine(line string) (RequestLine, error) {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return RequestLine{}, errors.New("bad request line syntax. Not enough parts")
	}

	method := parts[0]
	if !isUpper(method) {
		return RequestLine{}, errors.New("method section should be all uppercase alphabetic characters")
	}

	addr := parts[1]

	protocolName := parts[2]
	if !strings.HasPrefix(protocolName, "HTTP/") {
		return RequestLine{}, errors.New("protocol name does not begin with HTTP/")
	}
	if !strings.HasSuffix(protocolName, "1.1") {
		return RequestLine{}, errors.New("HTTP version does not match 1.1")
	}

	return RequestLine{
			Method:        method,
			RequestTarget: addr,
			HTTPVersion:   "1.1",
		},
		nil
}

func isUpper(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) || !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}
