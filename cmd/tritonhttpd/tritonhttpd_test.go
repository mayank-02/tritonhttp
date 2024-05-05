package main

import (
	"bufio"
	"bytes"
	"cse224/tritonhttp"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	HTTP1_1                = "HTTP/1.1"
	ErrSendingRequest      = "Error sending request"
	ErrParsingResponse     = "Error parsing response"
	ErrStatusMsg           = "Status code mismatch"
	ErrProtocolMsg         = "Protocol mismatch"
	ErrConnectionHeaderMsg = "Connection header mismatch"
)

var usehttpd = flag.String("usehttpd", "tritonhttp", "Which httpd server to use? ('tritonhttp' or 'go')")

func findhtdocs(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting cwd: %v\n", err.Error())
	}
	basedir := path.Dir(path.Dir(cwd))
	htdocsdir := path.Join(basedir, "tests", "htdocs1")

	return htdocsdir
}

func launchhttpd(t *testing.T) {
	switch *usehttpd {
	case "tritonhttp":
		launchtritonhttpd(t)
	case "go":
		launchgohttpd(t)
	default:
		t.Fatalf("Invalid server type %v (must be 'tritonhttp' or 'go')", *usehttpd)
	}
}

func launchgohttpd(t *testing.T) {
	htdocs := findhtdocs(t)
	s := &http.Server{
		Addr:    ":8080",
		Handler: http.FileServer(http.Dir(htdocs)),
	}
	go s.ListenAndServe()
}

func launchtritonhttpd(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(cwd)
	t.Log(cwd)
	virtualHosts := tritonhttp.ParseVHConfigFile("../../virtual_hosts.yaml", "../../docroot_dirs")
	s := &tritonhttp.Server{
		Addr:         ":8080",
		VirtualHosts: virtualHosts,
	}
	go s.ListenAndServe()
}

func TestGoFetch1(t *testing.T) {
	launchhttpd(t)

	req := "GET / HTTP/1.1\r\n" +
		"Host: website1\r\n" +
		"Connection: close\r\n" +
		"User-Agent: gotest\r\n" +
		"\r\n"

	respbytes, _, err := tritonhttp.Fetch("localhost", "8080", []byte(req))
	require.NoError(t, err, ErrSendingRequest)

	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(respbytes)), nil)
	require.NoError(t, err, ErrParsingResponse)

	assert.Equal(t, HTTP1_1, resp.Proto)
	assert.Equal(t, 200, resp.StatusCode, ErrStatusMsg)
	assert.Equal(t, true, resp.Close, ErrConnectionHeaderMsg)
	assert.NotEmpty(t, resp.Header.Get("Date"))
	assert.NotEmpty(t, resp.Header.Get("Last-Modified"))
	assert.Equal(t, int64(377), resp.ContentLength, "Content-Length mismatch")
	assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))

	resp.Body.Close()
}

func TestGoFetch2(t *testing.T) {
	launchhttpd(t)

	req := "GET / HTTP/1.1\r\n" +
		"Host: website1\r\n" +
		"User-Agent: gotest\r\n" +
		"\r\n" +
		"GET /notfound.html HTTP/1.1\r\n" +
		"Host: website1\r\n" +
		"User-Agent: gotest\r\n" +
		"Connection: close\r\n" +
		"\r\n"

	respbytes, _, err := tritonhttp.Fetch("localhost", "8080", []byte(req))
	require.NoError(t, err, ErrSendingRequest)

	respreader := bufio.NewReader(bytes.NewReader(respbytes))

	// Response 1
	resp, err := http.ReadResponse(respreader, nil)
	require.NoError(t, err, "Error parsing the response 1")

	assert.Equal(t, HTTP1_1, resp.Proto, ErrProtocolMsg)
	assert.Equal(t, 200, resp.StatusCode, ErrStatusMsg)
	assert.Equal(t, int64(377), resp.ContentLength, "Content-Length mismatch")

	indexbytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Error reading response body 1")
	assert.Equal(t, int(resp.ContentLength), len(indexbytes), "Response body length mismatch 1")
	resp.Body.Close()

	// Response 2
	resp, err = http.ReadResponse(respreader, nil)
	require.NoError(t, err, "Error parsing the response 2")

	assert.Equal(t, HTTP1_1, resp.Proto, ErrProtocolMsg)
	assert.Equal(t, true, resp.Close, ErrConnectionHeaderMsg)
	assert.Equal(t, 404, resp.StatusCode, ErrStatusMsg)
	resp.Body.Close()
}

func TestGoFetch3(t *testing.T) {
	launchhttpd(t)

	req := "foobar\r\n" +
		"Host: website1\r\n" +
		"Connection: close\r\n" +
		"User-Agent: gotest\r\n" +
		"\r\n"

	respbytes, _, err := tritonhttp.Fetch("localhost", "8080", []byte(req))
	require.NoError(t, err, ErrSendingRequest)

	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(respbytes)), nil)
	require.NoError(t, err, ErrParsingResponse)

	assert.Equal(t, HTTP1_1, resp.Proto, ErrProtocolMsg)
	assert.Equal(t, 400, resp.StatusCode, ErrStatusMsg)
	assert.Equal(t, true, resp.Close, ErrConnectionHeaderMsg)
	assert.NotEmpty(t, resp.Header.Get("Date"))
	resp.Body.Close()
}

func TestConcurrentRequests(t *testing.T) {
	launchhttpd(t)

	var wg sync.WaitGroup
	requestFunc := func() {
		defer wg.Done()
		req := "GET / HTTP/1.1\r\nHost: website1\r\n\r\n"
		respbytes, _, err := tritonhttp.Fetch("localhost", "8080", []byte(req))
		assert.NoError(t, err, ErrSendingRequest)
		resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(respbytes)), nil)
		assert.NoError(t, err, ErrParsingResponse)
		assert.Equal(t, "200 OK", resp.Status, ErrStatusMsg)
		resp.Body.Close()
	}

	concurrentUsers := 10
	wg.Add(concurrentUsers)
	for i := 0; i < concurrentUsers; i++ {
		go requestFunc()
	}
	wg.Wait()
}

func TestSuccessfulResponseHeaders(t *testing.T) {
	launchhttpd(t)

	req := "GET / HTTP/1.1\r\n" +
		"Host: website3\r\n" +
		"Connection: close\r\n" +
		"User-Agent: gotest\r\n" +
		"\r\n"

	respbytes, _, err := tritonhttp.Fetch("localhost", "8080", []byte(req))
	require.NoError(t, err, ErrSendingRequest)

	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(respbytes)), nil)
	require.NoError(t, err, ErrParsingResponse)

	assert.NotEmpty(t, resp.Header.Get("Date"), "Date header is missing")
	assert.NotEmpty(t, resp.Header.Get("Last-Modified"), "Last-Modified header is missing")
	assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"), "Content-Type header mismatch")
	assert.NotEmpty(t, resp.Header.Get("Content-Length"), "Content-Length header is missing")
	assert.Equal(t, "200 OK", resp.Status, ErrStatusMsg)

	resp.Body.Close()
}

func TestRealPathTraversalAttack(t *testing.T) {
	launchhttpd(t)

	req := "GET /../htdocs2/ HTTP/1.1\r\n" +
		"Host: website1\r\n" +
		"\r\n"

	respbytes, _, err := tritonhttp.Fetch("localhost", "8080", []byte(req))
	require.NoError(t, err, ErrSendingRequest)

	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(respbytes)), nil)
	require.NoError(t, err, ErrParsingResponse)

	assert.Equal(t, "404 Not Found", resp.Status, ErrStatusMsg)

	resp.Body.Close()
}

func TestFakePathTraversalAttack(t *testing.T) {
	launchhttpd(t)

	req := "GET /../htdocs2/../htdocs1/kitten.jpg HTTP/1.1\r\n" +
		"Host: website1\r\n" +
		"\r\n"

	respbytes, _, err := tritonhttp.Fetch("localhost", "8080", []byte(req))
	require.NoError(t, err, ErrSendingRequest)

	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(respbytes)), nil)
	require.NoError(t, err, ErrParsingResponse)

	assert.Equal(t, "200 OK", resp.Status, ErrStatusMsg)

	resp.Body.Close()
}

func TestEmptyRequest(t *testing.T) {
	launchhttpd(t)

	req := ""

	respbytes, _, err := tritonhttp.Fetch("localhost", "8080", []byte(req))
	require.NoError(t, err, ErrSendingRequest)

	_, err = http.ReadResponse(bufio.NewReader(bytes.NewReader(respbytes)), nil)
	assert.ErrorIs(t, err, io.ErrUnexpectedEOF, "Expected an error parsing the response")
}

func TestPartialRequest(t *testing.T) {
	launchhttpd(t)

	req := "GET / HTTP/1.1\r\n"

	respbytes, _, err := tritonhttp.Fetch("localhost", "8080", []byte(req))
	require.NoError(t, err, ErrSendingRequest)

	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(respbytes)), nil)
	require.NoError(t, err, ErrParsingResponse)
	t.Log(resp.Header)

	assert.Equal(t, "400 Bad Request", resp.Status, ErrStatusMsg)
	assert.NotEmpty(t, resp.Header.Get("Date"), "Date header missing")
	assert.Equal(t, true, resp.Close, ErrConnectionHeaderMsg)
}

func TestSlowButCompleteHTTPRequest(t *testing.T) {
	launchhttpd(t)

	conn, err := net.Dial("tcp", "localhost:8080")
	require.NoError(t, err, "Failed to dial the server")
	defer conn.Close()

	// Send complete request in fragments
	fragments := []string{
		"GET / HTTP/1.1\r\nHost:", // 0th second
		"web",                     // 3rd second
		"site",                    // 6th second
		"1\r\n\r\n",               // 9th second
	}

	for _, frag := range fragments {
		_, err := fmt.Fprint(conn, frag)
		require.NoError(t, err, "Failed to send fragment")

		time.Sleep(3 * time.Second)
	}

	response, err := bufio.NewReader(conn).ReadString('\n')
	require.NoError(t, err, "Failed to read response")

	assert.Contains(t, response, "HTTP/1.1 200 OK", "Expected HTTP/1.1 200 OK response")
}

func TestSlowAndIncompleteHTTPRequest(t *testing.T) {
	launchhttpd(t)

	conn, err := net.Dial("tcp", "localhost:8080")
	require.NoError(t, err, "Failed to dial the server")
	defer conn.Close()

	// Sending an incomplete request in fragments
	fragments := []string{
		"GET / HTTP/1.1\r\nHost:", // 0th second
		"web",                     // 3rd second
		"site",                    // 6th second
	}

	for _, frag := range fragments {
		_, err := fmt.Fprint(conn, frag)
		require.NoError(t, err, "Failed to send fragment")

		time.Sleep(3 * time.Second)
	}

	response, err := bufio.NewReader(conn).ReadString('\n')
	require.NoError(t, err, "Failed to read response")

	assert.Contains(t, response, "HTTP/1.1 400 Bad Request", "Expected HTTP/1.1 400 Bad Request response")
}

func TestHTTPRequestParsing(t *testing.T) {
	launchhttpd(t)

	tests := []struct {
		name           string
		request        string
		expectedStatus int
	}{
		{
			name: "Missing File",
			request: "GET /missing.html HTTP/1.1\r\n" +
				"Host: website1\r\n\r\n",
			expectedStatus: 404,
		},
		{
			name:           "Missing Host Header",
			request:        "GET /index.html HTTP/1.1\r\n\r\n",
			expectedStatus: 400,
		},
		{
			name: "Header Key with Leading White Space",
			request: "GET /index.html HTTP/1.1\r\n" +
				" Host: website1\r\n\r\n",
			expectedStatus: 400,
		},
		{
			name: "Incorrect Capitalization of Connection Header Value",
			request: "GET /index.html HTTP/1.1\r\n" +
				"Host: website1\r\n" +
				"Connection: ClOSe\r\n\r\n",
			expectedStatus: 200,
		},
		{
			name: "Header Key with Non-Alphanumeric Character",
			request: "GET /index.html HTTP/1.1\r\n" +
				"Host: website1\r\n" +
				"User-%Agent: gotest\r\n\r\n",
			expectedStatus: 400,
		},
		{
			name: "Header Value with Leading White Space",
			request: "GET /index.html HTTP/1.1\r\n" +
				"Host:                                   website3\r\n\r\n",
			expectedStatus: 200,
		},
		{
			name: "Header Value with Trailing White Space",
			request: "GET /index.html HTTP/1.1\r\n" +
				"Host:website3                       \r\n\r\n",
			expectedStatus: 404,
		},
		{
			name: "Mixed Case HTTP Header",
			request: "GET /index.html HTTP/1.1\r\n" +
				"HoSt: website2\r\n" +
				"Connection: close\r\n\r\n",
			expectedStatus: 200,
		},
		{
			name: "Unsupported HTTP Method",
			request: "PATCH /index.html HTTP/1.1\r\n" +
				"Host: website2\r\n\r\n",
			expectedStatus: 400,
		},
		{
			name: "Unsupported HTTP Version",
			request: "GET /index.html HTTP/2.0\r\n" +
				"Host: website2\r\n\r\n",
			expectedStatus: 400,
		},
		{
			name: "No Space After Colon In HTTP Header",
			request: "GET /index.html HTTP/1.1\r\n" +
				"Host:website2\r\n" +
				"User-agent:gotest\r\n\r\n",
			expectedStatus: 200,
		},
		{
			name: "Missing URL",
			request: "GET HTTP/1.1\r\n" +
				"Host:website2\r\n" +
				"User-agent:gotest\r\n\r\n",
			expectedStatus: 400,
		},
		{
			name: "Fetching Directory Instead of File",
			request: "GET /subdir HTTP/1.1\r\n" +
				"Host:website2\r\n" +
				"User-agent:gotest\r\n\r\n",
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			respbytes, _, err := tritonhttp.Fetch("localhost", "8080", []byte(tt.request))
			require.NoError(t, err, "%s for test: %s", ErrSendingRequest, tt.name)

			resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(respbytes)), nil)
			require.NoError(t, err, "%s for test: %s", ErrParsingResponse, tt.name)

			assert.Equal(t, HTTP1_1, resp.Proto)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Test %s: %s", tt.name, ErrStatusMsg)
			assert.NotEmpty(t, resp.Header.Get("Date"), "Date header missing")
			if resp.StatusCode == 400 {
				assert.Equal(t, true, resp.Close, "Test %s: %s", tt.name, ErrConnectionHeaderMsg)
			}
			if resp.StatusCode == 200 {
				assert.NotEmpty(t, resp.Header.Get("Content-Length"), "Test %s: Content-Length header missing", tt.name)
				assert.NotEmpty(t, resp.Header.Get("Content-Type"), "Test %s: Content-Type header missing", tt.name)
				assert.NotEmpty(t, resp.Header.Get("Last-Modified"), "Test %s: Last-Modified header missing", tt.name)
			}
			resp.Body.Close()
		})
	}
}

func TestAllFilesInHtdocs(t *testing.T) {
	launchhttpd(t)

	virtualHosts := tritonhttp.ParseVHConfigFile("../../virtual_hosts.yaml", "../../docroot_dirs")

	for hostname, docRoot := range virtualHosts {
		err := filepath.Walk(docRoot, func(path string, info os.FileInfo, err error) error {
			require.NoError(t, err, "Error walking through docRoot")

			if info.IsDir() {
				return nil
			}

			testfile := strings.TrimPrefix(path, docRoot)
			t.Run(testfile, func(t *testing.T) {
				req := fmt.Sprintf("GET %s HTTP/1.1\r\n"+
					"Host: "+hostname+"\r\n"+
					"Connection: close\r\n"+
					"User-Agent: gotest\r\n"+
					"\r\n", testfile)

				respbytes, _, err := tritonhttp.Fetch("localhost", "8080", []byte(req))
				require.NoError(t, err, ErrSendingRequest)

				resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(respbytes)), nil)
				require.NoError(t, err, ErrParsingResponse)

				assert.Equal(t, HTTP1_1, resp.Proto, ErrProtocolMsg)
				assert.Equal(t, 200, resp.StatusCode, ErrStatusMsg)

				respcontenttype := resp.Header.Get("Content-Type")
				assert.NotEmpty(t, respcontenttype, "Content-Type header is missing")

				origmimetype := mime.TypeByExtension(filepath.Ext(path))
				assert.Contains(t, origmimetype, respcontenttype, "Content-Type mismatch")

				assert.Equal(t, info.Size(), resp.ContentLength, "Content-Length mismatch")

				origcontents, err := os.ReadFile(path)
				require.NoError(t, err, "Error reading input file")

				respcontents, err := io.ReadAll(resp.Body)
				require.NoError(t, err, "Error reading response body")
				resp.Body.Close()

				assert.True(t, bytes.Equal(origcontents, respcontents), "Response body does not match original file contents")
			})

			return nil
		})
		require.NoError(t, err, "Error walking the file tree")
	}
}
