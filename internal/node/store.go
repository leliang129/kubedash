package node

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// ErrNotFound indicates the node does not exist in the mock store.
var ErrNotFound = errors.New("node not found")

// UsageMetric describes resource consumption relative to capacity.
type UsageMetric struct {
	Used       float64 `json:"used"`
	Capacity   float64 `json:"capacity"`
	Unit       string  `json:"unit"`
	Percentage float64 `json:"percentage"`
}

// PodSummary reports pod allocation for a node.
type PodSummary struct {
	Running  int `json:"running"`
	Pending  int `json:"pending"`
	Capacity int `json:"capacity"`
}

// Condition represents the status of a node subsystem.
type Condition struct {
	Type           string `json:"type"`
	Status         string `json:"status"`
	Message        string `json:"message"`
	LastHeartbeat  string `json:"lastHeartbeat"`
	LastTransition string `json:"lastTransition"`
}

// NodeSummary powers the nodes list view.
type NodeSummary struct {
	Name           string      `json:"name"`
	Status         string      `json:"status"`
	Roles          []string    `json:"roles"`
	Age            string      `json:"age"`
	KubeletVersion string      `json:"kubeletVersion"`
	CPU            UsageMetric `json:"cpu"`
	Memory         UsageMetric `json:"memory"`
	Pods           PodSummary  `json:"pods"`
}

// NodeDetail extends NodeSummary with additional metadata.
type NodeDetail struct {
	NodeSummary
	Architecture     string            `json:"architecture"`
	OSImage          string            `json:"osImage"`
	KernelVersion    string            `json:"kernelVersion"`
	ContainerRuntime string            `json:"containerRuntime"`
	Labels           map[string]string `json:"labels"`
	Taints           []string          `json:"taints"`
	Conditions       []Condition       `json:"conditions"`
}

type record struct {
	Name             string
	Status           string
	Roles            []string
	CreatedAt        time.Time
	KubeletVersion   string
	CPUUsed          float64
	CPUCapacity      float64
	MemoryUsed       float64
	MemoryCapacity   float64
	PodRunning       int
	PodPending       int
	PodCapacity      int
	Architecture     string
	OSImage          string
	KernelVersion    string
	ContainerRuntime string
	Labels           map[string]string
	Taints           []string
	Conditions       []conditionRecord
}

type conditionRecord struct {
	Type           string
	Status         string
	Message        string
	LastHeartbeat  time.Time
	LastTransition time.Time
}

// Store manages mock node data with concurrency safety.
type Store struct {
	mu    sync.RWMutex
	items map[string]record
}

// NewStore returns a mock store seeded with deterministic nodes.
func NewStore(now time.Time) *Store {
	s := &Store{items: make(map[string]record)}
	for _, rec := range defaultSeed(now) {
		s.items[rec.Name] = rec
	}
	return s
}

// List returns sorted node summaries.
func (s *Store) List(now time.Time) []NodeSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]NodeSummary, 0, len(s.items))
	for _, rec := range s.items {
		result = append(result, toSummary(rec, now))
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.Compare(result[i].Name, result[j].Name) < 0
	})

	return result
}

// Get returns a node detail by name.
func (s *Store) Get(name string, now time.Time) (NodeDetail, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rec, ok := s.items[name]
	if !ok {
		return NodeDetail{}, ErrNotFound
	}

	return toDetail(rec, now), nil
}

func toSummary(rec record, now time.Time) NodeSummary {
	age := formatAge(now.Sub(rec.CreatedAt))
	cpu := UsageMetric{
		Used:       rec.CPUUsed,
		Capacity:   rec.CPUCapacity,
		Unit:       "cores",
		Percentage: percentage(rec.CPUUsed, rec.CPUCapacity),
	}
	mem := UsageMetric{
		Used:       rec.MemoryUsed,
		Capacity:   rec.MemoryCapacity,
		Unit:       "GiB",
		Percentage: percentage(rec.MemoryUsed, rec.MemoryCapacity),
	}

	return NodeSummary{
		Name:           rec.Name,
		Status:         rec.Status,
		Roles:          append([]string{}, rec.Roles...),
		Age:            age,
		KubeletVersion: rec.KubeletVersion,
		CPU:            cpu,
		Memory:         mem,
		Pods: PodSummary{
			Running:  rec.PodRunning,
			Pending:  rec.PodPending,
			Capacity: rec.PodCapacity,
		},
	}
}

func toDetail(rec record, now time.Time) NodeDetail {
	summary := toSummary(rec, now)

	conditions := make([]Condition, 0, len(rec.Conditions))
	for _, c := range rec.Conditions {
		conditions = append(conditions, Condition{
			Type:           c.Type,
			Status:         c.Status,
			Message:        c.Message,
			LastHeartbeat:  c.LastHeartbeat.Format(time.RFC3339),
			LastTransition: c.LastTransition.Format(time.RFC3339),
		})
	}

	labels := make(map[string]string, len(rec.Labels))
	for k, v := range rec.Labels {
		labels[k] = v
	}

	return NodeDetail{
		NodeSummary:      summary,
		Architecture:     rec.Architecture,
		OSImage:          rec.OSImage,
		KernelVersion:    rec.KernelVersion,
		ContainerRuntime: rec.ContainerRuntime,
		Labels:           labels,
		Taints:           append([]string{}, rec.Taints...),
		Conditions:       conditions,
	}
}

func percentage(used, capacity float64) float64 {
	if capacity <= 0 {
		return 0
	}
	pct := (used / capacity) * 100
	if pct < 0 {
		return 0
	}
	if pct > 100 {
		return 100
	}
	return round(pct, 1)
}

func round(value float64, decimals int) float64 {
	base := mathPow10(decimals)
	return float64(int64(value*base+0.5)) / base
}

func mathPow10(power int) float64 {
	result := 1.0
	for power > 0 {
		result *= 10
		power--
	}
	return result
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
	base := now.Add(-72 * time.Hour)

	return []record{
		{
			Name:             "node-1",
			Status:           "Ready",
			Roles:            []string{"control-plane", "master"},
			CreatedAt:        base,
			KubeletVersion:   "v1.28.3",
			CPUUsed:          6.5,
			CPUCapacity:      16,
			MemoryUsed:       48,
			MemoryCapacity:   128,
			PodRunning:       45,
			PodPending:       2,
			PodCapacity:      110,
			Architecture:     "amd64",
			OSImage:          "Ubuntu 22.04 LTS",
			KernelVersion:    "5.15.0-78-generic",
			ContainerRuntime: "containerd://1.7.8",
			Labels: map[string]string{
				"kubernetes.io/hostname":      "node-1",
				"topology.kubernetes.io/zone": "cn-shanghai-a",
			},
			Taints: []string{"node-role.kubernetes.io/control-plane:NoSchedule"},
			Conditions: []conditionRecord{
				{
					Type:           "Ready",
					Status:         "True",
					Message:        "Node is ready",
					LastHeartbeat:  now,
					LastTransition: base,
				},
			},
		},
		{
			Name:             "node-2",
			Status:           "Ready",
			Roles:            []string{"worker"},
			CreatedAt:        base.Add(6 * time.Hour),
			KubeletVersion:   "v1.28.3",
			CPUUsed:          9,
			CPUCapacity:      32,
			MemoryUsed:       72,
			MemoryCapacity:   256,
			PodRunning:       68,
			PodPending:       5,
			PodCapacity:      150,
			Architecture:     "amd64",
			OSImage:          "Ubuntu 22.04 LTS",
			KernelVersion:    "5.15.0-78-generic",
			ContainerRuntime: "containerd://1.7.8",
			Labels: map[string]string{
				"kubernetes.io/hostname":      "node-2",
				"topology.kubernetes.io/zone": "cn-shanghai-b",
				"nodepool":                    "blue",
			},
			Taints: []string{},
			Conditions: []conditionRecord{
				{
					Type:           "Ready",
					Status:         "True",
					Message:        "Node is ready",
					LastHeartbeat:  now,
					LastTransition: base.Add(2 * time.Hour),
				},
				{
					Type:           "DiskPressure",
					Status:         "False",
					Message:        "No disk pressure",
					LastHeartbeat:  now.Add(-10 * time.Minute),
					LastTransition: base.Add(12 * time.Hour),
				},
			},
		},
		{
			Name:             "node-3",
			Status:           "NotReady",
			Roles:            []string{"worker"},
			CreatedAt:        base.Add(12 * time.Hour),
			KubeletVersion:   "v1.28.3",
			CPUUsed:          2,
			CPUCapacity:      16,
			MemoryUsed:       24,
			MemoryCapacity:   128,
			PodRunning:       12,
			PodPending:       8,
			PodCapacity:      110,
			Architecture:     "amd64",
			OSImage:          "Ubuntu 22.04 LTS",
			KernelVersion:    "5.15.0-78-generic",
			ContainerRuntime: "containerd://1.7.8",
			Labels: map[string]string{
				"kubernetes.io/hostname":      "node-3",
				"topology.kubernetes.io/zone": "cn-shanghai-c",
				"nodepool":                    "green",
			},
			Taints: []string{"maintenance=true:NoSchedule"},
			Conditions: []conditionRecord{
				{
					Type:           "Ready",
					Status:         "False",
					Message:        "Kubelet stopped posting node status",
					LastHeartbeat:  now.Add(-30 * time.Minute),
					LastTransition: now.Add(-30 * time.Minute),
				},
				{
					Type:           "NetworkUnavailable",
					Status:         "True",
					Message:        "Cilium agent not ready",
					LastHeartbeat:  now.Add(-45 * time.Minute),
					LastTransition: now.Add(-45 * time.Minute),
				},
			},
		},
	}
}
