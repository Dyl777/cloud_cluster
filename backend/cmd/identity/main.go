package main

import (
	"fmt"
	"os"

	"github.com/gpuhub/cloud/internal/identity"
	"github.com/gpuhub/cloud/internal/shared/httpx"
)

func main() {
	port := os.Getenv("IDENTITY_PORT")
	if port == "" {
		port = "8081"
	}
	s := httpx.NewServer(":" + port)
	h := identity.NewHandler(identity.New())
	h.Routes(s.Mux())
	fmt.Println("identity service on :" + port)
	if err := s.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
