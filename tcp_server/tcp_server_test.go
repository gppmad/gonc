package tcp_server_test

import (
	"errors"
	"io"
	"net"
	"testing"
	"time"

	tcp_server "github.com/gppmad/gonc/tcp_server"
)

// Create a mock listener
type MockListener struct {
	// Derive from the listener interfaces the other methods that I don't need to use.
	net.Listener
	connections chan net.Conn
}

func (l *MockListener) Accept() (net.Conn, error) {
	conn := <-l.connections
	return conn, nil
}

// Create a conn listener
type MockConn struct {
	net.Conn
	closed chan struct{} // channel to signal when Close() is called

}

func (c *MockConn) Read(b []byte) (n int, err error) {
	return n, io.EOF
}

func (c *MockConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (c *MockConn) Close() error {
	close(c.closed)
	return nil
}

func TestHappyPath(t *testing.T) {
	connChan := make(chan net.Conn)
	mockListener := &MockListener{connections: connChan}
	server := tcp_server.NewTcpServer(mockListener, nil, nil)

	// Start the server
	go server.Start()

	// Create a mock connection with a channel to signal when it's closed
	mockConn := &MockConn{closed: make(chan struct{})}

	// Send the connection through our channel
	// This will cause Accept() to return this connection
	connChan <- mockConn

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

// SimpleListener only implements Close for testing Close() functionality
type SimpleListener struct {
	net.Listener
	closeCalled bool
}

func (l *SimpleListener) Close() error {
	l.closeCalled = true
	return nil
}

func (l *SimpleListener) Accept() (net.Conn, error) {
	// This shouldn't be called during Close() testing
	return nil, errors.New("Accept not implemented")
}

func (l *SimpleListener) Addr() net.Addr {
	return nil
}

func TestClose(t *testing.T) {
	// Test case 1: nil listener
	server := &tcp_server.TcpServer{Listener: nil}
	if err := server.Close(); err == nil {
		t.Error("expected error for nil listener, got nil")
	}

	// Test case 2: initialized listener
	mock := &SimpleListener{}
	server = &tcp_server.TcpServer{Listener: mock}
	if err := server.Close(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Optional: verify Close was called
	if !mock.closeCalled {
		t.Error("listener.Close() was not called")
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

		err := tcp_server.DefaultHandler(mockConn, input, output)
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

		err := tcp_server.DefaultHandler(mockConn, input, output)
		if err.Error() != expectedErr.Error() {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}
