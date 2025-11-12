package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/healthz", healthzHandler)

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Orbital Rush Server starting on %s\n", addr)
	fmt.Println("Health check available at /healthz")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]string{
		"status": "ok",
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding healthz response: %v", err)
	}
}
