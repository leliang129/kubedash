package node

import (
	"testing"
	"time"
)

func TestListOrderingAndMetrics(t *testing.T) {
	now := time.Date(2024, 7, 12, 12, 0, 0, 0, time.UTC)
	store := NewStore(now)

	summaries := store.List(now)
	if len(summaries) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(summaries))
	}

	if summaries[0].Name != "node-1" {
		t.Fatalf("expected node-1 first, got %s", summaries[0].Name)
	}

	if summaries[0].CPU.Unit != "cores" {
		t.Fatalf("unexpected cpu unit %s", summaries[0].CPU.Unit)
	}

	if summaries[1].Pods.Capacity <= 0 {
		t.Fatalf("expected positive pod capacity")
	}
}

func TestGetDetail(t *testing.T) {
	now := time.Now()
	store := NewStore(now)

	detail, err := store.Get("node-2", now)
	if err != nil {
		t.Fatalf("get node detail: %v", err)
	}

	if detail.Name != "node-2" {
		t.Fatalf("unexpected name %s", detail.Name)
	}

	if len(detail.Conditions) == 0 {
		t.Fatalf("expected conditions to be populated")
	}

	if detail.Conditions[0].LastHeartbeat == "" {
		t.Fatalf("expected heartbeat timestamp")
	}

	if _, err := store.Get("missing", now); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
