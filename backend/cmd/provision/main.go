package main

import (
	"fmt"
	"os"

	"github.com/gpuhub/cloud/internal/provision"
	"github.com/gpuhub/cloud/internal/shared/httpx"
)

func main() {
	port := os.Getenv("PROVISION_PORT")
	if port == "" {
		port = "8085"
	}
	s := httpx.NewServer(":" + port)
	h := provision.NewHandler(provision.New())
	h.Routes(s.Mux())
	fmt.Println("provision on " + port)
	if err := s.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
