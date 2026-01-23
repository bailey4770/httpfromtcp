// Package headers parses incoming data stream and fills a
package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Get(key string) (string, bool) {
	cleaned := strings.TrimSpace(strings.ToLower(key))

	val, ok := h[cleaned]
	return val, ok
}

func (h Headers) Set(key, value string) {
	if existingValue, ok := h.Get(key); ok {
		combinedValue := existingValue + ", " + strings.TrimSpace(value)
		h[strings.ToLower(key)] = combinedValue
	} else {
		h[strings.ToLower(key)] = strings.TrimSpace(value)
	}
}

func (h Headers) Override(key, value string) {
	h[strings.ToLower(key)] = strings.TrimSpace(value)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))

	switch idx {
	case -1:
		return 0, false, nil
	case 0:
		return len(crlf), true, nil
	}

	headerLine := string(data[:idx])
	key, value, found := strings.Cut(strings.TrimSpace(headerLine), ":")

	if !found {
		return 0, false, errors.New("could not find delimiter : in header line")
	}

	if key[len(key)-1] == ' ' {
		return 0, false, errors.New("cannot have space between key and colon")
	} else if !isValidFieldName(key) {
		return 0, false, fmt.Errorf("invalid field name. %s does not pass valid field name checks", key)
	}

	h.Set(key, value)
	return idx + len(crlf), false, nil
}

func isValidFieldName(s string) bool {
	if len(s) == 0 {
		return false
	}

	for i := 0; i < len(s); i++ {
		c := s[i]

		switch {
		case c >= 'A' && c <= 'Z':
		case c >= 'a' && c <= 'z':
		case c >= '0' && c <= '9':
		case c == '!' || c == '#' || c == '$' || c == '%' ||
			c == '&' || c == '\'' || c == '*' || c == '+' ||
			c == '-' || c == '.' || c == '^' || c == '_' ||
			c == '`' || c == '|' || c == '~':
		default:
			return false
		}
	}

	return true
}
