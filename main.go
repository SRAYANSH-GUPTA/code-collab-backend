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
	"codecollab/metrics"
	"codecollab/middleware"
	"codecollab/utils"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// @title Code Collaboration Backend API
// @version 1.0
// @description Backend API for real-time code collaboration platform with WebSocket communication and voice calls using SFU architecture
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@codecollab.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @schemes ws wss http https

// @securityDefinitions.apikey BearerAuth
// @in query
// @name token
// @description Authentication token passed as query parameter for WebSocket connections

func main() {

	cfg := config.Load()

	logger := utils.NewLogger("main")

	handlers.InitVoiceService(cfg.STUNServers, cfg.TURNServers, cfg.TURNUsername, cfg.TURNPassword, cfg.PublicIP, cfg.UDPMinPort, cfg.UDPMaxPort)
	logger.Info("Voice service initialized with %d STUN and %d TURN servers (public IP: %s)", len(cfg.STUNServers), len(cfg.TURNServers), cfg.PublicIP)

	metrics.StartSystemMetricsCollector()
	logger.Info("System metrics collector started")

	lokiLogger := utils.NewLokiLogger(cfg.LokiURL, map[string]string{
		"environment": cfg.Env,
	})
	logger.Info("Loki logger initialized: %s", cfg.LokiURL)

	mux := http.NewServeMux()

	
	mux.HandleFunc("/ws", handlers.HandleWebSocket(cfg))


	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/health", handlers.HandleHealth)
	apiMux.Handle("/metrics", promhttp.Handler())
	apiMux.HandleFunc("/swagger.yaml", handlers.ServeSwaggerYAML)
	apiMux.HandleFunc("/docs", handlers.ServeSwaggerUI)
	apiMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"Code Linting Platform API","version":"1.0.0","endpoints":{"/ws":"WebSocket endpoint","/health":"Health check","/docs":"API Documentation","/metrics":"Prometheus metrics"}}`)
	})

	wrappedAPIHandler := middleware.LoggingMiddleware(lokiLogger)(middleware.MetricsMiddleware(apiMux))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" {

			mux.ServeHTTP(w, r)
		} else {

			wrappedAPIHandler.ServeHTTP(w, r)
		}
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("WebSocket endpoint: ws://localhost:%s/ws", cfg.Port)
		logger.Info("Health check: http://localhost:%s/health", cfg.Port)
		logger.Info("API Documentation: http://localhost:%s/docs", cfg.Port)
		logger.Info("Prometheus metrics: http://localhost:%s/metrics", cfg.Port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server shutdown")
}
