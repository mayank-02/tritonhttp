package tritonhttp

import (
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Response struct {
	Proto      string // e.g. "HTTP/1.1"
	StatusCode int    // e.g. 200
	StatusText string // e.g. "OK"

	// Headers stores all headers to write to the response.
	Headers map[string]string

	// Request is the valid request that leads to this response.
	// It could be nil for responses not resulting from a valid request.
	// Hint: you might need this to handle the "Connection: Close" requirement
	Request *Request

	// FilePath is the local path to the file to serve.
	// It could be "", which means there is no file to serve.
	FilePath string
}

// NewResponse create new instance of Response with the given request and status code.
func NewResponse(s *Server, request *Request, statusCode int) Response {
	r := Response{
		Proto:      "HTTP/1.1",
		StatusCode: statusCode,
		StatusText: StatusCodeText[statusCode],
		Headers:    make(map[string]string),
		Request:    request,
		FilePath:   "",
	}
	if statusCode == 200 {
		r.FilePath = filepath.Clean(s.VirtualHosts[request.Host] + request.URL)
		if !strings.HasPrefix(r.FilePath, s.VirtualHosts[request.Host]) {
			log.Printf("Trying to access file: %v outside document root: %v", r.FilePath, s.VirtualHosts[request.Host])
			r.StatusCode = 404
			r.StatusText = StatusCodeText[404]
			r.FilePath = ""
			return r
		}

		fileinfo, err := os.Stat(r.FilePath)
		if err != nil {
			log.Printf("Error getting file info: %v", err)
			r.StatusCode = 404
			r.StatusText = StatusCodeText[404]
			r.FilePath = ""
			return r
		}

		r.Headers["Content-Length"] = fmt.Sprintf("%v", fileinfo.Size())
		r.Headers["Content-Type"] = mime.TypeByExtension(filepath.Ext(r.FilePath))
		r.Headers["Date"] = FormatTime(time.Now())
		r.Headers["Last-Modified"] = FormatTime(fileinfo.ModTime())
		if request.Close {
			r.Headers["Connection"] = "close"
		}
	}
	return r
}

func (res *Response) Write(w io.Writer) error {
	// Write status line
	statusLine := fmt.Sprintf("%v %v %v\r\n", res.Proto, res.StatusCode, StatusCodeText[res.StatusCode])
	if _, err := fmt.Fprint(w, statusLine); err != nil {
		return err
	}

	// Write headers sorted by keys
	keys := make([]string, 0, len(res.Headers))
	for key := range res.Headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := res.Headers[key]
		headerLine := fmt.Sprintf("%v: %v\r\n", key, value)
		if _, err := fmt.Fprint(w, headerLine); err != nil {
			return err
		}
	}

	// Write a blank line to separate headers and body
	if _, err := fmt.Fprintf(w, "\r\n"); err != nil {
		return err
	}

	// Write body if there is any
	if res.FilePath == "" {
		return nil
	}

	// Open file to serve
	file, err := os.Open(res.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write file contents directly to the provided writer
	if _, err := io.Copy(w, file); err != nil {
		return err
	}

	return nil
}
