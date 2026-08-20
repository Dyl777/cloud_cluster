package main

import (
	"fmt"
	"os"

	"github.com/gpuhub/cloud/internal/provision"
	"github.com/gpuhub/cloud/internal/shared/db"
	"github.com/gpuhub/cloud/internal/shared/httpx"
)

func main() {
	port := os.Getenv("PROVISION_PORT")
	if port == "" {
		port = "8085"
	}

	var svc *provision.Service
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
		svc = provision.NewPG(provision.NewPGStore(sqlDB))
	} else {
		svc = provision.New()
	}

	s := httpx.NewServer(":" + port)
	h := provision.NewHandler(svc)
	h.Routes(s.Mux())
	fmt.Println("provision on " + port)
	if err := s.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}