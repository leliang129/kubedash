package logs

import (
	"testing"
	"time"
)

func TestListLogsWithFilters(t *testing.T) {
	freeze := time.Date(2024, 7, 12, 10, 0, 0, 0, time.UTC)
	store := NewStore(freeze)

	logs := store.ListLogs(freeze, LogFilter{Namespace: "prod", Limit: 5})
	if len(logs) == 0 {
		t.Fatalf("expected prod logs")
	}
	for _, entry := range logs {
		if entry.Namespace != "prod" {
			t.Fatalf("unexpected namespace: %s", entry.Namespace)
		}
		if entry.Timestamp == "" {
			t.Fatalf("expected timestamp to be filled")
		}
	}

	errLogs := store.ListLogs(freeze, LogFilter{Level: "error", Limit: 10})
	if len(errLogs) == 0 {
		t.Fatalf("expected error logs")
	}
	for _, entry := range errLogs {
		if entry.Level != LevelError {
			t.Fatalf("expected error level, got %s", entry.Level)
		}
	}
}

func TestListEventsOrdering(t *testing.T) {
	freeze := time.Date(2024, 7, 12, 10, 0, 0, 0, time.UTC)
	store := NewStore(freeze)

	events := store.ListEvents(freeze)
	if len(events) == 0 {
		t.Fatalf("expected events")
	}

	for i := 1; i < len(events); i++ {
		if events[i].Timestamp > events[i-1].Timestamp {
			t.Fatalf("events not sorted descending: %s > %s", events[i].Timestamp, events[i-1].Timestamp)
		}
	}
}
