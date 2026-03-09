package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	marketstateapi "github.com/crypto-market-copilot/alerts/services/market-state-api"
)

func main() {
	provider := marketstateapi.NewDeterministicProvider()
	server, err := marketstateapi.NewServer(serverAddress(), provider)
	if err != nil {
		log.Fatalf("create market-state-api server: %v", err)
	}
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
