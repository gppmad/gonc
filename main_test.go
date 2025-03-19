package main

import (
	"bytes"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

// Mock implementation is a simple implementation of the Reader interface.
type MockReader struct {
	data string
	pos  int
}

// NewMockReader creates a new MockReader.
func NewMockReader(data string) *MockReader {
	return &MockReader{data: data}
}

// Read reads up to len(p) bytes into p from the predefined string.
func (r *MockReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// MockWriter is a simple implementation of the Writer interface that writes to a buffer.
type MockWriter struct {
	buf bytes.Buffer
}

// NewMockWriter creates a new MockWriter.
func NewMockWriter() *MockWriter {
	return &MockWriter{}
}

// Write writes len(p) bytes from p to the buffer.
func (w *MockWriter) Write(p []byte) (n int, err error) {
	return w.buf.Write(p)
}

// String returns the contents of the buffer as a string.
func (w *MockWriter) String() string {
	return w.buf.String()
}

// Mock implementation of net.Conn

// MockConn is a mock implementation of the net.Conn interface.
type MockConn struct {
	ReadBuffer  bytes.Buffer
	WriteBuffer bytes.Buffer
	Closed      bool
}

// NewMockConn creates a new MockConn.
func NewMockConn() *MockConn {
	return &MockConn{}
}

// Read reads data from the ReadBuffer.
func (c *MockConn) Read(b []byte) (n int, err error) {
	if c.Closed {
		return 0, errors.New("connection is closed")
	}
	return c.ReadBuffer.Read(b)
}

// Write writes data to the WriteBuffer.
func (c *MockConn) Write(b []byte) (n int, err error) {
	if c.Closed {
		return 0, errors.New("connection is closed")
	}
	return c.WriteBuffer.Write(b)
}

// Close closes the connection.
func (c *MockConn) Close() error {
	c.Closed = true
	return nil
}

// LocalAddr returns the local network address.
func (c *MockConn) LocalAddr() net.Addr {
	return &net.IPAddr{IP: net.IPv4(127, 0, 0, 1)}
}

// RemoteAddr returns the remote network address.
func (c *MockConn) RemoteAddr() net.Addr {
	return &net.IPAddr{IP: net.IPv4(127, 0, 0, 1)}
}

// SetDeadline sets the read and write deadlines associated with the connection.
func (c *MockConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future Read calls.
func (c *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls.
func (c *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestTcpClient(t *testing.T) {
	mockReader := NewMockReader("Hello")
	mockWriter := NewMockWriter()
	mockConn := NewMockConn()

	tcpClient := NewTcpClient(mockConn, mockReader, mockWriter)

	err := tcpClient.Start()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockWriter.String() != "Hello" {
		t.Fatalf("expected to write %q, wrote %q", "Hello", mockWriter.String())
	}

}
