package tcp_server

import (
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

// mockConn implements a fake network connection that we can control
type mockConn struct {
	net.Conn               // embed net.Conn interface
	closed   chan struct{} // channel to signal when Close() is called
}

// Close signals that the connection was closed by closing the channel
func (m *mockConn) Close() error {
	close(m.closed)
	return nil
}

// Read returns EOF immediately to prevent blocking
func (m *mockConn) Read(b []byte) (n int, err error) {
	return 0, io.EOF
}

// Write pretends to write data successfully
func (m *mockConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

// mockListener simulates a network listener that we can control
type mockListener struct {
	acceptCalled bool          // tracks if Accept was called
	connections  chan net.Conn // channel to control what connections are returned
}

// Accept waits for a connection to be sent on the connections channel
func (m *mockListener) Accept() (net.Conn, error) {
	m.acceptCalled = true
	// Block until a connection is provided through the channel
	conn := <-m.connections
	return conn, nil
}

func (m *mockListener) Close() error   { return nil }
func (m *mockListener) Addr() net.Addr { return nil }

func TestStartCreatesGoroutine(t *testing.T) {
	// Create channel to control when connections are returned by Accept
	connChan := make(chan net.Conn)
	mock := &mockListener{connections: connChan}
	server := NewTcpServer(mock, nil, nil)

	// Start server in a goroutine because Start() blocks
	go server.Start()

	// Create a mock connection with a channel to signal when it's closed
	mockConn := &mockConn{closed: make(chan struct{})}

	// Send the connection through our channel
	// This will cause Accept() to return this connection
	connChan <- mockConn

	// Check if the connection was handled by verifying Close was called
	select {
	case <-mockConn.closed:
		// Success: this means:
		// 1. Start() accepted the connection
		// 2. Started a goroutine with the handler
		// 3. Handler called Close() on the connection (due to defer)
	case <-time.After(100 * time.Millisecond):
		// If we get here, either:
		// 1. The goroutine wasn't created
		// 2. The handler didn't run
		// 3. Close wasn't called
		t.Error("connection was not handled in goroutine")
	}
}

func TestClose(t *testing.T) {
	// Test case 1: nil listener
	server := &TcpServer{Listener: nil}
	if err := server.Close(); err == nil {
		t.Error("expected error for nil listener, got nil")
	}

	// Test case 2: initialized listener
	mock := &mockListener{}
	server = &TcpServer{Listener: mock}
	if err := server.Close(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// Test the errors in the handler

// mockConnRW is a mock implementation of net.Conn that allows us to:
// - Inject custom readers and writers for testing different scenarios
// - Track when the connection is closed
// - Control the behavior of Read and Write operations
type mockConnRW struct {
	net.Conn               // Embed net.Conn interface
	reader   io.Reader     // Custom reader for controlling Read behavior
	writer   io.Writer     // Custom writer for controlling Write behavior
	closed   chan struct{} // Channel to track when Close is called
}

func (m *mockConnRW) Read(p []byte) (n int, err error) {
	return m.reader.Read(p) // Delegate to custom reader
}

func (m *mockConnRW) Write(p []byte) (n int, err error) {
	return m.writer.Write(p) // Delegate to custom writer
}

func (m *mockConnRW) Close() error {
	close(m.closed) // Signal that Close was called
	return nil
}

// mockReader simulates a reader that returns a specified error and then EOF
type mockReader struct {
	err       error
	callCount int
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	m.callCount++
	if m.callCount == 1 && m.err != nil {
		return 0, m.err
	}
	return 0, io.EOF // Return EOF after first call or if no error specified
}

// mockWriter simulates a writer that always succeeds
type mockWriter struct{}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestDefaultHandler(t *testing.T) {
	// Test case 1: Error when copying from input to connection
	t.Run("input copy error", func(t *testing.T) {
		expectedErr := errors.New("input error")
		mockConn := &mockConnRW{
			reader: &mockReader{}, // Will return EOF
			writer: &mockWriter{}, // Always succeeds
			closed: make(chan struct{}),
		}

		// Create input that will error on first call and then EOF
		input := &mockReader{err: expectedErr}
		output := &mockWriter{}

		err := defaultHandler(mockConn, input, output)
		if err.Error() != expectedErr.Error() {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})

	// Test case 2: Error when copying from connection to output
	t.Run("connection read error", func(t *testing.T) {
		expectedErr := errors.New("conn read error")
		mockConn := &mockConnRW{
			reader: &mockReader{err: expectedErr}, // Connection read will error then EOF
			writer: &mockWriter{},                 // Always succeeds
			closed: make(chan struct{}),
		}

		// Create normal input/output (input will return EOF)
		input := &mockReader{}
		output := &mockWriter{}

		err := defaultHandler(mockConn, input, output)
		if err.Error() != expectedErr.Error() {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}
