package provision

import "testing"

func TestLifecycleMemoryStore(t *testing.T) {
	svc := New()
	inst := svc.Create("i1", "u1", "RTX 4090", 1, "mock")
	if inst.Status != "running" || inst.Provider != "mock" {
		t.Fatalf("created = %+v", inst)
	}

	launched := svc.Launch(Instance{ID: "i2", UserID: "u1", GPUName: "H100", NumGPUs: 2})
	if launched.Status != "running" {
		t.Fatalf("launch status = %q", launched.Status)
	}

	if _, ok := svc.Get("i1"); !ok {
		t.Fatal("expected i1 present")
	}
	if _, ok := svc.Get("missing"); ok {
		t.Fatal("missing instance reported present")
	}

	if up, ok := svc.Update("i1", func(i *Instance) { i.Status = "stopped" }); !ok || up.Status != "stopped" {
		t.Fatalf("update = %+v ok=%v", up, ok)
	}
	if _, ok := svc.Update("missing", func(i *Instance) {}); ok {
		t.Fatal("update of missing instance reported ok")
	}

	if !svc.Destroy("i1") {
		t.Fatal("first destroy should report true")
	}
	if svc.Destroy("i1") {
		t.Fatal("second destroy should report false")
	}
	if _, ok := svc.Get("i1"); ok {
		t.Fatal("destroyed instance still present")
	}
}

func TestListFiltersByUser(t *testing.T) {
	svc := New()
	svc.Create("a", "u1", "H100", 1, "mock")
	svc.Create("b", "u2", "A100", 1, "mock")
	if l := svc.List("u1"); len(l) != 1 || l[0].ID != "a" {
		t.Fatalf("u1 list = %+v", l)
	}
	if l := svc.List(""); len(l) != 2 {
		t.Fatalf("all list len = %d, want 2", len(l))
	}
}