// Command compat exposes a REST-compatibility gateway mirroring the
// console.vast.ai /api/v0 surface so unmodified official Vast tools
// (vast-cli, vast-sdk, skypilot-vast, vast-pyworker) talk to this platform.
package main

import (
	"fmt"
	"os"

	"github.com/gpuhub/cloud/internal/compat"
	"github.com/gpuhub/cloud/internal/marketplace"
	"github.com/gpuhub/cloud/internal/provision"
	"github.com/gpuhub/cloud/internal/settlement"
	"github.com/gpuhub/cloud/internal/shared/db"
	"github.com/gpuhub/cloud/internal/shared/httpx"
	"github.com/gpuhub/cloud/internal/wallet"
)

func main() {
	port := os.Getenv("COMPAT_PORT")
	if port == "" {
		port = "8092"
	}
	apiKey := os.Getenv("COMPAT_API_KEY")
	if apiKey == "" {
		apiKey = "demo-api-key"
	}
	userID := os.Getenv("COMPAT_USER_ID")
	if userID == "" {
		userID = "user-demo"
	}

	bus := settlement.New()
	var wal *wallet.Service
	if url := os.Getenv("DATABASE_URL"); url != "" {
		sqlDB, err := db.Connect(url)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := db.MigrateAll(url); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		wal = wallet.NewPG(wallet.NewPGStore(sqlDB))
	} else {
		wal = wallet.New()
	}
	reg := marketplace.NewRegistry(
		marketplace.NewMockProvider("mock"),
	)
	prov := provision.New()
	_ = bus

	srv := compat.NewServer(apiKey, userID, reg, prov, wal)

	base := httpx.NewServer(":" + port)
	mux := base.Mux()
	mux.Handle("/", srv) // compat ServeHTTP runs auth then routes /api/v0/*

	fmt.Printf("compat gateway (vast /api/v0) on :%s  api_key=%s\n", port, apiKey)
	if err := base.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
