package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Start sets up and starts the HTTP server
func Start(handler http.Handler, addr string, cleanup func()) {
	defer cleanup()

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Run in goroutine
	go func() {
		fmt.Printf("🚀 Server running at http://localhost%s\n", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	gracefulShutdown(server)
}

func gracefulShutdown(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server Shutdown Failed:%+v\n", err)
	}

	fmt.Println("✅ Server stopped gracefully")
}
