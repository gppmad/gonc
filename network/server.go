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
	IP         string
	Port       string
	RequireTLS bool
}

// NewServer creates a new network server based on config
func NewServer(config ServerConfig) (Server, error) {
	// Construct the full address with IP and port
	address := net.JoinHostPort(config.IP, config.Port)

	// Create a standard TCP listener
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	// Create and return TCP server
	return tcp_server.NewTcpServer(listener, os.Stdin, os.Stdout), nil
}
