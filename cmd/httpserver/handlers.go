package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/bailey4770/httpfromtcp/internal/headers"
	"github.com/bailey4770/httpfromtcp/internal/request"
	"github.com/bailey4770/httpfromtcp/internal/response"
)

func testHandler(w *response.Writer, req *request.Request) {
	headers := response.GetDefaultHeaders()
	headers.Override("Content-Type", "text/html")

	if req.RequestLine.RequestTarget == "/yourproblem" {
		msg := `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
		response.Write(w, response.StatusBadRequest, headers, []byte(msg))
		return
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		msg := `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
		response.Write(w, response.StatusInternalServerError, headers, []byte(msg))
		return
	}

	msg := `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

	response.Write(w, response.StatusOK, headers, []byte(msg))
}

func chunkedHandler(w *response.Writer, req *request.Request) {
	if !strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		return
	}

	subdomain := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	url := "https://httpbin.org" + subdomain

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error: could not GET from %v: %v", url, err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	respHeaders := headers.NewHeaders()
	for k, v := range resp.Header.Clone() {
		for _, i := range v {
			respHeaders.Set(k, i)
		}
	}

	respHeaders.Remove("Content-Length")
	respHeaders.Override("Transfer-Encoding", "chunked")
	respHeaders.SetTrailers("X-Content-Sha256", "X-Content-Length")

	response.StartStream(w, response.StatusCode(resp.StatusCode), respHeaders)
	var fullBody []byte

	const bufferSize = 1024
	buf := make([]byte, bufferSize)

	for {
		numBytesRead, err := resp.Body.Read(buf)
		log.Printf("Read %d from reponse body with error: %v", numBytesRead, err)
		if numBytesRead > 0 {
			fullBody = append(fullBody, buf[:numBytesRead]...)
			if _, err := w.WriteChunkedBody(buf[:numBytesRead]); err != nil {
				log.Printf("Error: could not write chunked body response: %v", err)
				return
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				if _, err := w.WriteChunkedBodyDone(); err != nil {
					log.Printf("Error, could not write chunked body done to response: %v", err)
					return
				}
				log.Print("Finished Streaming")
				break
			}
			log.Printf("Error: could not read from response body: %v", err)
			return
		}
	}

	checksum := sha256.Sum256(fullBody)

	trailers := headers.NewHeaders()
	trailers.Set("X-Content-SHA256", hex.EncodeToString(checksum[:]))
	trailers.Set("X-Content-Length", strconv.Itoa(len(fullBody)))
	if err := w.WriteTrailers(trailers); err != nil {
		log.Printf("Error: could not writer trailers: %v", err)
	}
}
