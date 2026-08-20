package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gpuhub/cloud/internal/shared/httpx"
	"github.com/gpuhub/cloud/internal/simbridge"
)

func main() {
	port := env("SIMBRIDGE_PORT", "9090")
	carrier := strings.ToLower(env("CARRIER", "mtn"))
	phone := env("PHONE", "")
	slot := envInt("SIM_SLOT", 0)

	var sims []simbridge.SIMSlot
	if phone != "" {
		sims = append(sims, simbridge.SIMSlot{Slot: slot, Number: phone, Carrier: carrier})
	}

	s := httpx.NewServer(":" + port)
	h := simbridge.NewHandler(simbridge.New(sims))
	h.Routes(s.Mux())
	fmt.Println("simbridge on :" + port, "— local USSD bridge for phone / SIM modem")
	if phone != "" {
		fmt.Println("  SIM slot", slot, "carrier", carrier, "number", phone)
	} else {
		fmt.Println("  set PHONE (+ optional CARRIER, SIM_SLOT) to register a SIM")
	}
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

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
