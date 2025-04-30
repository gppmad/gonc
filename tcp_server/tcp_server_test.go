package tcp_server

import (
	"bytes"
	"errors"
	"net"
	"strings"
	"testing"
	"time"
)

// MockListener implements net.Listener for testing
type MockListener struct {
	acceptChan chan net.Conn
	closeChan  chan struct{}
	closed     bool
}

func NewMockListener() *MockListener {
	return &MockListener{
		acceptChan: make(chan net.Conn, 1),
		closeChan:  make(chan struct{}),
	}
}

func (m *MockListener) Accept() (net.Conn, error) {
	select {
	case conn := <-m.acceptChan:
		return conn, nil
	case <-m.closeChan:
		return nil, errors.New("listener closed")
	}
}

func (m *MockListener) Close() error {
	if m.closed {
		return errors.New("listener already closed")
	}
	m.closed = true
	close(m.closeChan)
	return nil
}

func (m *MockListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080}
}

// MockConnection implements net.Conn for testing
type MockConnection struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	closed   bool
}

func NewMockConnection() *MockConnection {
	return &MockConnection{
		readBuf:  new(bytes.Buffer),
		writeBuf: new(bytes.Buffer),
	}
}

func (m *MockConnection) Read(b []byte) (n int, err error) {
	if m.closed {
		return 0, errors.New("connection closed")
	}
	return m.readBuf.Read(b)
}

func (m *MockConnection) Write(b []byte) (n int, err error) {
	if m.closed {
		return 0, errors.New("connection closed")
	}
	return m.writeBuf.Write(b)
}

func (m *MockConnection) Close() error {
	m.closed = true
	return nil
}

func (m *MockConnection) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080}
}

func (m *MockConnection) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 12345}
}

func (m *MockConnection) SetDeadline(t time.Time) error      { return nil }
func (m *MockConnection) SetReadDeadline(t time.Time) error  { return nil }
func (m *MockConnection) SetWriteDeadline(t time.Time) error { return nil }

func TestTcpServer(t *testing.T) {
	// Create mock listener and connection
	listener := NewMockListener()
	conn := NewMockConnection()

	// Setup test data
	testInput := "test input"
	testResponse := "test response"
	conn.readBuf.WriteString(testResponse)

	// Create server with mock I/O
	serverInput := bytes.NewBufferString(testInput)
	serverOutput := new(bytes.Buffer)
	server := NewTcpServer(listener, serverInput, serverOutput)

	// Start server in a goroutine
	go func() {
		listener.acceptChan <- conn
	}()

	// Start server and wait for a short time
	go server.Start()
	time.Sleep(100 * time.Millisecond)

	// Verify the connection received the input
	if got := conn.writeBuf.String(); got != testInput {
		t.Errorf("Expected connection to receive %q, got %q", testInput, got)
	}

	// Verify the server output
	if got := serverOutput.String(); got != testResponse {
		t.Errorf("Expected server output %q, got %q", testResponse, got)
	}
}

func TestTcpServerClose(t *testing.T) {
	listener := NewMockListener()
	server := NewTcpServer(listener, nil, nil)

	// Test closing the server
	err := server.Close()
	if err != nil {
		t.Errorf("Expected no error when closing server, got %v", err)
	}

	// Test closing an already closed server
	err = server.Close()
	if err == nil {
		t.Error("Expected error when closing already closed server, got nil")
	}
}

func TestTcpServerNilListener(t *testing.T) {
	server := NewTcpServer(nil, nil, nil)

	// Test starting server with nil listener
	err := server.Start()
	if err == nil {
		t.Error("Expected error when starting server with nil listener, got nil")
	}
	if !strings.Contains(err.Error(), "listener not initialized") {
		t.Errorf("Expected error to contain 'listener not initialized', got %v", err)
	}

	// Test closing server with nil listener
	err = server.Close()
	if err == nil {
		t.Error("Expected error when closing server with nil listener, got nil")
	}
	if !strings.Contains(err.Error(), "listener not initialized") {
		t.Errorf("Expected error to contain 'listener not initialized', got %v", err)
	}
}

// MockListenerError always returns an error on Accept
type MockListenerError struct{}

func (m *MockListenerError) Accept() (net.Conn, error) {
	return nil, errors.New("accept error")
}
func (m *MockListenerError) Close() error { return nil }
func (m *MockListenerError) Addr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 9999}
}

func TestTcpServerAcceptError(t *testing.T) {
	listener := &MockListenerError{}
	server := NewTcpServer(listener, nil, nil)
	err := server.Start()
	if err == nil || !strings.Contains(err.Error(), "accept error") {
		t.Errorf("Expected accept error, got %v", err)
	}
}

// MockWriterError always returns an error on Write
type MockWriterError struct{}

func (m *MockWriterError) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}

func TestTcpServerHandleConnectionOutputError(t *testing.T) {
	conn := NewMockConnection()
	conn.readBuf.WriteString("data") // so io.Copy tries to write
	server := NewTcpServer(nil, bytes.NewBufferString(""), &MockWriterError{})
	// Call handleConnection directly
	server.handleConnection(conn)
	// No panic = pass; can't check return, but coverage will hit the error path
}

// MockReaderError always returns an error on Read
type MockReaderError struct{}

func (m *MockReaderError) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestTcpServerHandleConnectionInputError(t *testing.T) {
	conn := NewMockConnection()
	server := NewTcpServer(nil, &MockReaderError{}, new(bytes.Buffer))
	// Call handleConnection directly
	server.handleConnection(conn)
	// No panic = pass; can't check return, but coverage will hit the error path
}
