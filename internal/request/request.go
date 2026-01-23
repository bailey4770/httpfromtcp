// Package request handles parsing a HTTP request from an io.Reader into a Request Struct
package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
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
	Body        []byte
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
		Body:    make([]byte, 0),
	}

	for req.state != doneParsing {
		if readToIndex >= len(buff) {
			newBuff := make([]byte, len(buff)*2)
			copy(newBuff, buff)
			buff = newBuff
		}

		numBytesRead, err := reader.Read(buff[readToIndex:])
		// Sample solution sees problem here - std io.Reader can read into buffer AND return EOF error.
		// We want to ensure we flush our buffer before handling EOF error

		if numBytesRead > 0 {
			readToIndex += numBytesRead

			numBytesParsed, parseErr := req.parse(buff[:readToIndex])
			if parseErr != nil {
				return nil, parseErr
			}

			copy(buff, buff[numBytesParsed:])
			readToIndex -= numBytesParsed
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != doneParsing {
					return nil, fmt.Errorf("incomplete request")
				}
				break
			}
			return nil, err
		}
	}
	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0

	for r.state != doneParsing {
		numBytesParsed, err := r.parseSingleChunk(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}

		if numBytesParsed == 0 {
			break
		}

		totalBytesParsed += numBytesParsed
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingleChunk(data []byte) (int, error) {
	switch r.state {
	case parsingRequestLine:
		numBytesParsed, err := r.parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if numBytesParsed > 0 {
			r.state = parsingHeaders
		}

		return numBytesParsed, nil

	case parsingHeaders:
		numBytesParsed, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if done {
			r.state = parsingBody
		}

		return numBytesParsed, nil

	case parsingBody:
		contentLenStr, ok := r.Headers.Get("Content-Length")
		if !ok {
			// We are assuming that since this header does not exist, there is no body
			r.state = doneParsing
			return len(data), nil
		}

		contentLength, err := strconv.Atoi(contentLenStr)
		if err != nil {
			return 0, err
		}

		r.Body = append(r.Body, data...)
		bodyLength := len(r.Body)
		if bodyLength > contentLength {
			return 0, fmt.Errorf("body is of length %d but header specified %d", bodyLength, contentLength)
		}

		if bodyLength == contentLength {
			r.state = doneParsing
		}

		return len(data), nil

	case doneParsing:
		return 0, errors.New("trying to read more data when request has finished parsing")

	default:
		return 0, errors.New("unknown state")
	}
}

func (r *Request) parseRequestLine(data []byte) (int, error) {
	if r.state != parsingRequestLine {
		return 0, errors.New("trying to read data in a done state")
	}

	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, nil
	}

	line := string(data[:idx])
	requestLine, err := requestLineFromString(line)
	if err != nil {
		return 0, err
	}

	numBytesParsed := idx + len(crlf)
	if numBytesParsed == 0 {
		return 0, nil
	}

	r.RequestLine = requestLine
	return numBytesParsed, nil
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
