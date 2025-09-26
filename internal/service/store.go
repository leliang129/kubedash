package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// ErrNotFound indicates the service does not exist in the mock store.
var ErrNotFound = errors.New("service not found")

// Port represents a single service port mapping.
type Port struct {
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	Port       int    `json:"port"`
	TargetPort int    `json:"targetPort"`
	NodePort   *int   `json:"nodePort,omitempty"`
}

// RelatedPod gives a lightweight view of pods selected by the service.
type RelatedPod struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Node      string `json:"node"`
}

// Summary powers the service list view.
type Summary struct {
	Name        string   `json:"name"`
	Namespace   string   `json:"namespace"`
	Type        string   `json:"type"`
	ClusterIP   string   `json:"clusterIP"`
	ExternalIPs []string `json:"externalIPs"`
	Ports       []Port   `json:"ports"`
	Status      string   `json:"status"`
	Age         string   `json:"age"`
}

// Detail extends Summary with selector metadata.
type Detail struct {
	Summary
	Selector    map[string]string `json:"selector"`
	Endpoints   []string          `json:"endpoints"`
	RelatedPods []RelatedPod      `json:"relatedPods"`
	CreatedAt   string            `json:"createdAt"`
	Description string            `json:"description"`
}

type record struct {
	Summary
	CreatedAt   time.Time
	Selector    map[string]string
	Endpoints   []string
	RelatedPods []RelatedPod
	Description string
}

// Store manages mock service data with concurrency safety.
type Store struct {
	mu    sync.RWMutex
	items map[string]record
}

// NewStore returns a mock service store seeded with deterministic services.
func NewStore(now time.Time) *Store {
	s := &Store{items: make(map[string]record)}
	for _, rec := range defaultSeed(now) {
		s.items[rec.Name] = rec
	}
	return s
}

// List returns sorted service summaries.
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

// Get returns a service detail by name.
func (s *Store) Get(name string, now time.Time) (Detail, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if rec, ok := s.items[name]; ok {
		summary := decorateSummary(rec.Summary, rec.CreatedAt, now)
		detail := Detail{
			Summary:     summary,
			Selector:    copyMap(rec.Selector),
			Endpoints:   append([]string{}, rec.Endpoints...),
			RelatedPods: append([]RelatedPod{}, rec.RelatedPods...),
			CreatedAt:   rec.CreatedAt.Format(time.RFC3339),
			Description: rec.Description,
		}
		return detail, nil
	}

	return Detail{}, ErrNotFound
}

func decorateSummary(sum Summary, createdAt, now time.Time) Summary {
	out := sum
	out.Age = formatAge(now.Sub(createdAt))
	out.ExternalIPs = append([]string{}, sum.ExternalIPs...)
	out.Ports = copyPorts(sum.Ports)
	return out
}

func copyPorts(ports []Port) []Port {
	out := make([]Port, len(ports))
	for i, p := range ports {
		item := p
		if p.NodePort != nil {
			np := *p.NodePort
			item.NodePort = &np
		}
		out[i] = item
	}
	return out
}

func copyMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
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

func intPtr(v int) *int {
	return &v
}

func defaultSeed(now time.Time) []record {
	base := now.Add(-48 * time.Hour)

	return []record{
		{
			Summary: Summary{
				Name:      "frontend",
				Namespace: "default",
				Type:      "ClusterIP",
				ClusterIP: "10.96.10.21",
				Ports: []Port{
					{Name: "http", Protocol: "TCP", Port: 80, TargetPort: 8080},
					{Name: "metrics", Protocol: "TCP", Port: 9000, TargetPort: 9000},
				},
				Status: "Active",
			},
			CreatedAt: base,
			Selector: map[string]string{
				"app":  "frontend",
				"tier": "web",
			},
			Endpoints: []string{"10.0.0.11:8080", "10.0.0.12:8080"},
			RelatedPods: []RelatedPod{
				{Name: "frontend-7d8fdc9f7c-abc12", Namespace: "default", Status: "Running", Node: "node-2"},
				{Name: "frontend-7d8fdc9f7c-def34", Namespace: "default", Status: "Running", Node: "node-3"},
			},
			Description: "核心入口流量的前端服务",
		},
		{
			Summary: Summary{
				Name:        "edge-gateway",
				Namespace:   "prod",
				Type:        "LoadBalancer",
				ClusterIP:   "10.96.20.5",
				ExternalIPs: []string{"203.0.113.24"},
				Ports: []Port{
					{Name: "http", Protocol: "TCP", Port: 80, TargetPort: 8080, NodePort: intPtr(30080)},
					{Name: "https", Protocol: "TCP", Port: 443, TargetPort: 8443, NodePort: intPtr(30443)},
				},
				Status: "Pending",
			},
			CreatedAt: base.Add(12 * time.Hour),
			Selector: map[string]string{
				"app":       "edge-gateway",
				"component": "ingress",
			},
			Endpoints: []string{"34.87.18.4:80", "34.87.18.4:443"},
			RelatedPods: []RelatedPod{
				{Name: "edge-gateway-7d8fdc9f7c-9012a", Namespace: "prod", Status: "Running", Node: "node-1"},
				{Name: "edge-gateway-7d8fdc9f7c-9012b", Namespace: "prod", Status: "Pending", Node: ""},
			},
			Description: "对外暴露的流量入口，等待负载均衡器分配公网 IP",
		},
		{
			Summary: Summary{
				Name:      "batch-metrics",
				Namespace: "batch",
				Type:      "NodePort",
				ClusterIP: "10.96.35.42",
				Ports: []Port{
					{Name: "http", Protocol: "TCP", Port: 8080, TargetPort: 8080, NodePort: intPtr(32045)},
				},
				Status: "Active",
			},
			CreatedAt: base.Add(30 * time.Hour),
			Selector: map[string]string{
				"job": "metrics",
			},
			Endpoints: []string{"10.0.12.5:8080"},
			RelatedPods: []RelatedPod{
				{Name: "batch-metrics-7c5d6f6b4d-xk9p2", Namespace: "batch", Status: "Running", Node: "node-2"},
			},
			Description: "批处理作业指标对接 Prometheus 的临时 NodePort 服务",
		},
	}
}
