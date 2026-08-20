package main

import (
	"fmt"
	"os"

	"github.com/gpuhub/cloud/internal/payments"
	"github.com/gpuhub/cloud/internal/shared/httpx"
)

func main() {
	port := os.Getenv("PAYMENTS_PORT")
	if port == "" {
		port = "8083"
	}
	walletURL := os.Getenv("WALLET_URL")
	if walletURL == "" {
		walletURL = "http://localhost:8082"
	}
	paymgrURL := os.Getenv("PAYMGR_URL")
	if paymgrURL == "" {
		paymgrURL = "http://localhost:8089"
	}
	s := httpx.NewServer(":" + port)
	h := payments.NewHandler(payments.NewOrchestrator(walletURL, paymgrURL))
	h.Routes(s.Mux())
	fmt.Println("payments service on :" + port)
	if err := s.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
