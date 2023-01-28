package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"cse224/tritonhttp"
)

// $ wwwfetch -req input.txt -resp response.dat hostname:port
// $ cat input.txt | wwwfetch -resp response.dat hostname:port
// $ cat input.txt | wwwfetch hostname:port > response.dat

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\t%s hostname:port\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	req := flag.String("req", "", "Request input file")
	resp := flag.String("resp", "", "Response output file")

	flag.Usage = usage
	flag.Parse()

	if len(flag.Args()) != 1 {
		usage()
	}

	if !strings.Contains(flag.Arg(0), ":") {
		fmt.Fprintf(os.Stderr, "command line must be of the form hostname:port\n")
		usage()
	}

	hostname := strings.Split(flag.Arg(0), ":")[0]
	port, err := strconv.Atoi(strings.Split(flag.Arg(0), ":")[1])
	if err != nil || port >= 65536 || port < 0 {
		fmt.Fprintf(os.Stderr, "Invalid port number: %s\n", strings.Split(flag.Arg(0), ":")[1])
		os.Exit(1)
	}

	// use the provided input file, or stdin if no file specified
	var inputFD *os.File
	if *req != "" {
		inputFD, err = os.Open(*req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file %s: %v\n", *req, err.Error())
			os.Exit(1)
		}
	} else {
		inputFD = os.Stdin
	}

	// Read the input from stdin or the given input file
	inp, err := io.ReadAll(bufio.NewReader(inputFD))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err.Error())
		os.Exit(1)
	}

	// fetch the web page
	respdata, duration, err := tritonhttp.Fetch(hostname, strconv.Itoa(port), inp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching web page: %v\n", err.Error())
		os.Exit(1)
	}

	// use the provided output file, or stdout if no file specified
	var outputFD *os.File
	if *resp != "" {
		outputFD, err = os.Create(*resp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening output file %s: %v\n", *resp, err.Error())
			os.Exit(1)
		}
		defer func() {
			err := outputFD.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error closing output file: %v\n", err)
				os.Exit(1)
			}
		}()
	} else {
		outputFD = os.Stdout
	}

	// Write the output to stdout or the given output file
	_, err = outputFD.Write(respdata)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err.Error())
		os.Exit(1)
	}

	log.Printf("Request contained %v bytes of data and took %v to retreive\n", len(respdata), duration)
}
