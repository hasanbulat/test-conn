package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

const version = "1.1.0"

var (
	mu          sync.Mutex
	connections int
)

func main() {
	log.Printf("Starting service version %s", version)
	go printConnectionsEvery5Seconds()
	http.HandleFunc("/connect", handleConnect)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	// Increment connection count
	mu.Lock()
	connections++
	connectionID := connections
	mu.Unlock()
	defer func() {
		// Decrement connection count when the function returns
		mu.Lock()
		connections--
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
		<-ticker.C
		mu.Lock()
		if connections > 0 {
			log.Printf("Current number of connections: %d", connections)
		}
		mu.Unlock()
	}
}
