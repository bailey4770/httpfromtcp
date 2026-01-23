package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := min(len(cr.data), cr.pos+cr.numBytesPerRead)
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequestLineParse(t *testing.T) {
	t.Run("Good GET Request line", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HTTPVersion)
	})

	t.Run("Good GET Request line with path", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 1,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HTTPVersion)
	})

	t.Run("Many bytes per read", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n Lorem ipsum dolor sit amet consectetur adipiscing elit quisque faucibus ex sapien vitae pellentesque sem placerat in id cursus mi pretium tellus duis convallis tempus leo eu aenean sed diam urna tempor pulvinar vivamus fringilla lacus nec metus bibendum egestas.",
			numBytesPerRead: 1,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HTTPVersion)
	})

	t.Run("Good POST Request with path", func(t *testing.T) {
		reader := &chunkReader{
			data:            "POST /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 1,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "POST", r.RequestLine.Method)
		assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HTTPVersion)
	})

	t.Run("Invalid number of parts in request line", func(t *testing.T) {
		reader := &chunkReader{
			data:            "/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 1,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})

	t.Run("Lowercase method", func(t *testing.T) {
		reader := &chunkReader{
			data:            "get /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 1,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})

	t.Run("Number included in method", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET1 /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 1,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})

	t.Run("Out of order request line", func(t *testing.T) {
		reader := &chunkReader{
			data:            "/coffee GET HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 1,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})

	t.Run("Invalid protocol name", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET /coffee TCP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 1,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})

	t.Run("Invalid HTTP version number", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET /coffee HTTP/1.3\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 1,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})
}

func TestHeadersParse(t *testing.T) {
	t.Run("Standard Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "localhost:42069", r.Headers["host"])
		assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
		assert.Equal(t, "*/*", r.Headers["accept"])
	})

	t.Run("Empty Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\n\r\n\r\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		require.NotNil(t, r.Headers)
		assert.Len(t, r.Headers, 0)
	})

	t.Run("Duplicate Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nUser-Agent: firefox\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "firefox, curl/7.81.0", r.Headers["user-agent"])
		assert.Equal(t, "*/*", r.Headers["accept"])
	})

	t.Run("Malformed Header", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
			numBytesPerRead: 3,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})

	t.Run("Missing end of headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n",
			numBytesPerRead: 3,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})
}
