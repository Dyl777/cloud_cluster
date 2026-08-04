package main

import (
	"fmt"
	"os"

	"github.com/gpuhub/cloud/internal/notify"
	"github.com/gpuhub/cloud/internal/shared/httpx"
)

func main() {
	port := os.Getenv("NOTIFY_PORT")
	if port == "" {
		port = "8087"
	}
	s := httpx.NewServer(":" + port)
	h := &notify.Handler{}
	h.Routes(s.Mux())
	fmt.Println("notify on " + port)
	if err := s.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
