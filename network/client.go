package network

import (
	"net"

	"github.com/gppmad/gonc/tcp_client"
)

// Client defines the common operations for all network clients
type Client interface {
	// Start initiates the connection and handles I/O
	Start() error

	// Close terminates the connection
	Close() error
}

// Config contains connection parameters
type Config struct {
	RemoteAddr string
}

// NewClient creates a new network client based on config
func NewClient(config Config) (Client, error) {
	// Connect to remote server
	conn, err := net.Dial("tcp", config.RemoteAddr)
	if err != nil {
		return nil, err
	}

	// Create and return TCP client
	return tcp_client.NewTcpClient(conn, nil, nil), nil
}
