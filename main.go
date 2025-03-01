// Basic usage, initiate a TCP Connection passing host and port as parameters.
package main

import (
	"io"
	"log"
	"net"
	"os"
)

func main() {

	// Connect to remote server.
	conn, err := net.Dial("tcp", "tcpbin.com:4242")
	if err != nil {
		// handle error
		log.Fatal("Error during connection")
	}
	defer conn.Close()

	// Read from the connection.
	go func() {
		_, err = io.Copy(os.Stdout, conn)
		if err != nil {
			log.Fatal("Error reading from the connection")
		}
	}()

	// Write to the connection.
	_, err = io.Copy(conn, os.Stdin)
	if err != nil {
		log.Fatal("Error writing in the connection")
	}

}
