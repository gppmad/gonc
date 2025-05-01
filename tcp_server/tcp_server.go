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
}

// NewTcpServer creates a new TCP server instance with the specified network listener and I/O streams.
//
// Parameters:
//   - listener: The TCP network listener that accepts incoming connections
//   - input: Reader for server input (defaults to os.Stdin if nil)
//   - output: Writer for server output (defaults to os.Stdout if nil)
//
// Returns:
//
//	A pointer to a fully initialized TcpServer instance ready to handle connections
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
	}
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

		go s.handleConnection(conn)
	}
}

// handleConnection manages a single client connection
func (s *TcpServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Create channels for error handling
	errChan := make(chan error, 1)

	// Read from the connection and write to output
	go func() {
		_, err := io.Copy(s.Output, conn)
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	// Read from input and write to connection
	_, err := io.Copy(conn, s.Input)
	if err != nil {
		return
	}

	// Check for errors from the read goroutine
	if err := <-errChan; err != nil {
		return
	}
}

// Close stops the server and closes the listener
func (s *TcpServer) Close() error {
	if s.Listener == nil {
		return errors.New("listener not initialized")
	}
	return s.Listener.Close()
}
