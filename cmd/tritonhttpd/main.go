package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"cse224/tritonhttp"
)

func main() {
	currDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Could not get current working directory: %v", err)
	}
	default_vh_config_path := filepath.Join(currDir, "virtual_hosts.yaml")
	default_docroot := filepath.Join(currDir, "docroot_dirs")

	// Parse command line flags
	var port = flag.Int("port", 8080, "the localhost port to listen on")
	var vh_config_path = flag.String("vh_config", default_vh_config_path, "path to the virtual hosting config file")
	var docroot_dirs_path = flag.String("docroot", default_docroot, "path to the directory that contains all docroot dirs")
	flag.Parse()

	// Log server configs
	fmt.Println()
	log.Print("Server configs:")
	log.Printf("  port: %v", *port)
	log.Printf("  path to virtual hosts config file: %v", *vh_config_path)
	log.Printf("  path to docroot directories: %v", *docroot_dirs_path)
	fmt.Println()

	virtualHosts := tritonhttp.ParseVHConfigFile(*vh_config_path, *docroot_dirs_path)

	// Start server
	addr := fmt.Sprintf(":%v", *port)

	log.Printf("Starting TritonHTTP server")
	log.Printf("You can browse the website at http://localhost:%v/", *port)
	s := &tritonhttp.Server{
		Addr:         addr,
		VirtualHosts: virtualHosts,
	}
	log.Fatal(s.ListenAndServe())
}
