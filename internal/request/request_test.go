package request

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse(t *testing.T) {
	t.Run("Good GET Request line", func(t *testing.T) {
		r, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HTTPVersion)
	})

	t.Run("Good GET Request line with path", func(t *testing.T) {
		r, err := RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HTTPVersion)
	})

	t.Run("Good POST Request with path", func(t *testing.T) {
		r, err := RequestFromReader(strings.NewReader("POST /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "POST", r.RequestLine.Method)
		assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HTTPVersion)
	})

	t.Run("Invalid number of parts in request line", func(t *testing.T) {
		_, err := RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
		require.Error(t, err)
	})

	t.Run("Lowercase method", func(t *testing.T) {
		_, err := RequestFromReader(strings.NewReader("get /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
		require.Error(t, err)
	})

	t.Run("Number included in method", func(t *testing.T) {
		_, err := RequestFromReader(strings.NewReader("GET1 /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
		require.Error(t, err)
	})

	t.Run("Out of order request line", func(t *testing.T) {
		_, err := RequestFromReader(strings.NewReader("/coffee GET HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
		require.Error(t, err)
	})

	t.Run("Invalid protocol name", func(t *testing.T) {
		_, err := RequestFromReader(strings.NewReader("GET /coffee TCP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
		require.Error(t, err)
	})

	t.Run("Invalid HTTP version number", func(t *testing.T) {
		_, err := RequestFromReader(strings.NewReader("GET /coffee HTTP/1.3\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
		require.Error(t, err)
	})
}
