package main

import (
	"fmt"
	"os"

	"github.com/gpuhub/cloud/internal/marketplace"
	"github.com/gpuhub/cloud/internal/shared/httpx"
)

func main() {
	port := os.Getenv("MARKETPLACE_PORT")
	if port == "" {
		port = "8084"
	}
	reg := marketplace.NewRegistry(
		marketplace.NewMockProvider("aws"),
		marketplace.NewMockProvider("azure"),
		marketplace.NewMockProvider("gcp"),
		marketplace.NewMockProvider("runpod"),
	)
	s := httpx.NewServer(":" + port)
	h := marketplace.NewHandler(reg)
	h.Routes(s.Mux())
	fmt.Println("marketplace service on :" + port)
	if err := s.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
