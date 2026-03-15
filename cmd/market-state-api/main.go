package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	marketstateapi "github.com/crypto-market-copilot/alerts/services/market-state-api"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	provider, err := newProvider()
	if err != nil {
		log.Fatalf("create market-state-api provider: %v", err)
	}
	if closer, ok := provider.(interface{ Close(context.Context) error }); ok {
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := closer.Close(shutdownCtx); err != nil {
				log.Printf("stop market-state-api provider: %v", err)
			}
		}()
	}
	server, err := marketstateapi.NewServer(serverAddress(), provider)
	if err != nil {
		log.Fatalf("create market-state-api server: %v", err)
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("shutdown market-state-api server: %v", err)
		}
	}()
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve market-state-api: %v", err)
	}
}

func serverAddress() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}
