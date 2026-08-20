package provision

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gpuhub/cloud/internal/shared/db"
)

// TestPostgresInstanceStore exercises the pg Store against a live Postgres.
// Skipped unless DATABASE_URL (schema owned by this platform) is set.
func TestPostgresInstanceStore(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set")
	}
	sqlDB, err := db.Connect(url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() }) // registered first, runs last
	if err := db.MigrateAll(url); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	svc := NewPG(NewPGStore(sqlDB))
	uid := fmt.Sprintf("pg-provision-%d", time.Now().UnixNano())
	id := "i-" + uid
	t.Cleanup(func() {
		_, _ = sqlDB.Exec("DELETE FROM provision_instance WHERE id = $1", id)
	})

	svc.Launch(Instance{ID: id, UserID: uid, GPUName: "H100", NumGPUs: 4, Status: "running", Price: 2.4})

	// A fresh Service over the same database must see the persisted instance.
	again := NewPG(NewPGStore(sqlDB))
	got, ok := again.Get(id)
	if !ok {
		t.Fatalf("reloaded instance not found")
	}
	if got.GPUName != "H100" || got.NumGPUs != 4 || got.Price != 2.4 {
		t.Fatalf("reloaded = %+v", got)
	}
	if _, ok := again.Update(id, func(i *Instance) { i.Status = "stopped" }); !ok {
		t.Fatalf("update failed")
	}
	if l := again.List(uid); len(l) != 1 || l[0].Status != "stopped" {
		t.Fatalf("list = %+v", l)
	}
	if !again.Destroy(id) {
		t.Fatalf("destroy failed")
	}
	if _, ok := again.Get(id); ok {
		t.Fatalf("destroyed instance still present")
	}
}