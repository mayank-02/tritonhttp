package tritonhttp

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"time"
)

// From a web page from hostname:port, returning the response as an array of bytes
// along with the duration for how long it took to retreive the page
func Fetch(hostname string, port string, inp []byte) ([]byte, time.Duration, error) {
	// connect to the server
	conn, err := net.DialTimeout("tcp", hostname+":"+port, CONNECT_TIMEOUT)
	if err != nil {
		log.Printf("Error connecting to web server: %v\n", err.Error())
		return nil, 0, errors.New("dial timeout")
	}

	// start timing this http session
	start := time.Now()

	// send the input to the server
	conn.SetWriteDeadline(time.Now().Add(SEND_TIMEOUT))
	_, err = conn.Write(inp)
	if err != nil {
		log.Printf("Error writing to web server: %v\n", err.Error())
		return nil, 0, errors.New("send error")
	}

	// read the input back from the server
	conn.SetReadDeadline(time.Now().Add(RECV_TIMEOUT))
	resp, err := io.ReadAll(bufio.NewReader(conn))
	if err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
		log.Printf("Error reading response from web server: %v\n", err.Error())
		return nil, 0, errors.New("recv error")
	}

	// stop timing this http session
	duration := time.Since(start)

	return resp, duration, nil
}
