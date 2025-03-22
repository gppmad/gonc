package tcp_client

import (
	"errors"
	"io"
	"net"
	"os"
)

type TcpClient struct {
	Input  io.Reader
	Output io.Writer
	Conn   net.Conn
}

func NewTcpClient(conn net.Conn, input io.Reader, output io.Writer) *TcpClient {
	if input == nil {
		input = os.Stdin
	}

	if output == nil {
		output = os.Stdout
	}

	return &TcpClient{Conn: conn, Input: input, Output: output}
}

// Before call this server initialize the connection with Connect()
func (c *TcpClient) Start() error {
	if c.Conn == nil {
		return errors.New("connect to the target before initialize a new connection")
	}

	errChan := make(chan error, 1)

	// Read from the connection.
	go func() {
		_, err := io.Copy(c.Output, c.Conn)
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	_, err := io.Copy(c.Conn, c.Input)
	if err != nil {
		return errors.New("error writing in the connection")
	}

	if err := <-errChan; err != nil {
		return errors.New("error reading from the connection: " + err.Error())
	}
	return nil
}

// Close the connection
func (c *TcpClient) Close() error {
	return c.Conn.Close()
}
