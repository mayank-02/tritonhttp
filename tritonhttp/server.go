package tritonhttp

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
)

const (
	StatusOK         = 200
	StatusBadRequest = 400
	StatusNotFound   = 404
	TCP              = "tcp"
)

var StatusCodeText = map[int]string{
	StatusOK:         "OK",
	StatusBadRequest: "Bad Request",
	StatusNotFound:   "Not Found",
}

type Server struct {
	// Addr specifies the TCP address for the server to listen on,
	// in the form "host:port". It shall be passed to net.Listen()
	// during ListenAndServe().
	Addr string // e.g. ":0"

	// VirtualHosts contains a mapping from host name to the docRoot path
	// (i.e. the path to the directory to serve static files from) for
	// all virtual hosts that this server supports
	VirtualHosts map[string]string
}

// ListenAndServe listens on the TCP network address s.Addr and then
// handles requests on incoming connections.
func (s *Server) ListenAndServe() error {

	// Validate all docRoots
	for _, docRoot := range s.VirtualHosts {
		// Shortest path name equivalent to path by *purely lexical processing*
		docrootPath := filepath.Clean(docRoot)

		// Check if the path exists
		fileInfo, err := os.Stat(docRoot)
		if err != nil {
			log.Fatalf("Docroot %s does not exist: %v", docRoot, err)
		}

		// Check if the path is a directory
		if !fileInfo.IsDir() {
			log.Fatalf("Docroot %s is not a directory", docrootPath)
		}
	}

	// Server listens on the configured address
	ln, err := net.Listen(TCP, s.Addr)
	if err != nil {
		return err
	}

	// Ensure the listener is closed on exit
	defer func() {
		err = ln.Close()
		if err != nil {
			log.Println("Failed to close listener", err)
		}
	}()
	log.Println("Listening on", ln.Addr())

	// Continuously accept new connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Failed to accept connection", err)
			continue
		}

		log.Println("Accepted connection from", conn.RemoteAddr())

		// Handle the connection in a new goroutine
		go s.HandleConnection(conn)
	}
}

// HandleConnection reads requests from the accepted conn and handles them.
func (s *Server) HandleConnection(conn net.Conn) {
	br := bufio.NewReader(conn)

	// Continuously read from  connection until EOF or timeout
	for {
		// Read next request from the client
		req, bytesRead, err := ReadRequest(conn, br)

		// Handle EOF
		if errors.Is(err, io.EOF) {
			log.Printf("Connection closed by %v", conn.RemoteAddr())
			conn.Close()
			return
		}

		// Handle Timeout
		if err, ok := err.(net.Error); ok && err.Timeout() {
			log.Printf("Connection to %v timed out after reading %v bytes", conn.RemoteAddr(), bytesRead)
			if bytesRead > 0 {
				res := NewResponse(s, req, StatusBadRequest)
				res.Write(conn)
			}
			log.Printf("Closing connection to %v", conn.RemoteAddr())
			conn.Close()
			return
		}

		// Handle the request which is not a GET and immediately close the connection and return
		if err != nil {
			log.Printf("Handle bad request for error: %v", err)
			res := NewResponse(s, req, StatusBadRequest)
			res.Write(conn)
			log.Printf("Closing connection to %v", conn.RemoteAddr())
			conn.Close()
			return
		}

		res := NewResponse(s, req, StatusOK)
		err = res.Write(conn)
		if err != nil {
			log.Println(err)
		}

		if req.Close {
			log.Printf("Closing connection to %v", conn.RemoteAddr())
			conn.Close()
			return
		}
		// We'll never close the connection and handle as many requests for this connection
		// and pass on this responsibility to the timeout mechanism
	}
}
