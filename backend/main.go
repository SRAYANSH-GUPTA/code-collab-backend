package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codecollab/config"
	"codecollab/handlers"
	"codecollab/utils"
)

func main() {
	
	cfg := config.Load()

	logger := utils.NewLogger("main")
	logger.Info("Starting Code Linting Platform Backend")
	logger.Info("Environment: %s", cfg.Env)
	logger.Info("Port: %s", cfg.Port)
	logger.Info("Mock Lambda: %v", cfg.UseMockLambda)
	logger.Info("Mock Auth: %v", cfg.UseMockAuth)


	http.HandleFunc("/ws", handlers.HandleWebSocket(cfg))
	http.HandleFunc("/health", handlers.HandleHealth)


	http.HandleFunc("/swagger.yaml", handlers.ServeSwaggerYAML)
	http.HandleFunc("/docs", handlers.ServeSwaggerUI)


	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"Code Linting Platform API","version":"1.0.0","endpoints":{"/ws":"WebSocket endpoint","/health":"Health check","/docs":"API Documentation"}}`)
	})

	
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}


	go func() {
		logger.Info("WebSocket endpoint: ws://localhost:%s/ws", cfg.Port)
		logger.Info("Health check: http://localhost:%s/health", cfg.Port)
		logger.Info("API Documentation: http://localhost:%s/docs", cfg.Port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server stopped")
}
