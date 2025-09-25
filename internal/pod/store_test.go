package pod

import (
	"testing"
	"time"
)

func TestListOrdering(t *testing.T) {
	now := time.Date(2024, 7, 12, 12, 0, 0, 0, time.UTC)
	store := NewStore(now)

	pods := store.List(now)
	if len(pods) != 3 {
		t.Fatalf("expected 3 pods, got %d", len(pods))
	}

	if pods[0].Namespace != "batch" {
		t.Fatalf("expected batch namespace first, got %s", pods[0].Namespace)
	}

	if pods[0].Age == "" {
		t.Fatalf("expected age to be populated")
	}
}

func TestGetDetail(t *testing.T) {
	now := time.Now()
	store := NewStore(now)

	detail, err := store.Get("frontend-7d8fdc9f7c-abc12", now)
	if err != nil {
		t.Fatalf("get pod detail: %v", err)
	}

	if detail.Namespace != "default" {
		t.Fatalf("unexpected namespace %s", detail.Namespace)
	}

	if len(detail.Containers) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(detail.Containers))
	}

	if len(detail.Logs) == 0 {
		t.Fatalf("expected logs to be populated")
	}

	if _, err := store.Get("missing", now); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
