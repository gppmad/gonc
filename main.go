// Basic usage, initiate a TCP Connection passing host and port as parameters.
package main

import (
	"errors"
	"fmt"
	"io"
	"log"
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
		fmt.Println("Data Read from Conn")
		_, err := io.Copy(c.Output, c.Conn)
		if err != nil {
			errChan <- err
			return
		}
		fmt.Println("Close Data Read from Conn")
		errChan <- nil
	}()

	fmt.Println("Data Write to Conn")
	_, err := io.Copy(c.Conn, c.Input)
	if err != nil {
		return errors.New("error writing in the connection")
	}

	if err := <-errChan; err != nil {
		return errors.New("error reading from the connection: " + err.Error())
	}
	fmt.Println("Close Data Write to Conn")
	return nil
}

// Close the connection
func (c *TcpClient) Close() error {
	return c.Conn.Close()
}

func main() {

	// Connect to remote server.
	remoteAddr := "tcpbin.com:4242"

	// Connect to remote server.
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		// handle error
		log.Fatal("Error during connection: ", err)
	}

	tcpClient := NewTcpClient(conn, nil, nil)
	fmt.Println("Connected to a TCP Server")

	// Start the connection
	err = tcpClient.Start()
	if err != nil {
		// handle error
		log.Fatal("Error during starting the proxy connection: ", err)
	}

	// Close the connection

	err = tcpClient.Close()
	if err != nil {
		// handle error
		log.Fatal("Error during closing the proxy connection: ", err)
	}

}
