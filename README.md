# httpfromtcp

A boot.dev guided project in Go that implements HTTP directly on top of raw TCP.
It provides custom request parsing, header handling, response writing,
and streaming support without relying on Go’s built-in net/http server.

The project demonstrates:

- Manual HTTP request parsing over TCP
- Custom response writer and header abstraction
- Status code handling and routing
- Chunked transfer encoding with trailers
- Streaming responses from an upstream server

## Running the Server

Build and run the server (adjust the command if your entry point differs):

```
go build
./httpfromtcp
```

By default, the server listens on the configured TCP port.

## Testing with curl

### Basic Requests

`curl -v http://localhost:8080/`

Should receive http response with `200 OK` status line.

`curl -v http://localhost:8080/yourproblem`

Should receive http response with `400 Bad Request` status line.

`curl -v http://localhost:8080/myproblem`

Should receive http response with `500 Internal Server Error` status line.

### Chunked Streaming + Trailers

A chunked streaming handler proxies requests to [httpbin](https://httpbin.org/)
and streams the response back to the requester.

`curl -v --raw http://localhost:8080/httpbin/stream/5`

To explicitly show trailers:

`curl -v --raw --trailers http://localhost:8080/httpbin/get`

You should see `Transfer-Encoding: Chunked` in the response headers
and trailer fields printed at the end of the response.

### Static Binary Content (Video Serving)

The videoHandler reads an MP4 file from disk and returns it with the appropriate
Content-Type, allowing standard HTTP clients to play or download the video
without any special handling.

As a natural follow-on from chunked-encoding, the server can now serve large,
non-text files. It streams the binary data, allowing video transfer.

`curl -v http://localhost:8080/video --output vim.mp4`

The downloaded file can be played normally in any media player, confirming that
the response framing, headers, and body handling are all correct.

Alternatively, navigate to `http://localhost:8080/video` in your browser
while the server is running.

## Things I Learned

- **HTTP is just a protocol on top of TCP**
Building everything from scratch made it clear how much convenience Go’s
net/http normally provides, and how explicit you must be about formatting,
ordering, and correctness when working directly over TCP.

- **Request parsing is deceptively tricky**
Correctly handling request lines, headers, and edge cases (like malformed input)
requires strict adherence to the HTTP/1.1 spec and extensive testing.

- **Response composition matters**
Status lines, headers, and bodies must be written in the correct order,
and small mistakes (like incorrect Content-Length) can completely break clients.

- **Chunked transfer encoding has real-world complexity**
Implementing chunked responses highlighted how streaming works in practice,
including chunk framing, termination, and how trailers are sent after the body.

- **Trailers are rarely used but very powerful**
Trailers allow metadata (like checksums or computed lengths) to be sent after
streaming completes—something that is not possible with traditional
content-length responses.

- **Streaming requires careful error handling**
Reading from an upstream response and writing chunks downstream requires handling
partial reads, EOFs, and write failures without corrupting the response.

- **HTTP clients are strict (for good reason)**
Tools like curl are excellent for validating correctness and quickly reveal
protocol violations that might otherwise go unnoticed.
