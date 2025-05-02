package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gppmad/gonc/network"
)

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  Client mode (default): gonc [options] HOST:PORT")
	fmt.Println("  Server mode: gonc -l [options] PORT")
	fmt.Println("\nOptions:")
	fmt.Println("  -tls          Use TLS for the connection")
	fmt.Println("  -l            Listen mode (server)")
	fmt.Println("  -h            Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  gonc example.com:8080     Connect to example.com on port 8080")
	fmt.Println("  gonc -tls example.com:443 Connect to example.com on port 443 using TLS")
	fmt.Println("  gonc -l 8080              Listen on port 8080")
	fmt.Println("  gonc -l -tls 443          Listen on port 443 using TLS")
}

func validateArgs(serverMode bool, args []string) bool {
	if len(args) != 1 {
		if serverMode {
			fmt.Println("Error: Server mode requires a PORT argument")
		} else {
			fmt.Println("Error: Client mode requires a HOST:PORT argument")
		}
		return false
	}

	if serverMode {
		// Server mode: validate port only
		port := args[0]

		// Check if port is numeric
		for _, c := range port {
			if c < '0' || c > '9' {
				fmt.Println("Error: PORT must be numeric")
				return false
			}
		}

		// Could add port range validation if needed
		portNum := 0
		fmt.Sscanf(port, "%d", &portNum)
		if portNum <= 0 || portNum > 65535 {
			fmt.Println("Error: PORT must be between 1 and 65535")
			return false
		}
	} else {
		// Client mode: validate host:port format
		remoteAddr := args[0]

		// Check for presence of ":" separator
		parts := strings.Split(remoteAddr, ":")
		if len(parts) != 2 {
			fmt.Println("Error: Client mode requires HOST:PORT format")
			return false
		}

		host, port := parts[0], parts[1]

		// Validate host
		if host == "" {
			fmt.Println("Error: HOST cannot be empty")
			return false
		}

		// Validate port (must be numeric)
		for _, c := range port {
			if c < '0' || c > '9' {
				fmt.Println("Error: PORT must be numeric")
				return false
			}
		}

		// Could add port range validation if needed
		portNum := 0
		fmt.Sscanf(port, "%d", &portNum)
		if portNum <= 0 || portNum > 65535 {
			fmt.Println("Error: PORT must be between 1 and 65535")
			return false
		}
	}

	return true
}

func runClient(host, port string, requireTLS bool) error {
	address := fmt.Sprintf("%s:%s", host, port)

	config := network.ClientConfig{
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
	fmt.Printf("Starting server on port %s (TLS: %v)\n", port, requireTLS)
	return fmt.Errorf("not implemented yet")
}

func main() {

	// Get the flags and parse them
	requireTLS := flag.Bool("tls", false, "Use TLS for the connection")
	serverMode := flag.Bool("l", false, "Listen mode - start server instead of client")
	helpFlag := flag.Bool("h", false, "Show help")

	flag.Parse()

	// Show help if requested or if no arguments provided
	if *helpFlag {
		printUsage()
		os.Exit(0) // Exit with success code when showing help
	}

	// Get the args
	args := flag.Args()

	// Validate arguments based on mode
	if !validateArgs(*serverMode, args) {
		printUsage()
		os.Exit(1)
	}

	// Run in appropriate mode
	var err error
	if *serverMode {
		err = runServer(args[0], *requireTLS)

	} else {
		// host and port are provided with this syntax host:port
		parts := strings.Split(args[0], ":")
		err = runClient(parts[0], parts[1], *requireTLS)
	}
	if err != nil {
		log.Fatal(err)
	}

}
