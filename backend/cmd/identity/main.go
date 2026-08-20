package main

import (
	"fmt"
	"os"

	"github.com/gpuhub/cloud/internal/identity"
	"github.com/gpuhub/cloud/internal/shared/db"
	"github.com/gpuhub/cloud/internal/shared/httpx"
)

func main() {
	port := os.Getenv("IDENTITY_PORT")
	if port == "" {
		port = "8081"
	}

	var svc *identity.Service
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
		svc = identity.NewPG(identity.NewPGUserStore(sqlDB))
	} else {
		svc = identity.New()
	}
	svc.Seed("usr-superadmin", "admin@gpuhub.dev", "Platform Superadmin")

	s := httpx.NewServer(":" + port)
	h := identity.NewHandler(svc)
	h.Routes(s.Mux())
	fmt.Println("identity service on :" + port)
	if err := s.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}