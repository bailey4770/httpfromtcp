package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHeaders(t *testing.T) {
	t.Run("Valid single header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "localhost:42069", headers["Host"])
		assert.Equal(t, len(data)-2, n)
		assert.False(t, done)
	})

	t.Run("Valid single header with white space", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("       Host: localhost:42069       \r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "localhost:42069", headers["Host"])
		assert.Equal(t, len(data)-2, n)
		assert.False(t, done)
	})

	t.Run("Valid new header with existing headers", func(t *testing.T) {
		headers := NewHeaders()
		headers["Host"] = "localhost:42069"
		data := []byte("Content-Type: json \r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "localhost:42069", headers["Host"])
		assert.Equal(t, "json", headers["Content-Type"])
		assert.Equal(t, len(data)-2, n)
		assert.False(t, done)
	})

	t.Run("Valid done", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, 2, n)
		assert.True(t, done)
	})

	t.Run("Invalid spacing header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("       Host : localhost:42069       \r\n\r\n")
		n, done, err := headers.Parse(data)
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})
}
