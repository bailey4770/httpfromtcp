// Package headers parses incoming data stream and fills a
package headers

import (
	"bytes"
	"errors"
	"strings"
)

const crlf = "\r\n"

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))

	switch idx {
	case -1:
		return 0, false, nil
	case 0:
		return len(crlf), true, nil
	}

	headerLine := string(data[:idx])
	key, value, found := strings.Cut(headerLine, ":")

	if !found {
		return 0, false, errors.New("could not find delimiter : in header line")
	}

	if key[len(key)-1] == ' ' {
		return 0, false, errors.New("cannot have space between key and colon")
	}

	h[strings.TrimSpace(key)] = strings.TrimSpace(value)

	return idx + len(crlf), false, nil
}

func NewHeaders() Headers {
	return make(Headers)
}
