package network

import (
	"net"
	"os"

	"github.com/gppmad/gonc/tcp_server"
)

// Server defines the common operations for all network servers
type Server interface {
	// Start begins accepting connections and handling them
	Start() error

	// Close stops the server and closes the listener
	Close() error
}

// ServerConfig contains server configuration parameters
type ServerConfig struct {
	RemoteAddr string
	RequireTLS bool
}

// NewServer creates a new network server based on config
func NewServer(config ClientConfig) (Server, error) {

	// Create a standard TCP listener
	listener, err := net.Listen("tcp", config.RemoteAddr)
	if err != nil {
		return nil, err
	}

	// Create and return TCP server
	return tcp_server.NewTcpServer(listener, os.Stdin, os.Stdout), nil
}
