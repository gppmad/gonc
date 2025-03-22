package tcp_client

import (
	"bytes"
	"errors"
	"net"
	"strings"
	"testing"
	"time"
)

// MockNetworkConnection simulates a TCP connection with separate buffers
// for incoming data (from server to client) and outgoing data (from client to server)
type MockNetworkConnection struct {
	IncomingBuffer bytes.Buffer // Data coming from server to client
	OutgoingBuffer bytes.Buffer // Data going from client to server
	Closed         bool
}

// NewMockNetworkConnection creates a new MockNetworkConnection.
func NewMockNetworkConnection() *MockNetworkConnection {
	return &MockNetworkConnection{}
}

// Read simulates reading data sent from the server
func (c *MockNetworkConnection) Read(b []byte) (n int, err error) {
	if c.Closed {
		return 0, errors.New("connection is closed")
	}
	return c.IncomingBuffer.Read(b)
}

// Write simulates sending data to the server
func (c *MockNetworkConnection) Write(b []byte) (n int, err error) {
	if c.Closed {
		return 0, errors.New("connection is closed")
	}
	return c.OutgoingBuffer.Write(b)
}

// Close closes the connection.
func (c *MockNetworkConnection) Close() error {
	c.Closed = true
	return nil
}

// LocalAddr returns the local network address.
func (c *MockNetworkConnection) LocalAddr() net.Addr {
	return &net.IPAddr{IP: net.IPv4(127, 0, 0, 1)}
}

// RemoteAddr returns the remote network address.
func (c *MockNetworkConnection) RemoteAddr() net.Addr {
	return &net.IPAddr{IP: net.IPv4(127, 0, 0, 1)}
}

// SetDeadline sets the read and write deadlines associated with the connection.
func (c *MockNetworkConnection) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future Read calls.
func (c *MockNetworkConnection) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls.
func (c *MockNetworkConnection) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestTcpClient(t *testing.T) {
	// Setup mock connection with simulated server response
	mockConnection := NewMockNetworkConnection()
	mockConnection.IncomingBuffer.WriteString("Hello World")

	// Mock stdin (user input to be sent to server)
	userInput := bytes.NewBufferString("Hello World")

	// Mock stdout (where server responses will be written)
	serverOutput := new(bytes.Buffer)

	// Create client with mocks
	tcpClient := NewTcpClient(mockConnection, userInput, serverOutput)

	// Start in goroutine since it blocks
	done := make(chan struct{})
	go func() {
		tcpClient.Start()
		close(done)
	}()

	// Let the goroutine execute
	time.Sleep(50 * time.Millisecond)
	mockConnection.Close()
	<-done

	// Verify server received what we sent from stdin
	if mockConnection.OutgoingBuffer.String() != "Hello World" {
		t.Fatalf("server should receive %q from stdin, got %q",
			"Hello World", mockConnection.OutgoingBuffer.String())
	}

	// Verify stdout received what server sent back
	if serverOutput.String() != "Hello World" {
		t.Fatalf("stdout should display %q from server, got %q",
			"Hello World", serverOutput.String())
	}
}

func TestTcpClientError(t *testing.T) {
	// Create a mock connection that will generate errors
	mockConnection := NewMockNetworkConnection()

	// Immediately close the connection to simulate failure
	mockConnection.Close()

	// Setup input/output
	userInput := bytes.NewBufferString("Hello World")
	serverOutput := new(bytes.Buffer)

	// Create client with the closed connection
	tcpClient := NewTcpClient(mockConnection, userInput, serverOutput)

	// The Start method should return an error
	err := tcpClient.Start()

	// Test should fail if no error is returned
	if err == nil {
		t.Fatal("expected an error when using a closed connection, got nil")
	}

	// Verify the error message contains information about the connection
	if !strings.Contains(err.Error(), "connection") {
		t.Fatalf("expected error message to mention connection issues, got: %v", err)
	}
}
