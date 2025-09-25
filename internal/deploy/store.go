package deploy

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// ErrNotFound indicates the deployment was not found.
var ErrNotFound = errors.New("deployment not found")

// ErrInvalidReplicas indicates the desired replica count is invalid for mock.
var ErrInvalidReplicas = errors.New("invalid replica count")

// Summary represents deployment information shown in the table.
type Summary struct {
	Name            string   `json:"name"`
	Namespace       string   `json:"namespace"`
	ReadyReplicas   int      `json:"readyReplicas"`
	UpdatedReplicas int      `json:"updatedReplicas"`
	DesiredReplicas int      `json:"desiredReplicas"`
	Strategy        string   `json:"strategy"`
	Images          []string `json:"images"`
	Age             string   `json:"age"`
	Status          string   `json:"status"`
}

// Detail extends the summary with template metadata.
type Detail struct {
	Summary
	Labels      map[string]string `json:"labels"`
	Selector    map[string]string `json:"selector"`
	Containers  []Container       `json:"containers"`
	Conditions  []Condition       `json:"conditions"`
	Revision    int               `json:"revision"`
	LastUpdated string            `json:"lastUpdated"`
}

// Container summarises the pod template containers.
type Container struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	Ports []int  `json:"ports"`
}

// Condition models a deployment condition entry.
type Condition struct {
	Type           string `json:"type"`
	Status         string `json:"status"`
	Message        string `json:"message"`
	LastUpdate     string `json:"lastUpdate"`
	LastTransition string `json:"lastTransition"`
}

type record struct {
	Summary
	CreatedAt  time.Time
	Revision   int
	Labels     map[string]string
	Selector   map[string]string
	Containers []Container
	Conditions []conditionRecord
	LastUpdate time.Time
}

type conditionRecord struct {
	Type           string
	Status         string
	Message        string
	LastUpdate     time.Time
	LastTransition time.Time
}

// Store keeps deployment mock data in memory.
type Store struct {
	mu    sync.RWMutex
	items map[string]record
}

// NewStore seeds deployments with deterministic data.
func NewStore(now time.Time) *Store {
	s := &Store{items: make(map[string]record)}
	for _, rec := range defaultSeed(now) {
		s.items[key(rec.Namespace, rec.Name)] = rec
	}
	return s
}

// List returns deployments sorted by namespace/ name.
func (s *Store) List(now time.Time) []Summary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Summary, 0, len(s.items))
	for _, rec := range s.items {
		out = append(out, decorateSummary(rec.Summary, rec.CreatedAt, now))
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Namespace == out[j].Namespace {
			return strings.Compare(out[i].Name, out[j].Name) < 0
		}
		return strings.Compare(out[i].Namespace, out[j].Namespace) < 0
	})

	return out
}

// Get returns the deployment detail by name.
func (s *Store) Get(name string, now time.Time) (Detail, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, rec := range s.items {
		if rec.Name == name {
			return toDetail(rec, now), nil
		}
	}
	return Detail{}, ErrNotFound
}

// Scale updates the desired replicas for a deployment.
func (s *Store) Scale(name string, replicas int, now time.Time) (Detail, error) {
	if replicas < 0 || replicas > 200 {
		return Detail{}, ErrInvalidReplicas
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for key, rec := range s.items {
		if rec.Name == name {
			rec.DesiredReplicas = replicas
			if rec.ReadyReplicas > replicas {
				rec.ReadyReplicas = replicas
			}
			if rec.UpdatedReplicas > replicas {
				rec.UpdatedReplicas = replicas
			}
			rec.LastUpdate = now
			rec.Revision++
			s.items[key] = rec
			return toDetail(rec, now), nil
		}
	}

	return Detail{}, ErrNotFound
}

func decorateSummary(sum Summary, created time.Time, now time.Time) Summary {
	out := sum
	out.Age = formatAge(now.Sub(created))
	if out.ReadyReplicas == out.DesiredReplicas {
		out.Status = "Healthy"
	} else if out.ReadyReplicas == 0 {
		out.Status = "Down"
	} else {
		out.Status = "Updating"
	}
	return out
}

func toDetail(rec record, now time.Time) Detail {
	summary := decorateSummary(rec.Summary, rec.CreatedAt, now)

	labels := copyMap(rec.Labels)
	selector := copyMap(rec.Selector)

	conditions := make([]Condition, 0, len(rec.Conditions))
	for _, c := range rec.Conditions {
		conditions = append(conditions, Condition{
			Type:           c.Type,
			Status:         c.Status,
			Message:        c.Message,
			LastUpdate:     c.LastUpdate.Format(time.RFC3339),
			LastTransition: c.LastTransition.Format(time.RFC3339),
		})
	}

	return Detail{
		Summary:     summary,
		Labels:      labels,
		Selector:    selector,
		Containers:  append([]Container{}, rec.Containers...),
		Conditions:  conditions,
		Revision:    rec.Revision,
		LastUpdated: rec.LastUpdate.Format(time.RFC3339),
	}
}

func copyMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
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

func key(namespace, name string) string {
	return namespace + "/" + name
}

func defaultSeed(now time.Time) []record {
	base := now.Add(-72 * time.Hour)

	return []record{
		{
			Summary: Summary{
				Name:            "frontend",
				Namespace:       "default",
				ReadyReplicas:   4,
				UpdatedReplicas: 4,
				DesiredReplicas: 4,
				Strategy:        "RollingUpdate",
				Images:          []string{"registry.local/frontend:2.3.1"},
			},
			CreatedAt: base,
			Revision:  7,
			Labels: map[string]string{
				"app":     "frontend",
				"version": "v2",
			},
			Selector: map[string]string{
				"app": "frontend",
			},
			Containers: []Container{
				{Name: "frontend", Image: "registry.local/frontend:2.3.1", Ports: []int{80, 443}},
			},
			Conditions: []conditionRecord{
				{
					Type:           "Available",
					Status:         "True",
					Message:        "Deployment has minimum availability.",
					LastUpdate:     now.Add(-2 * time.Hour),
					LastTransition: base.Add(12 * time.Hour),
				},
			},
			LastUpdate: now.Add(-30 * time.Minute),
		},
		{
			Summary: Summary{
				Name:            "backend",
				Namespace:       "prod",
				ReadyReplicas:   5,
				UpdatedReplicas: 3,
				DesiredReplicas: 6,
				Strategy:        "RollingUpdate",
				Images:          []string{"registry.local/backend:1.12.0"},
			},
			CreatedAt: base.Add(-24 * time.Hour),
			Revision:  21,
			Labels: map[string]string{
				"app":  "backend",
				"tier": "api",
			},
			Selector: map[string]string{
				"app": "backend",
			},
			Containers: []Container{
				{Name: "api", Image: "registry.local/backend:1.12.0", Ports: []int{8080}},
			},
			Conditions: []conditionRecord{
				{
					Type:           "Progressing",
					Status:         "True",
					Message:        "ReplicaSet backend-546cdfd756 is progressing.",
					LastUpdate:     now.Add(-10 * time.Minute),
					LastTransition: now.Add(-10 * time.Minute),
				},
			},
			LastUpdate: now.Add(-5 * time.Minute),
		},
		{
			Summary: Summary{
				Name:            "batch-jobs",
				Namespace:       "batch",
				ReadyReplicas:   0,
				UpdatedReplicas: 0,
				DesiredReplicas: 2,
				Strategy:        "Recreate",
				Images:          []string{"registry.local/batch:0.5.0"},
			},
			CreatedAt: base.Add(-6 * time.Hour),
			Revision:  4,
			Labels: map[string]string{
				"app":  "batch-jobs",
				"team": "data",
			},
			Selector: map[string]string{
				"app": "batch-jobs",
			},
			Containers: []Container{
				{Name: "runner", Image: "registry.local/batch:0.5.0", Ports: []int{}},
			},
			Conditions: []conditionRecord{
				{
					Type:           "Available",
					Status:         "False",
					Message:        "Deployment does not have minimum availability.",
					LastUpdate:     now.Add(-50 * time.Minute),
					LastTransition: now.Add(-50 * time.Minute),
				},
			},
			LastUpdate: now.Add(-45 * time.Minute),
		},
	}
}
