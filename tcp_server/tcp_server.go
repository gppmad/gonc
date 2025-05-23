package tcp_server

import (
	"errors"
	"io"
	"net"
	"os"
)

// TcpServer represents a TCP server that can accept connections
type TcpServer struct {
	Listener net.Listener
	Input    io.Reader
	Output   io.Writer

	// Adding a new field to control server behavior.
	// This is the function called everytime the the listener accepts a connection.
	Handler func(conn net.Conn, input io.Reader, output io.Writer) error
}

// NewTcpServer creates a new TCP server instance with the specified components
func NewTcpServer(listener net.Listener, input io.Reader, output io.Writer) *TcpServer {
	if input == nil {
		input = os.Stdin
	}

	if output == nil {
		output = os.Stdout
	}

	return &TcpServer{
		Listener: listener,
		Input:    input,
		Output:   output,
		Handler:  DefaultHandler, // Set default handler
	}
}

// DefaultHandler is the standard connection handling logic
func DefaultHandler(conn net.Conn, input io.Reader, output io.Writer) error {
	defer conn.Close()

	// Create channels for error handling
	errChan := make(chan error, 1)

	// Read from the connection and write to output
	go func() {
		_, err := io.Copy(output, conn)
		errChan <- err
	}()

	// Read from input and write to connection
	_, err := io.Copy(conn, input)
	if err != nil {
		<-errChan // Drain channel
		return err
	}

	// Check for errors from the read goroutine
	if err := <-errChan; err != nil {
		return err
	}

	return nil
}

// Start begins accepting connections and handling them
func (s *TcpServer) Start() error {
	if s.Listener == nil {
		return errors.New("listener not initialized")
	}

	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			return err
		}

		go s.Handler(conn, s.Input, s.Output)
	}
}

// Close stops the server and closes the listener
func (s *TcpServer) Close() error {
	if s.Listener == nil {
		return errors.New("listener not initialized")
	}
	return s.Listener.Close()
}
