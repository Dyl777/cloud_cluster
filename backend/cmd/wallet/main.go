package main

import (
	"fmt"
	"os"

	"github.com/gpuhub/cloud/internal/shared/httpx"
	"github.com/gpuhub/cloud/internal/wallet"
)

func main() {
	port := os.Getenv("WALLET_PORT")
	if port == "" {
		port = "8082"
	}
	s := httpx.NewServer(":" + port)
	h := wallet.NewHandler(wallet.New())
	h.Routes(s.Mux())
	fmt.Println("wallet service on :" + port)
	if err := s.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
