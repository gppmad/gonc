package network

import (
	"crypto/tls"
	"net"
	"os"

	"github.com/gppmad/gonc/tcp_client"
	"github.com/gppmad/gonc/tls_client"
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
	RequireTLS bool
}

// NewClient creates a new network client based on config
func NewClient(config Config) (Client, error) {

	if config.RequireTLS {
		// Connect to remote server with a TLS connection.
		conn, err := tls_client.Connect(config.RemoteAddr, &tls.Config{})
		if err != nil {
			return nil, err
		}
		return tls_client.NewTlsClient(conn, os.Stdin, os.Stdout), nil
	} else {
		// Connect to remote server using a standard TCP connection
		conn, err := net.Dial("tcp", config.RemoteAddr)
		if err != nil {
			return nil, err
		}

		// Create and return TCP client
		return tcp_client.NewTcpClient(conn, os.Stdin, os.Stdout), nil
	}

}
