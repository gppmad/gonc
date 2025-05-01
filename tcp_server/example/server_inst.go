package example

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gppmad/gonc/tcp_server"
)

func example() {

	// Create a TCP listener on port 8080
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}

	tcp_server := tcp_server.NewTcpServer(listener, nil, nil)
	// Setup signal handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start signal handler goroutine
	go func() {
		sig := <-signalChan
		fmt.Printf("Received signal: %v\n", sig)
		if err := tcp_server.Close(); err != nil {
			log.Printf("Error closing server: %v", err)
		}
		fmt.Println("Server shut down gracefully")
		os.Exit(0)
	}()

	// Run server in main thread
	fmt.Println("Server started on :8080")
	if err := tcp_server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
