package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Println("Connected to chat server. Type a message and press Enter.")

	// Goroutine to listen for incoming messages from Server
	go func() {
		reader := bufio.NewReader(conn)
		for {
			msg, err := reader.ReadString('\n')
			if err == io.EOF {
				fmt.Println("Disconnected from server.")
				os.Exit(0)
			}
			// Print message from other users
			fmt.Print(msg)
		}
	}()

	// Main thread: Read from user input and send to Server
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		fmt.Fprintf(conn, "%s\n", text)
	}
}