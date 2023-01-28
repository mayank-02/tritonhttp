package main

import (
	"net/http"
	"os"
	"path"
	"testing"
)

func findhtdocs(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting cwd: %v\n", err.Error())
	}
	basedir := path.Dir(path.Dir(cwd))
	htdocsdir := path.Join(basedir, "docroot_dirs", "htdocs1")

	return htdocsdir
}

func launchgohttpd(t *testing.T) *http.Server {
	htdocs := findhtdocs(t)
	s := &http.Server{
		Addr:    ":8080",
		Handler: http.FileServer(http.Dir(htdocs)),
	}
	t.Logf("Launching web server on http://localhost:8080/")
	go s.ListenAndServe()
	return s
}

func TestGet1(t *testing.T) {
	s := launchgohttpd(t)

	resp, err := http.Get("http://localhost:8080/index.html")
	if err != nil {
		t.Fatalf("Error issuing request: %v\n", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Expected status code of 200, but got %v instead\n", resp.StatusCode)
	}

	t.Log("Closing web server")
	s.Close()
}

func TestGet2(t *testing.T) {
	s := launchgohttpd(t)

	resp, err := http.Get("http://localhost:8080/cat.html")
	if err != nil {
		t.Fatalf("Error issuing request: %v\n", err.Error())
	}

	if resp.StatusCode != 404 {
		t.Fatalf("Expected status code of 404, but got %v instead\n", resp.StatusCode)
	}

	t.Log("Closing web server")
	s.Close()
}
