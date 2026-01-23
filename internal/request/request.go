// Package request handles parsing a HTTP request from an io.Reader into a Request Struct
package request

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"unicode"

	"github.com/bailey4770/httpfromtcp/internal/headers"
)

type requestState int

const (
	parsingRequestLine requestState = iota
	parsingHeaders
	parsingBody
	doneParsing
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       requestState
}

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

const (
	crlf       = "\r\n"
	bufferSize = 8
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	buff := make([]byte, bufferSize)
	readToIndex := 0

	req := &Request{
		Headers: headers.NewHeaders(),
		state:   parsingRequestLine,
	}

	for {
		if readToIndex == len(buff) {
			newBuff := make([]byte, len(buff)*2)
			_ = copy(newBuff, buff)
			buff = newBuff
		}

		n, err := reader.Read(buff[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				req.state = doneParsing
				break
			} else {
				return &Request{}, err
			}
		}

		readToIndex += n

		var consumed int

		switch req.state {
		case parsingRequestLine:
			consumed, err = req.parse(buff[:readToIndex])
			if err != nil {
				return &Request{}, err
			}

		case parsingHeaders:
			var done bool

			consumed, done, err = req.Headers.Parse(buff[:readToIndex])
			if err != nil {
				return &Request{}, err
			} else if done {
				req.state = parsingBody
			}
		}

		if consumed > 0 {
			copy(buff, buff[consumed:readToIndex])
			readToIndex -= consumed
		}

	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.state != parsingRequestLine {
		return 0, errors.New("trying to read data in a done state")
	}

	n, requestLine, err := parseRequestLine(data)
	if err != nil {
		return 0, err
	} else if n == 0 {
		return 0, nil
	}

	r.RequestLine = requestLine
	r.state = parsingHeaders
	return n, nil
}

func parseRequestLine(data []byte) (int, RequestLine, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, RequestLine{}, nil
	}

	line := string(data[:idx])
	requestLine, err := requestLineFromString(line)
	if err != nil {
		return 0, RequestLine{}, err
	}

	return idx + len(crlf), requestLine, nil
}

func requestLineFromString(line string) (RequestLine, error) {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return RequestLine{}, errors.New("bad request-line syntax. Not enough parts")
	}

	method := parts[0]
	if !isUpper(method) {
		return RequestLine{}, errors.New("method section should be all uppercase alphabetic characters")
	}

	addr := parts[1]

	protocolName := strings.Split(parts[2], "/")
	if protocolName[0] != "HTTP" {
		return RequestLine{}, errors.New("protocol is not HTTP")
	}
	if protocolName[1] != "1.1" {
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
