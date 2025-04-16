package tls_client

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"os"
)

var tlsDial = tls.Dial

type TlsClient struct {
	Input  io.Reader
	Output io.Writer
	Conn   net.Conn
}

func NewTlsClient(conn net.Conn, input io.Reader, output io.Writer) *TlsClient {
	if input == nil {
		input = os.Stdin
	}

	if output == nil {
		output = os.Stdout
	}

	return &TlsClient{Conn: conn, Input: input, Output: output}
}

// Before call this server initialize the connection with Connect()
func (c *TlsClient) Start() error {
	if c.Conn == nil {
		return errors.New("connect to the target before initialize a new connection")
	}

	errChan := make(chan error, 1)

	// Read from the connection.
	go func() {
		_, err := io.Copy(c.Output, c.Conn)
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	_, err := io.Copy(c.Conn, c.Input)
	if err != nil {
		return errors.New("error writing in the connection")
	}

	// Check errors from during connection reading.
	if err := <-errChan; err != nil {
		return errors.New("error reading from the connection: " + err.Error())
	}
	return nil
}

// Close the connection
func (c *TlsClient) Close() error {
	return c.Conn.Close()
}

// Helper function to establish a TLS connection
func Connect(address string, config *tls.Config) (*tls.Conn, error) {
	if config == nil {
		config = &tls.Config{
			InsecureSkipVerify: false,
		}
	}

	// Infer ServerName from address if not set
	if config.ServerName == "" {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return nil, errors.New("invalid address format")
		}
		config.ServerName = host
	}

	return tlsDial("tcp", address, config)
}
