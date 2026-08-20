package main

import (
	"fmt"
	"os"

	"github.com/gpuhub/cloud/internal/settlement"
	"github.com/gpuhub/cloud/internal/shared/events"
	"github.com/gpuhub/cloud/internal/shared/httpx"
)

func main() {
	port := os.Getenv("SETTLEMENT_PORT")
	if port == "" {
		port = "8086"
	}
	s := httpx.NewServer(":" + port)
	svc := settlement.New()
	svc.SetBus(events.NewFromEnv("settlement"))
	h := settlement.NewHandler(svc)
	h.Routes(s.Mux())
	fmt.Println("settlement on " + port)
	if err := s.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
