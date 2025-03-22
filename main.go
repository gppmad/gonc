package main

import (
	"fmt"
	"log"
	"net"

	tcp_client "github.com/gppmad/gonc/tcp_client"
)

func main() {

	// Connect to remote server.
	remoteAddr := "tcpbin.com:4242"

	// Connect to remote server.
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		// handle error
		log.Fatal("Error during connection: ", err)
	}

	tcpClient := tcp_client.NewTcpClient(conn, nil, nil)
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
