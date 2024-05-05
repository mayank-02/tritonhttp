package tritonhttp

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
	"unicode"
)

type Request struct {
	Method   string // e.g. "GET"
	URL      string // e.g. "/path/to/a/file"
	Protocol string // e.g. "HTTP/1.1"

	// Headers stores the key-value HTTP headers
	Headers map[string]string

	Host  string // determine from the "Host" header
	Close bool   // determine from the "Connection" header
}

func validHTTPMethod(method string) bool {
	return method == "GET"
}

func validHTTPVersion(version string) bool {
	return version == "HTTP/1.1"
}

// The key starts the line, followed by a colon and zero or more spaces, and then the value (each
// key-value pair is terminated by a CRLF delimiter). <key> is composed of one or more
// alphanumeric or the hyphen "-" character (i.e. <key> cannot be empty). It is case-insensitive.
// <value> can be any string not starting with space, and not containing CRLF. It is
// case-sensitive. As a special case <value> can be an empty string.
func validHTTPHeader(key string, value string) bool {
	if key == "" {
		return false
	}

	for _, c := range key {
		if !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '-' {
			return false
		}
	}

	return !strings.Contains(value, "\r\n")
}

// The URL specifies the location of the resource the client is interested in. Examples include
// /images/myimg.jpg and /classes/fall/cs101/index.html. A well-formed URL always starts with a /
// character. If the slash is missing, send back a 400 error.
func validURL(url string) bool {
	return strings.HasPrefix(url, "/")
}

// ReadRequest reads and parses an incoming request from br.
func ReadRequest(conn net.Conn, br *bufio.Reader) (request *Request, bytesRead int, err error) {
	bytesRead = 0
	line, err := ReadLine(conn, br)
	bytesRead += len(line)
	if err != nil {
		return nil, bytesRead, err
	}

	fields := strings.SplitN(line, " ", 3)
	if len(fields) != 3 {
		return nil, bytesRead, fmt.Errorf("invalid start line, got %v", line)
	}

	request = &Request{Method: fields[0], URL: fields[1], Protocol: fields[2], Headers: make(map[string]string)}

	// Read other lines of requests
	for {
		line, err := ReadLine(conn, br)
		bytesRead += len(line)
		if err != nil {
			return nil, bytesRead, err
		}

		if line == "" {
			// This marks header end
			break
		}

		key, value, err := parseHTTPHeader(line)
		if err != nil {
			return nil, bytesRead, err
		}

		if !validHTTPHeader(key, value) {
			return nil, bytesRead, fmt.Errorf("invalid HTTP header: %q", line)
		}

		request.Headers[key] = value
		if key == "Host" {
			request.Host = value
		} else if key == "Connection" {
			request.Close = value == "close"
		}
	}

	// HTTP version must be HTTP/1.1
	if !validHTTPVersion(request.Protocol) {
		return nil, bytesRead, fmt.Errorf("invalid HTTP version: %q", request.Protocol)
	}

	// HTTP method must be GET
	if !validHTTPMethod(request.Method) {
		return nil, bytesRead, fmt.Errorf("invalid HTTP method: %q", request.Method)
	}

	// Check the URL
	if !validURL(request.URL) {
		return nil, bytesRead, fmt.Errorf("invalid URL: %q", request.URL)
	}

	// Append index.html to the URL if it ends with a slash
	if strings.HasSuffix(request.URL, "/") {
		request.URL += "index.html"
	}

	// Host header must be present
	if request.Host == "" {
		return nil, bytesRead, fmt.Errorf("missing Host header")
	}

	return request, bytesRead, nil
}

func parseHTTPHeader(line string) (string, string, error) {
	// Split the line into key and value
	fields := strings.SplitN(line, ":", 2)

	// Missing colon
	if len(fields) != 2 {
		return "", "", fmt.Errorf("HTTP header missing colon: %q", line)
	}

	// Canonicalize the key
	key := CanonicalHeaderKey(fields[0])

	// Trim leading spaces from the value
	value := strings.TrimLeftFunc(fields[1], unicode.IsSpace)

	return key, value, nil
}

func ReadLine(conn net.Conn, br *bufio.Reader) (string, error) {
	var line []byte
	for {
		// Set timeout
		if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
			log.Printf("Failed to set timeout for connection %v", conn)
			conn.Close()
			return string(line), err
		}

		// Read one byte at a time and append to line
		s := make([]byte, 1)
		bytesRead, err := br.Read(s)
		if bytesRead > 0 {
			line = append(line, s[:bytesRead]...)
		}

		// If line ends in \r\n, return after trimming it
		if len(line) >= 2 && bytes.Equal(line[len(line)-2:], []byte("\r\n")) {
			return string(line[:len(line)-2]), nil
		}

		// Return the error with any data read so far
		if err != nil {
			return string(line), err
		}
	}
}
