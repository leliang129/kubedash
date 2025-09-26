package service

import (
	"testing"
	"time"
)

func TestStoreListDeterministicOrder(t *testing.T) {
	freeze := time.Date(2024, 7, 12, 10, 0, 0, 0, time.UTC)
	store := NewStore(freeze)

	items := store.List(freeze)

	if len(items) != 3 {
		t.Fatalf("expected 3 services, got %d", len(items))
	}

	if items[0].Name != "batch-metrics" || items[0].Namespace != "batch" {
		t.Fatalf("expected batch-metrics first, got %s/%s", items[0].Namespace, items[0].Name)
	}

	if items[1].Name != "frontend" || items[1].Namespace != "default" {
		t.Fatalf("expected frontend second, got %s/%s", items[1].Namespace, items[1].Name)
	}

	if items[2].Name != "edge-gateway" || items[2].Namespace != "prod" {
		t.Fatalf("expected edge-gateway third, got %s/%s", items[2].Namespace, items[2].Name)
	}

	for _, item := range items {
		if item.Age == "" {
			t.Fatalf("service %s age was empty", item.Name)
		}
	}
}

func TestStoreGetDetail(t *testing.T) {
	freeze := time.Date(2024, 4, 20, 9, 30, 0, 0, time.UTC)
	store := NewStore(freeze)

	detail, err := store.Get("edge-gateway", freeze)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if detail.Name != "edge-gateway" {
		t.Fatalf("expected edge-gateway, got %s", detail.Name)
	}

	if detail.Status != "Pending" {
		t.Fatalf("unexpected status %s", detail.Status)
	}

	if detail.Selector["app"] != "edge-gateway" {
		t.Fatalf("selector app mismatch: %v", detail.Selector["app"])
	}

	if len(detail.Endpoints) == 0 {
		t.Fatalf("expected endpoints")
	}

	if detail.CreatedAt != freeze.Add(-36*time.Hour).Format(time.RFC3339) {
		t.Fatalf("unexpected createdAt: %s", detail.CreatedAt)
	}

	if detail.Age != "1d12h" {
		t.Fatalf("unexpected age: %s", detail.Age)
	}

	if _, err := store.Get("ghost", freeze); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
