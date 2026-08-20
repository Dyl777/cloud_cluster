package main

import (
	"fmt"
	"os"

	"github.com/gpuhub/cloud/internal/notify"
	"github.com/gpuhub/cloud/internal/shared/events"
	"github.com/gpuhub/cloud/internal/shared/httpx"
)

func main() {
	port := os.Getenv("NOTIFY_PORT")
	if port == "" {
		port = "8087"
	}
	h := &notify.Handler{}

	bus := events.NewFromEnv("notify")
	for _, topic := range events.AllTopics {
		bus.Subscribe(topic, h.Handle)
	}
	defer bus.Close()

	s := httpx.NewServer(":" + port)
	h.Routes(s.Mux())
	fmt.Printf("notify on %s (bus: %T)\n", port, bus)
	if err := s.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}