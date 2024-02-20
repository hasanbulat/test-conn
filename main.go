package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const version = "1.1.0"

var (
	mu          sync.Mutex
	connections int
	wg          sync.WaitGroup
)

func main() {
	log.Printf("Starting service version %s", version)
	go printConnectionsEvery5Seconds()

	// Create a context to manage the lifecycle of the server
	ctx, cancel := context.WithCancel(context.Background())

	// Handle OS signals for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Println("Received stop signal. Shutting down gracefully...")
		cancel() // Cancel the context to stop the server gracefully
	}()

	// Start the HTTP server with the provided context
	server := &http.Server{Addr: ":8080"}
	http.HandleFunc("/connect", handleConnect)
	wg.Add(1) // Increment wait group counter for the server
	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %s", err)
		}
	}()

	// Wait for all connections to be closed or the context being canceled
	go func() {
		wg.Wait()
		cancel() // Cancel the context after all connections are closed
	}()

	// Wait for the stop signal or the context being canceled
	<-ctx.Done()

	// Shutdown the server gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down server: %s", err)
	}

	log.Println("Server stopped")
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	// Increment connection count and add to the wait group
	mu.Lock()
	connections++
	wg.Add(1)
	connectionID := connections
	mu.Unlock()
	defer func() {
		// Decrement connection count and wait group when the function returns
		mu.Lock()
		connections--
		wg.Done()
		mu.Unlock()
	}()

	// Set headers
	w.Header().Set("Content-Type", "application/json")

	// Start time
	startTime := time.Now()

	// Simulate some processing for 60 seconds
	for {
		elapsedTime := time.Since(startTime)
		if elapsedTime.Seconds() >= 60 {
			break
		}
		log.Printf("Connection %d - Time elapsed: %.0f seconds", connectionID, elapsedTime.Seconds())
		time.Sleep(1 * time.Second)
	}

	// Create a response JSON message
	response := map[string]interface{}{
		"message":     "Connection kept for 60 seconds. Task completed.",
		"connections": connections,
	}

	// Marshal the response JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

	// Write the JSON response
	_, err = w.Write(jsonResponse)
	if err != nil {
		http.Error(w, "Error writing JSON response", http.StatusInternalServerError)
		return
	}
}

func printConnectionsEvery5Seconds() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			mu.Lock()
			if connections > 0 {
				log.Printf("Current number of connections: %d", connections)
			}
			mu.Unlock()
		}
	}
}
