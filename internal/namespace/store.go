package namespace

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	// ErrExists indicates namespace already present.
	ErrExists = errors.New("namespace already exists")
	// ErrInvalidName signals the provided name violates Kubernetes naming rules.
	ErrInvalidName = errors.New("invalid namespace name")
)

var namespaceNameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// Namespace represents the JSON payload returned to the frontend.
type Namespace struct {
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	Age       string            `json:"age"`
	CreatedAt string            `json:"createdAt"`
	Labels    map[string]string `json:"labels,omitempty"`
}

type record struct {
	Name      string
	Status    string
	CreatedAt time.Time
	Labels    map[string]string
}

// Store keeps in-memory namespace state for the mock API.
type Store struct {
	mu    sync.RWMutex
	items map[string]record
}

// NewStore seeds a namespace store with deterministic mock data.
func NewStore(now time.Time) *Store {
	s := &Store{
		items: make(map[string]record),
	}

	for _, rec := range defaultSeed(now) {
		s.items[rec.Name] = rec
	}

	return s
}

// List returns all namespaces sorted alphabetically.
func (s *Store) List(now time.Time) []Namespace {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Namespace, 0, len(s.items))
	for _, rec := range s.items {
		out = append(out, toNamespace(rec, now))
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})

	return out
}

// Create inserts a new namespace if it does not yet exist.
func (s *Store) Create(name string, now time.Time, labels map[string]string) (Namespace, error) {
	clean := strings.TrimSpace(name)
	if clean == "" || len(clean) > 63 || !namespaceNameRegex.MatchString(clean) {
		return Namespace{}, ErrInvalidName
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.items[clean]; exists {
		return Namespace{}, ErrExists
	}

	createdAt := now.UTC()
	rec := record{
		Name:      clean,
		Status:    "Active",
		CreatedAt: createdAt,
		Labels:    mergeLabels(clean, labels),
	}
	s.items[clean] = rec
	return toNamespace(rec, now), nil
}

// Delete removes the namespace if it exists.
func (s *Store) Delete(name string) bool {
	clean := strings.TrimSpace(name)
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.items[clean]; !exists {
		return false
	}

	delete(s.items, clean)
	return true
}

func toNamespace(rec record, now time.Time) Namespace {
	age := formatAge(now.Sub(rec.CreatedAt))
	return Namespace{
		Name:      rec.Name,
		Status:    rec.Status,
		Age:       age,
		CreatedAt: rec.CreatedAt.Format(time.RFC3339),
		Labels:    rec.Labels,
	}
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

func mergeLabels(name string, custom map[string]string) map[string]string {
	labels := map[string]string{
		"kubernetes.io/metadata.name":  name,
		"app.kubernetes.io/managed-by": "mock-dashboard",
	}

	for k, v := range custom {
		labels[k] = v
	}

	return labels
}

func defaultSeed(now time.Time) []record {
	base := now.Add(-48 * time.Hour)
	return []record{
		{
			Name:      "default",
			Status:    "Active",
			CreatedAt: base,
			Labels: map[string]string{
				"kubernetes.io/metadata.name": "default",
			},
		},
		{
			Name:      "kube-system",
			Status:    "Active",
			CreatedAt: base.Add(-24 * time.Hour),
			Labels: map[string]string{
				"kubernetes.io/metadata.name":        "kube-system",
				"pod-security.kubernetes.io/enforce": "privileged",
			},
		},
		{
			Name:      "monitoring",
			Status:    "Active",
			CreatedAt: base.Add(6 * time.Hour),
			Labels: map[string]string{
				"team":                        "sre",
				"kubernetes.io/metadata.name": "monitoring",
			},
		},
	}
}
