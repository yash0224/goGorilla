// client.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

type Message struct {
	Array []int `json:"array"`
}

func main() {
	// Connect to WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer conn.Close()

	// Handle graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Create a scanner for reading user input
	scanner := bufio.NewScanner(os.Stdin)

	// Start a goroutine to handle server responses
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading message:", err)
				return
			}

			var response Message
			if err := json.Unmarshal(msg, &response); err != nil {
				log.Println("Error unmarshaling response:", err)
				continue
			}

			fmt.Printf("Received sorted array: %v\n", response.Array)
			fmt.Print("\nEnter numbers to sort (space-separated) or 'exit' to quit: ")
		}
	}()

	fmt.Println("Connected to server. Enter numbers to sort (space-separated) or 'exit' to quit.")
	for {
		fmt.Print("Enter numbers: ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if input == "exit" {
			break
		}

		// Convert space-separated string to array of integers
		numStrings := strings.Fields(input)
		numbers := make([]int, 0, len(numStrings))

		// Parse each number
		for _, numStr := range numStrings {
			num, err := strconv.Atoi(numStr)
			if err != nil {
				fmt.Printf("Invalid number: %s\n", numStr)
				continue
			}
			numbers = append(numbers, num)
		}

		if len(numbers) == 0 {
			fmt.Println("Please enter valid numbers")
			continue
		}

		// Prepare and send message
		message := Message{
			Array: numbers,
		}

		messageJSON, err := json.Marshal(message)
		if err != nil {
			log.Println("Error marshaling message:", err)
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, messageJSON); err != nil {
			log.Println("Error writing message:", err)
			break
		}

		fmt.Printf("Sent unsorted array: %v\n", numbers)
	}

	// Clean shutdown
	select {
	case <-interrupt:
		fmt.Println("\nInterrupt received, closing connection...")
	default:
		fmt.Println("\nClosing connection...")
	}

	// Close websocket connection
	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("Error during closing websocket:", err)
	}
	select {
	case <-done:
	}
}
