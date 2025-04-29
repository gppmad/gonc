package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gppmad/gonc/network"
)

func main() {

	requireTLS := flag.Bool("tls", false, "Use TLS for the connection")

	// Parse the flags
	flag.Parse()

	// Get host and port from arguments
	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("Usage: gonc [options] host port")
		flag.PrintDefaults()
		os.Exit(1)
	}

	host := args[0]
	port := args[1]

	address := fmt.Sprintf("%s:%s", host, port)

	// Create configuration
	config := network.Config{
		RemoteAddr: address,
		RequireTLS: *requireTLS,
	}

	// Create client based on configuration
	client, err := network.NewClient(config)
	if err != nil {
		log.Fatal("Error creating client: ", err)
	}

	// Update connection message based on protocol
	if *requireTLS {
		fmt.Println("Connected to a TLS Server")
	} else {
		fmt.Println("Connected to a TCP Server")
	}

	// Start the connection
	err = client.Start()
	if err != nil {
		// handle error
		log.Fatal("Error during starting the proxy connection: ", err)
	}

	// Close the connection
	err = client.Close()
	if err != nil {
		// handle error
		log.Fatal("Error during closing the proxy connection: ", err)
	}

}
