package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Check for the correct number of arguments
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage:\t%s docRoot port\n", os.Args[0])
		os.Exit(1)
	}

	// Parse the command line arguments
	docroot := os.Args[1]
	port := os.Args[2]

	log.Printf("Using docRoot: %v", docroot)
	log.Printf("Using port: %v", port)

	// Create a new server and start listening
	s := &http.Server{
		Addr:    ":" + port,
		Handler: http.FileServer(http.Dir(docroot)),
	}
	log.Fatal(s.ListenAndServe())
}
