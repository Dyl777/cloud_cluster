package main

import (
	"fmt"
	"os"

	"github.com/gpuhub/cloud/internal/paymgr"
	"github.com/gpuhub/cloud/internal/shared/httpx"
)

func main() {
	port := env("PAYMGR_PORT", "8089")
	gatewayURL := env("MOBILE_GATEWAY_URL", "http://localhost:8091")
	vmURL := env("MOBILE_VM_URL", "http://localhost:8090")

	s := httpx.NewServer(":" + port)
	h := paymgr.NewHandler(paymgr.New(gatewayURL, vmURL))
	h.Routes(s.Mux())
	fmt.Println("paymgr on :" + port)
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
