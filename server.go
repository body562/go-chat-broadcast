package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
)

// Server configuration
const (
	Port = ":8080"
)

// Client structure to hold connection info
type Client struct {
	conn net.Conn
	id   string
}

// Message structure to pass data to the broadcaster
type Message struct {
	sender  net.Conn // We need this to avoid self-echo
	content string
}

// Global state
var (
	// Map of active clients: Key = net.Conn, Value = User ID
	clients   = make(map[net.Conn]string)
	clientsMu sync.Mutex // Mutex to protect the clients map

	// Channel to handle all incoming messages efficiently
	broadcast = make(chan Message)
)

func main() {
	// Start the listener
	listener, err := net.Listen("tcp", Port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Printf("Server started on port %s\n", Port)

	// Start the broadcaster goroutine in the background
	go handleMessages()

	// Accept loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		// Handle each client in a new goroutine
		go handleClient(conn)
	}
}

// handleClient: Manages a single client connection
func handleClient(conn net.Conn) {
	defer conn.Close()

	// 1. Get User ID (Simple implementation: Use Remote Address)
	userID := conn.RemoteAddr().String()

	// 2. Add to shared list (Protected by Mutex)
	clientsMu.Lock()
	clients[conn] = userID
	clientsMu.Unlock()

	// 3. Notify others that user joined
	joinMsg := fmt.Sprintf("User [%s] joined", userID)
	broadcast <- Message{sender: conn, content: joinMsg}

	// 4. Listen for incoming messages from this client
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		
		// Send received message to the broadcast channel
		fullMsg := fmt.Sprintf("User [%s]: %s", userID, text)
		broadcast <- Message{sender: conn, content: fullMsg}
	}

	// 5. Cleanup on disconnect
	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()

	// Notify others of disconnect (optional, but good practice)
	leaveMsg := fmt.Sprintf("User [%s] left", userID)
	broadcast <- Message{sender: conn, content: leaveMsg}
}

// handleMessages: The central hub that broadcasts to all clients
func handleMessages() {
	for {
		// Grab next message from channel
		msg := <-broadcast

		// Lock the map to iterate safely
		clientsMu.Lock()
		for clientConn := range clients {
			// CRITICAL: No self-echo check
			// We only send if the clientConn is NOT the sender
			if clientConn != msg.sender {
				fmt.Fprintln(clientConn, msg.content)
			}
		}
		clientsMu.Unlock()
	}
}