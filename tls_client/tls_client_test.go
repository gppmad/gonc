package tls_client

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

// Mock connection implementing net.Conn interface with explicit stdin/stdout simulation
type mockStreamConn struct {
	simulatedStdoutData []byte        // Data that will be read from connection (simulating stdout)
	simulatedStdinBuf   *bytes.Buffer // Buffer that stores data written to connection (simulating stdin)
	closed              bool
	readError           error
	writeError          error
}

func (m *mockStreamConn) Read(p []byte) (n int, err error) {
	if m.readError != nil {
		return 0, m.readError
	}
	if len(m.simulatedStdoutData) == 0 {
		return 0, io.EOF
	}
	n = copy(p, m.simulatedStdoutData)
	m.simulatedStdoutData = m.simulatedStdoutData[n:]
	return n, nil
}

func (m *mockStreamConn) Write(p []byte) (n int, err error) {
	if m.writeError != nil {
		return 0, m.writeError
	}
	return m.simulatedStdinBuf.Write(p)
}

func (m *mockStreamConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockStreamConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8443}
}

func (m *mockStreamConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 43123}
}

func (m *mockStreamConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockStreamConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockStreamConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestTlsClientStart(t *testing.T) {
	// Setup
	simulatedInput := bytes.NewBufferString("test input")
	simulatedOutput := new(bytes.Buffer)
	simulatedStdinBuf := new(bytes.Buffer)
	mockConn := &mockStreamConn{
		simulatedStdoutData: []byte("test response"),
		simulatedStdinBuf:   simulatedStdinBuf,
	}

	client := NewTlsClient(mockConn, simulatedInput, simulatedOutput)

	// Test
	err := client.Start()

	// Verify results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if got, want := simulatedOutput.String(), "test response"; got != want {
		t.Errorf("Expected output %q, got %q", want, got)
	}

	if got, want := simulatedStdinBuf.String(), "test input"; got != want {
		t.Errorf("Expected input %q, got %q", want, got)
	}
}

func TestTlsClientStartWithReadError(t *testing.T) {
	// Setup
	simulatedInput := bytes.NewBufferString("test input")
	simulatedOutput := new(bytes.Buffer)
	mockConn := &mockStreamConn{
		simulatedStdoutData: []byte("test response"),
		simulatedStdinBuf:   new(bytes.Buffer),
		readError:           errors.New("simulated read error"),
	}

	client := NewTlsClient(mockConn, simulatedInput, simulatedOutput)

	// Test
	err := client.Start()

	// Verify error is propagated
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "simulated read error") {
		t.Errorf("Expected error to contain 'simulated read error', got %v", err)
	}
}

func TestTlsClientStartWithWriteError(t *testing.T) {
	simulatedInput := bytes.NewBufferString("test input")
	simulatedOutput := new(bytes.Buffer)
	mockConn := &mockStreamConn{
		simulatedStdoutData: []byte("test response"),
		simulatedStdinBuf:   new(bytes.Buffer),
		writeError:          errors.New("simulated write error"),
	}

	client := NewTlsClient(mockConn, simulatedInput, simulatedOutput)

	err := client.Start()

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "error writing in the connection") {
		t.Errorf("Expected error to contain 'error writing in the connection', got %v", err)
	}
}

func TestTlsClientStartWithNilConn(t *testing.T) {
	// Setup
	client := &TlsClient{
		Input:  bytes.NewBufferString(""),
		Output: new(bytes.Buffer),
		Conn:   nil,
	}

	// Test
	err := client.Start()

	// Verify error
	if err == nil {
		t.Error("Expected error for nil connection, got nil")
	}
}

func TestTlsClientClose(t *testing.T) {
	// Setup
	mockConn := &mockStreamConn{
		simulatedStdoutData: []byte{},
		simulatedStdinBuf:   new(bytes.Buffer),
	}

	client := NewTlsClient(mockConn, nil, nil)

	// Test
	err := client.Close()

	// Verify results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !mockConn.closed {
		t.Error("Expected connection to be closed")
	}
}

func TestConnect(t *testing.T) {
	originalTLSDial := tlsDial
	defer func() { tlsDial = originalTLSDial }()

	t.Run("successful connection", func(t *testing.T) {
		tlsDial = func(network, addr string, config *tls.Config) (*tls.Conn, error) {
			return &tls.Conn{}, nil
		}

		conn, err := Connect("example.com:443", nil)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if conn == nil {
			t.Error("Expected connection, got nil")
		}
	})

	t.Run("failed connection", func(t *testing.T) {
		tlsDial = func(network, addr string, config *tls.Config) (*tls.Conn, error) {
			return nil, errors.New("mock connection error")
		}

		conn, err := Connect("example.com:443", nil)
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if conn != nil {
			t.Error("Expected nil connection, got connection")
		}
	})
}
