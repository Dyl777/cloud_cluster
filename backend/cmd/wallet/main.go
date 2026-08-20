package main

import (
	"fmt"
	"os"

	"github.com/gpuhub/cloud/internal/shared/db"
	"github.com/gpuhub/cloud/internal/shared/httpx"
	"github.com/gpuhub/cloud/internal/wallet"
)

func main() {
	port := os.Getenv("WALLET_PORT")
	if port == "" {
		port = "8082"
	}

	var svc *wallet.Service
	if url := os.Getenv("DATABASE_URL"); url != "" {
		sqlDB, err := db.Connect(url)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := db.MigrateAll(sqlDB); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		svc = wallet.NewPG(wallet.NewPGStore(sqlDB))
	} else {
		svc = wallet.New()
	}

	s := httpx.NewServer(":" + port)
	h := wallet.NewHandler(svc)
	h.Routes(s.Mux())
	fmt.Println("wallet service on :" + port)
	if err := s.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}