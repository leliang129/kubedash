package pod

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// ErrNotFound indicates the pod does not exist in the store.
var ErrNotFound = errors.New("pod not found")

// Summary represents the data shown in the pods list view.
type Summary struct {
	Name            string   `json:"name"`
	Namespace       string   `json:"namespace"`
	Status          string   `json:"status"`
	ReadyContainers string   `json:"readyContainers"`
	Restarts        int      `json:"restarts"`
	Age             string   `json:"age"`
	Node            string   `json:"node"`
	Images          []string `json:"images"`
}

// Container describes a single container in the pod detail view.
type Container struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Ready        bool   `json:"ready"`
	RestartCount int    `json:"restartCount"`
}

// Event represents a pod event entry.
type Event struct {
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// Detail includes a pod summary plus container info, logs and events.
type Detail struct {
	Summary
	Containers []Container `json:"containers"`
	Logs       []string    `json:"logs"`
	Events     []Event     `json:"events"`
}

type record struct {
	Summary
	CreatedAt  time.Time
	Containers []Container
	Events     []Event
	Logs       []string
}

// Store keeps in-memory mock pod data.
type Store struct {
	mu    sync.RWMutex
	items map[string]record
}

// NewStore seeds pods with deterministic data.
func NewStore(now time.Time) *Store {
	s := &Store{items: make(map[string]record)}
	for _, rec := range defaultSeed(now) {
		s.items[key(rec.Namespace, rec.Name)] = rec
	}
	return s
}

// List returns pods sorted by namespace/name.
func (s *Store) List(now time.Time) []Summary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summaries := make([]Summary, 0, len(s.items))
	for _, rec := range s.items {
		summaries = append(summaries, decorateSummary(rec.Summary, rec.CreatedAt, now))
	}

	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Namespace == summaries[j].Namespace {
			return strings.Compare(summaries[i].Name, summaries[j].Name) < 0
		}
		return strings.Compare(summaries[i].Namespace, summaries[j].Namespace) < 0
	})

	return summaries
}

// Get fetches pod detail by name (unique across cluster for this mock).
func (s *Store) Get(name string, now time.Time) (Detail, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, rec := range s.items {
		if rec.Name == name {
			summary := decorateSummary(rec.Summary, rec.CreatedAt, now)
			detail := Detail{
				Summary:    summary,
				Containers: append([]Container{}, rec.Containers...),
				Logs:       append([]string{}, rec.Logs...),
				Events:     decorateEvents(rec.Events, now),
			}
			return detail, nil
		}
	}

	return Detail{}, ErrNotFound
}

func decorateSummary(sum Summary, createdAt, now time.Time) Summary {
	out := sum
	out.Age = formatAge(now.Sub(createdAt))
	return out
}

func decorateEvents(events []Event, now time.Time) []Event {
	out := make([]Event, 0, len(events))
	for _, ev := range events {
		item := ev
		if item.Timestamp == "" {
			item.Timestamp = now.Format(time.RFC3339)
		}
		out = append(out, item)
	}
	return out
}

func key(namespace, name string) string {
	return namespace + "/" + name
}

func formatAge(d time.Duration) string {
	if d < 0 {
		d = 0
	}

	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute

	switch {
	case days > 0:
		return fmt.Sprintf("%dd%dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh%dm", hours, minutes)
	default:
		secs := int(d/time.Second) % 60
		return fmt.Sprintf("%dm%ds", minutes, secs)
	}
}

func defaultSeed(now time.Time) []record {
	base := now.Add(-6 * time.Hour)

	return []record{
		{
			Summary: Summary{
				Name:            "frontend-7d8fdc9f7c-abc12",
				Namespace:       "default",
				Status:          "Running",
				ReadyContainers: "2/2",
				Restarts:        1,
				Age:             "",
				Node:            "node-2",
				Images:          []string{"nginx:1.25", "busybox:1.36"},
			},
			CreatedAt: base,
			Containers: []Container{
				{Name: "frontend", Image: "nginx:1.25", Ready: true, RestartCount: 1},
				{Name: "sidecar", Image: "busybox:1.36", Ready: true, RestartCount: 0},
			},
			Logs: []string{
				"[INFO] 10:15:01 request handled /",
				"[INFO] 10:15:02 request handled /healthz",
				"[WARN] 10:16:12 upstream latency 240ms",
			},
			Events: []Event{
				{Type: "Normal", Reason: "Pulled", Message: "Successfully pulled image nginx:1.25"},
				{Type: "Normal", Reason: "Started", Message: "Started container frontend"},
			},
		},
		{
			Summary: Summary{
				Name:            "backend-76c4d5f6d6-xyz89",
				Namespace:       "prod",
				Status:          "Running",
				ReadyContainers: "1/1",
				Restarts:        0,
				Age:             "",
				Node:            "node-3",
				Images:          []string{"golang:1.21"},
			},
			CreatedAt: base.Add(-2 * time.Hour),
			Containers: []Container{
				{Name: "backend", Image: "golang:1.21", Ready: true, RestartCount: 0},
			},
			Logs: []string{
				"[INFO] 09:10:04 processed job 2384",
				"[INFO] 09:12:51 processed job 2385",
				"[INFO] 09:15:13 cache warmup complete",
			},
			Events: []Event{
				{Type: "Normal", Reason: "ScalingReplicaSet", Message: "Scaled up replica set backend-76c4d5f6d6 to 3"},
			},
		},
		{
			Summary: Summary{
				Name:            "jobs-runner-bb7d67f4f6-123zt",
				Namespace:       "batch",
				Status:          "Pending",
				ReadyContainers: "0/1",
				Restarts:        0,
				Age:             "",
				Node:            "",
				Images:          []string{"python:3.12"},
			},
			CreatedAt: base.Add(-30 * time.Minute),
			Containers: []Container{
				{Name: "worker", Image: "python:3.12", Ready: false, RestartCount: 0},
			},
			Logs: []string{
				"[INFO] job queued",
			},
			Events: []Event{
				{Type: "Warning", Reason: "FailedScheduling", Message: "0/3 nodes available: insufficient memory."},
			},
		},
	}
}
