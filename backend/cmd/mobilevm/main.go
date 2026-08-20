package main

import (
	"fmt"
	"os"

	"github.com/gpuhub/cloud/internal/mobilevm"
	"github.com/gpuhub/cloud/internal/shared/httpx"
)

func main() {
	port := os.Getenv("MOBILE_VM_PORT")
	if port == "" {
		port = "8090"
	}
	s := httpx.NewServer(":" + port)
	h := mobilevm.NewHandler(mobilevm.New())
	h.Routes(s.Mux())
	fmt.Println("mobilevm on :" + port)
	if err := s.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
