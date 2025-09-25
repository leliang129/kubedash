package deploy

import (
	"testing"
	"time"
)

func TestListOrdering(t *testing.T) {
	now := time.Date(2024, 7, 12, 12, 0, 0, 0, time.UTC)
	store := NewStore(now)

	deployments := store.List(now)
	if len(deployments) != 3 {
		t.Fatalf("expected 3 deployments, got %d", len(deployments))
	}

	if deployments[0].Namespace != "batch" {
		t.Fatalf("expected batch namespace first, got %s", deployments[0].Namespace)
	}

	if deployments[0].Age == "" {
		t.Fatalf("expected age to be populated")
	}
}

func TestGetAndScale(t *testing.T) {
	now := time.Now()
	store := NewStore(now)

	detail, err := store.Get("frontend", now)
	if err != nil {
		t.Fatalf("get deployment detail: %v", err)
	}

	if detail.DesiredReplicas != 4 {
		t.Fatalf("unexpected desired replicas %d", detail.DesiredReplicas)
	}

	scaled, err := store.Scale("frontend", 6, now.Add(1*time.Minute))
	if err != nil {
		t.Fatalf("scale deployment: %v", err)
	}

	if scaled.DesiredReplicas != 6 {
		t.Fatalf("expected desired replicas 6, got %d", scaled.DesiredReplicas)
	}

	if _, err := store.Scale("frontend", -1, now); err != ErrInvalidReplicas {
		t.Fatalf("expected ErrInvalidReplicas, got %v", err)
	}

	if _, err := store.Get("missing", now); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if _, err := store.Scale("missing", 2, now); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound on scale, got %v", err)
	}
}
