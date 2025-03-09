// Basic usage, initiate a TCP Connection passing host and port as parameters.
package main

import (
	"errors"
	"io"
	"log"
	"net"
	"os"
)

type TcpClient struct {
	RemoteAddr string
	Input      io.Reader
	Output     io.Writer
	conn       net.Conn
}

func NewTcpClient(RemoteAddr string, input io.Reader, output io.Writer) *TcpClient {
	if input == nil {
		input = os.Stdin
	}

	if output == nil {
		output = os.Stdout
	}

	return &TcpClient{RemoteAddr: RemoteAddr, Input: input, Output: output}
}

func (c *TcpClient) Connect() error {
	// Connect to remote server.
	conn, err := net.Dial("tcp", c.RemoteAddr)
	if err != nil {
		return err
	}
	// Assign the conn to the struct field.
	c.conn = conn
	return nil
}

// Before call this server initialize the connection with Connect()
func (c *TcpClient) Start() error {
	if c.conn == nil {
		return errors.New("connect to the target before initialize a new connection")
	}
	defer c.conn.Close()

	errChan := make(chan error, 1)

	// Read from the connection.
	go func() {
		_, err := io.Copy(c.Output, c.conn)
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	_, err := io.Copy(c.conn, c.Input)
	if err != nil {
		return errors.New("error writing in the connection")
	}

	if err := <-errChan; err != nil {
		return errors.New("error reading from the connection: " + err.Error())
	}
	return nil
}

func main() {

	// Connect to remote server.
	remoteAddr := "tcpbin.com:4242"
	tcpClient := NewTcpClient(remoteAddr, nil, nil)

	err := tcpClient.Connect()
	if err != nil {
		// handle error
		log.Fatal("Error during connection: ", err)
	}

	err = tcpClient.Start()
	if err != nil {
		// handle error
		log.Fatal("Error during proxy operation: ", err)
	}
}
