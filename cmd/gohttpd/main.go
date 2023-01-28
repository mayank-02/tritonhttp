package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\t%s doc_root port\n", os.Args[0])
	os.Exit(1)
}

func main() {

	if len(os.Args) != 3 {
		usage()
	}

	docroot := os.Args[1]
	port := os.Args[2]

	log.Printf("Using doc_root: %v", docroot)
	log.Printf("Using port: %v", port)

	s := &http.Server{
		Addr:    ":" + port,
		Handler: http.FileServer(http.Dir(docroot)),
	}
	log.Fatal(s.ListenAndServe())
}
