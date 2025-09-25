package namespace

import (
	"testing"
	"time"
)

func TestStoreListOrdering(t *testing.T) {
	now := time.Date(2024, 7, 12, 12, 0, 0, 0, time.UTC)
	store := NewStore(now)

	namespaces := store.List(now)
	if len(namespaces) != 3 {
		t.Fatalf("expected 3 namespaces, got %d", len(namespaces))
	}

	if namespaces[0].Name != "default" {
		t.Fatalf("expected default first, got %s", namespaces[0].Name)
	}

	if namespaces[0].Age == "" {
		t.Fatal("expected age to be populated")
	}
}

func TestStoreCreateAndDelete(t *testing.T) {
	now := time.Date(2024, 7, 12, 12, 0, 0, 0, time.UTC)
	store := NewStore(now)

	ns, err := store.Create("staging", now, map[string]string{"team": "platform"})
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}

	if ns.Name != "staging" {
		t.Fatalf("unexpected namespace name %s", ns.Name)
	}

	if ns.Labels["team"] != "platform" {
		t.Fatalf("expected custom label to be preserved")
	}

	if _, err := store.Create("staging", now, nil); err != ErrExists {
		t.Fatalf("expected ErrExists, got %v", err)
	}

	if !store.Delete("staging") {
		t.Fatalf("expected delete to succeed")
	}

	if store.Delete("staging") {
		t.Fatalf("expected delete to fail for missing namespace")
	}
}

func TestStoreCreateInvalidName(t *testing.T) {
	now := time.Now()
	store := NewStore(now)

	if _, err := store.Create("Invalid_Name", now, nil); err != ErrInvalidName {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}
}
