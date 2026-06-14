package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go-shorturl/internal/handler"
	"go-shorturl/internal/store"
)

func main() {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	store := store.NewMemoryStore()
	h := handler.New(store)

	mux := http.NewServeMux()

	// API
	mux.HandleFunc("/api/shorten", h.Shorten)
	mux.HandleFunc("/api/stats/", h.Stats)
	mux.HandleFunc("/api/health", h.Health)

	// Short URL redirect + home page
	mux.HandleFunc("/", h.Redirect)

	// Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		log.Println("Shutting down...")
		os.Exit(0)
	}()

	log.Printf("🚀 ShortURL service starting on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
