package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gppmad/gonc/network"
)

// Check args and exit the program in case they don't satisfied the input structure.
func checkArgs(listen *bool, args []string) {
	if *listen && len(args) < 1 {
		fmt.Println("Usage (server): gonc [options] -l port")
		flag.PrintDefaults()
		os.Exit(1)
	} else if !*listen && len(args) < 2 {
		fmt.Println("Usage (client): gonc [options] host port")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

// Run Client or Server mode
func run(listen *bool, args []string, requireTLS *bool) {
	var err error
	if *listen {
		err = runServer(args[0], *requireTLS)
	} else {
		err = runClient(args[0], args[1], *requireTLS)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func runClient(host, port string, requireTLS bool) error {
	address := fmt.Sprintf("%s:%s", host, port)

	config := network.Config{
		RemoteAddr: address,
		RequireTLS: requireTLS,
	}

	client, err := network.NewClient(config)
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	if requireTLS {
		fmt.Println("Connected to a TLS Server")
	} else {
		fmt.Println("Connected to a TCP Server")
	}

	if err := client.Start(); err != nil {
		return fmt.Errorf("error during starting the proxy connection: %w", err)
	}

	if err := client.Close(); err != nil {
		return fmt.Errorf("error during closing the proxy connection: %w", err)
	}

	return nil
}

func runServer(port string, requireTLS bool) error {
	// TODO: Implement server logic
	fmt.Printf("Starting server on port %s (TLS: %v)\n", port, requireTLS)
	return fmt.Errorf("server mode not implemented yet")
}

func main() {

	// Get the flags and parse them
	requireTLS := flag.Bool("tls", false, "Use TLS for the connection")
	listen := flag.Bool("l", false, "Listen mode - start server instead of client")

	flag.Parse()

	args := flag.Args()
	checkArgs(listen, flag.Args())

	run(listen, args, requireTLS)
}
