package main

import (
	"fmt"
	"os"

	"github.com/gpuhub/cloud/internal/mobilegateway"
	"github.com/gpuhub/cloud/internal/shared/httpx"
)

func main() {
	port := env("MOBILE_GATEWAY_PORT", "8091")
	token := os.Getenv("GATEWAY_TOKEN")
	s := httpx.NewServer(":" + port)
	h := mobilegateway.NewHandler(mobilegateway.New(token))
	h.Routes(s.Mux())
	fmt.Println("mobilegateway on :" + port)
	if err := s.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
