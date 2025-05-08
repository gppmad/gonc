package tcp_client

import (
	"bytes"
	"errors"
	"net"
	"strings"
	"testing"
	"time"
)

type myConn struct {
	OutgoingBuffer bytes.Buffer // from the client to server
	IncomingBuffer bytes.Buffer // from the server to client
	closeFlag      bool
	readError      error // Error to return after first successful read
}

func (c *myConn) Read(b []byte) (n int, err error) {
	if c.closeFlag {
		return 0, errors.New("connection is closed, cannot read")
	}

	// If no bytes are read and we have a readError, return it
	n, err = c.IncomingBuffer.Read(b)
	if n == 0 && c.readError != nil {
		return 0, c.readError
	}

	return n, err
}

func (c *myConn) Write(b []byte) (n int, err error) {
	if c.closeFlag {
		return 0, errors.New("connection is closed, cannot write")
	}
	return c.OutgoingBuffer.Write(b)
}

func (c *myConn) Close() error {
	c.closeFlag = true
	return nil
}

func (c *myConn) LocalAddr() net.Addr {
	return &net.IPAddr{IP: net.IPv4(127, 0, 0, 1)}
}

func (c *myConn) RemoteAddr() net.Addr {
	return &net.IPAddr{IP: net.IPv4(127, 0, 0, 1)}
}

func (c *myConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *myConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *myConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestServerStart(t *testing.T) {
	// I want to better describe this test, to make my understanding more solid.
	// I assign an input value to the connection (using a buffer coming from a fixed string),
	// L1: So my TCP Client will accept my mocked connection, this input and a pointer to an empty buffer which represent the output.
	// L2: My mocked connection since at the end contains two buffers, will be instantiated with a message to
	// the "incomingBuffer" and the an empty buffer as "outgoingBuffer"
	// The value of the test is to test that :
	// at level 1 (L1) the output is basically equal to the incomingBuffer
	// at level 2 (L2) the input is basically equal to the outgoingBuffer

	// Define something that mock the standard input
	inputStr := "my input"
	input := bytes.NewBufferString(inputStr)

	// Define something that mock the incoming from the server
	incomingMessage := bytes.NewBufferString("this is coming from the server")

	// Populate the connection
	myConn := &myConn{}
	myConn.IncomingBuffer.WriteString(incomingMessage.String())
	myConnOutput := new(bytes.Buffer)

	// New TCP Server
	client := NewTcpClient(myConn, input, myConnOutput)
	client.Start()

	// Checks
	if myConn.OutgoingBuffer.String() != inputStr {
		t.Fatalf("server should receive %q from stdin, got %q",
			"my input", myConn.OutgoingBuffer.String())
	}

	if myConnOutput.String() != incomingMessage.String() {
		t.Fatalf("stdout should display %q from server, got %q",
			"this is coming from the server", myConnOutput.String())
	}
}

func TestServerErrorConnClosed(t *testing.T) {
	// Define something that mock the standard input
	inputStr := "my input"
	input := bytes.NewBufferString(inputStr)

	// Define something that mock the incoming from the server
	incomingMessage := bytes.NewBufferString("this is coming from the server")

	// Populate the connection
	myConn := &myConn{}
	myConn.IncomingBuffer.WriteString(incomingMessage.String())
	myConnOutput := new(bytes.Buffer)

	// Close immediately the connection
	myConn.Close()
	client := NewTcpClient(myConn, input, myConnOutput)
	err := client.Start()

	if err == nil {
		t.Fatal("The connection should return an error")
	}

	// Verify the error message contains information about the connection
	if !strings.Contains(err.Error(), "connection") {
		t.Fatalf("expected error message to mention connection issues, got: %v", err)
	}

}

func TestServerErrorReadingFromConnection(t *testing.T) {
	inputStr := "my input"
	input := bytes.NewBufferString(inputStr)

	myConn := &myConn{
		readError: errors.New("simulated read error"),
	}
	// Add just enough data for one read operation
	myConn.IncomingBuffer.WriteString("initial data")
	output := new(bytes.Buffer)

	client := NewTcpClient(myConn, input, output)
	err := client.Start()

	if err == nil {
		t.Fatal("Expected an error from reading the connection, but got nil")
	}

	if !strings.Contains(err.Error(), "error reading from the connection") {
		t.Fatalf("Expected error message to mention reading from connection, got: %v", err)
	}
}
