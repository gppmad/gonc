package tcp_client

import (
	"bytes"
	"net"
	"testing"
	"time"
)

type myConn struct {
	OutgoingBuffer bytes.Buffer // from the client to server
	IncomingBuffer bytes.Buffer // from the server to client
	closeFlag      bool
}

func (c *myConn) Read(b []byte) (n int, err error) {
	return c.IncomingBuffer.Read(b)
}

func (c *myConn) Write(b []byte) (n int, err error) {
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
