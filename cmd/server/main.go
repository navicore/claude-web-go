package main

import (
	"log"
	"net/http"
	"os"

	"claude-web-go/internal/api"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server, err := api.NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	
	router := mux.NewRouter()

	router.HandleFunc("/api/chat", server.HandleChat).Methods("POST")
	router.HandleFunc("/api/files/{sessionId}/{filename}", server.HandleFile).Methods("GET")
	router.HandleFunc("/api/ws", server.HandleWebSocket)
	
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	handler := c.Handler(router)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}