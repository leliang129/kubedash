package cluster

import (
	"testing"
	"time"
)

func TestMockOverview(t *testing.T) {
	now := time.Date(2024, 7, 12, 15, 30, 0, 0, time.UTC)
	overview := MockOverview(now)

	if overview.Info.Name != "Mock Production Cluster" {
		t.Fatalf("unexpected cluster name: %s", overview.Info.Name)
	}

	if overview.ResourceUsage.CPU.Unit != "cores" {
		t.Errorf("expected CPU unit cores, got %s", overview.ResourceUsage.CPU.Unit)
	}

	if len(overview.RecentEvents) != 3 {
		t.Fatalf("expected 3 events, got %d", len(overview.RecentEvents))
	}

	expectedTimestamp := "2024-07-12T15:30:00Z"
	if overview.ResourceUsage.Memory.Timestamp != expectedTimestamp {
		t.Errorf("memory timestamp mismatch, got %s", overview.ResourceUsage.Memory.Timestamp)
	}

	if overview.RecentEvents[0].Timestamp != expectedTimestamp {
		t.Errorf("event timestamp mismatch, got %s", overview.RecentEvents[0].Timestamp)
	}
}
