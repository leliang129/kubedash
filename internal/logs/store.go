package logs

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Level represents the severity of a log entry.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// LogEntry is returned via the logs API for UI consumption.
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Level     Level  `json:"level"`
	Message   string `json:"message"`
}

// Event represents a cluster event in the timeline.
type Event struct {
	Timestamp string `json:"timestamp"`
	Namespace string `json:"namespace"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	Count     int    `json:"count"`
}

// LogFilter narrows down the log entries returned from the store.
type LogFilter struct {
	Namespace string
	Pod       string
	Level     string
	Limit     int
}

type logRecord struct {
	Namespace string
	Pod       string
	Level     Level
	Message   string
	CreatedAt time.Time
}

type eventRecord struct {
	Namespace string
	Kind      string
	Name      string
	Type      string
	Reason    string
	Message   string
	Count     int
	Occurred  time.Time
}

// Store contains mock log lines and events with thread-safety.
type Store struct {
	mu     sync.RWMutex
	logs   []logRecord
	events []eventRecord
}

// NewStore seeds the store with deterministic diagnostic data.
func NewStore(now time.Time) *Store {
	s := &Store{}
	s.logs = defaultLogs(now)
	s.events = defaultEvents(now)
	return s
}

// ListLogs returns log entries sorted by recency with optional filtering.
func (s *Store) ListLogs(now time.Time, filter LogFilter) []LogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	namespace := strings.TrimSpace(strings.ToLower(filter.Namespace))
	pod := strings.TrimSpace(strings.ToLower(filter.Pod))
	level := strings.TrimSpace(strings.ToUpper(filter.Level))

	result := make([]LogEntry, 0, limit)

	for _, rec := range s.logs {
		if namespace != "" && strings.ToLower(rec.Namespace) != namespace {
			continue
		}
		if pod != "" && strings.ToLower(rec.Pod) != pod {
			continue
		}
		if level != "" && string(rec.Level) != level {
			continue
		}

		entry := LogEntry{
			Timestamp: rec.CreatedAt.Format(time.RFC3339),
			Namespace: rec.Namespace,
			Pod:       rec.Pod,
			Level:     rec.Level,
			Message:   rec.Message,
		}
		result = append(result, entry)
		if len(result) >= limit {
			break
		}
	}

	return result
}

// ListEvents returns cluster events ordered by most recent first.
func (s *Store) ListEvents(now time.Time) []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]Event, 0, len(s.events))
	for _, rec := range s.events {
		events = append(events, Event{
			Timestamp: rec.Occurred.Format(time.RFC3339),
			Namespace: rec.Namespace,
			Kind:      rec.Kind,
			Name:      rec.Name,
			Type:      rec.Type,
			Reason:    rec.Reason,
			Message:   rec.Message,
			Count:     rec.Count,
		})
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp > events[j].Timestamp
	})

	return events
}

func defaultLogs(now time.Time) []logRecord {
	base := now.Add(-5 * time.Minute)

	return []logRecord{
		{Namespace: "default", Pod: "frontend-7d8fdc9f7c-abc12", Level: LevelInfo, Message: "GET / 200 18ms", CreatedAt: base},
		{Namespace: "default", Pod: "frontend-7d8fdc9f7c-abc12", Level: LevelInfo, Message: "GET /healthz 200 5ms", CreatedAt: base.Add(20 * time.Second)},
		{Namespace: "default", Pod: "frontend-7d8fdc9f7c-def34", Level: LevelWarn, Message: "upstream latency 430ms", CreatedAt: base.Add(38 * time.Second)},
		{Namespace: "prod", Pod: "edge-gateway-7d8fdc9f7c-9012a", Level: LevelError, Message: "listener restart due to config reload", CreatedAt: base.Add(64 * time.Second)},
		{Namespace: "prod", Pod: "edge-gateway-7d8fdc9f7c-9012a", Level: LevelInfo, Message: "probe /ready succeeded", CreatedAt: base.Add(80 * time.Second)},
		{Namespace: "prod", Pod: "backend-76c4d5f6d6-xyz89", Level: LevelInfo, Message: "processed job 3498", CreatedAt: base.Add(95 * time.Second)},
		{Namespace: "prod", Pod: "backend-76c4d5f6d6-xyz89", Level: LevelWarn, Message: "cache miss ratio 42% exceeds threshold", CreatedAt: base.Add(2 * time.Minute)},
		{Namespace: "batch", Pod: "batch-runner-5b87d7fbc6-kx912", Level: LevelInfo, Message: "scheduled job nightly-sync", CreatedAt: base.Add(150 * time.Second)},
		{Namespace: "batch", Pod: "batch-runner-5b87d7fbc6-kx912", Level: LevelError, Message: "job nightly-sync failed: no nodes available", CreatedAt: base.Add(165 * time.Second)},
		{Namespace: "batch", Pod: "batch-runner-5b87d7fbc6-kx912", Level: LevelInfo, Message: "retry scheduled in 30s", CreatedAt: base.Add(3 * time.Minute)},
	}
}

func defaultEvents(now time.Time) []eventRecord {
	return []eventRecord{
		{
			Namespace: "default",
			Kind:      "Deployment",
			Name:      "frontend",
			Type:      "Normal",
			Reason:    "ScalingReplicaSet",
			Message:   "Scaled up replica set frontend-7d8fdc9f7c to 3",
			Count:     1,
			Occurred:  now.Add(-12 * time.Minute),
		},
		{
			Namespace: "prod",
			Kind:      "Pod",
			Name:      "edge-gateway-7d8fdc9f7c-9012a",
			Type:      "Warning",
			Reason:    "FailedScheduling",
			Message:   "0/3 nodes available: Insufficient MEMORY",
			Count:     3,
			Occurred:  now.Add(-9 * time.Minute),
		},
		{
			Namespace: "prod",
			Kind:      "Pod",
			Name:      "edge-gateway-7d8fdc9f7c-9012b",
			Type:      "Normal",
			Reason:    "Pulled",
			Message:   "Successfully pulled image registry.local/edge:2.1.0",
			Count:     1,
			Occurred:  now.Add(-8 * time.Minute),
		},
		{
			Namespace: "batch",
			Kind:      "Job",
			Name:      "nightly-sync",
			Type:      "Warning",
			Reason:    "BackoffLimitExceeded",
			Message:   "Job has reached the specified backoff limit",
			Count:     1,
			Occurred:  now.Add(-5 * time.Minute),
		},
		{
			Namespace: "batch",
			Kind:      "Pod",
			Name:      "batch-runner-5b87d7fbc6-kx912",
			Type:      "Normal",
			Reason:    "Scheduled",
			Message:   "Successfully assigned batch/batch-runner-5b87d7fbc6-kx912 to node-2",
			Count:     1,
			Occurred:  now.Add(-3 * time.Minute),
		},
	}
}

// AppendLog adds a new log line to the store ensuring recency ordering.
func (s *Store) AppendLog(entry LogEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec := logRecord{
		Namespace: entry.Namespace,
		Pod:       entry.Pod,
		Level:     Level(strings.ToUpper(string(entry.Level))),
		Message:   entry.Message,
		CreatedAt: parseTimestamp(entry.Timestamp, time.Now()),
	}
	s.logs = append([]logRecord{rec}, s.logs...)
}

func parseTimestamp(ts string, fallback time.Time) time.Time {
	if ts == "" {
		return fallback
	}
	if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
		return parsed
	}
	return fallback
}

// AppendEvent adds a new event to the store.
func (s *Store) AppendEvent(ev Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec := eventRecord{
		Namespace: ev.Namespace,
		Kind:      ev.Kind,
		Name:      ev.Name,
		Type:      ev.Type,
		Reason:    ev.Reason,
		Message:   ev.Message,
		Count:     ev.Count,
		Occurred:  parseTimestamp(ev.Timestamp, time.Now()),
	}
	s.events = append([]eventRecord{rec}, s.events...)
}

// UniqueNamespaces returns the distinct namespaces used in either logs or events.
func (s *Store) UniqueNamespaces() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	set := make(map[string]struct{})
	for _, rec := range s.logs {
		set[rec.Namespace] = struct{}{}
	}
	for _, rec := range s.events {
		set[rec.Namespace] = struct{}{}
	}

	namespaces := make([]string, 0, len(set))
	for ns := range set {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)
	return namespaces
}

// UniquePods returns the distinct pod names present in the logs.
func (s *Store) UniquePods() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	set := make(map[string]struct{})
	for _, rec := range s.logs {
		set[rec.Pod] = struct{}{}
	}
	pods := make([]string, 0, len(set))
	for pod := range set {
		pods = append(pods, pod)
	}
	sort.Strings(pods)
	return pods
}

// UniqueLevels returns the available log levels.
func (s *Store) UniqueLevels() []Level {
	return []Level{LevelInfo, LevelWarn, LevelError}
}

// DescribeFilters returns contextual information for filter dropdowns.
func (s *Store) DescribeFilters() map[string]any {
	return map[string]any{
		"namespaces": s.UniqueNamespaces(),
		"pods":       s.UniquePods(),
		"levels":     s.UniqueLevels(),
	}
}

// Summarize returns a concise overview for status widgets.
func (s *Store) Summarize(now time.Time) map[string]any {
	logs := s.ListLogs(now, LogFilter{Limit: 10})
	events := s.ListEvents(now)
	return map[string]any{
		"recent": len(logs),
		"events": len(events),
	}
}

// FormatRelative renders a human readable relative duration.
func FormatRelative(now, target time.Time) string {
	if target.After(now) {
		target = now
	}
	span := now.Sub(target)

	minutes := int(span / time.Minute)
	hours := minutes / 60
	days := hours / 24

	switch {
	case days > 0:
		return fmt.Sprintf("%dd%dh", days, hours%24)
	case hours > 0:
		return fmt.Sprintf("%dh%dm", hours, minutes%60)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}
