package cluster

import "time"

// ClusterOverview aggregates summary data shown on the dashboard.
type ClusterOverview struct {
	Info          ClusterInfo   `json:"info"`
	ResourceUsage ResourceUsage `json:"resourceUsage"`
	RecentEvents  []Event       `json:"recentEvents"`
}

// ClusterInfo describes the basic cluster metadata displayed in the UI.
type ClusterInfo struct {
	Name             string `json:"name"`
	Version          string `json:"version"`
	NodeCount        int    `json:"nodeCount"`
	NamespaceCount   int    `json:"namespaceCount"`
	RunningPodCount  int    `json:"runningPodCount"`
	PendingPodCount  int    `json:"pendingPodCount"`
	FailedPodCount   int    `json:"failedPodCount"`
	TotalPodCapacity int    `json:"totalPodCapacity"`
}

// ResourceUsage summarises utilisation metrics.
type ResourceUsage struct {
	CPU    UsageMetric `json:"cpu"`
	Memory UsageMetric `json:"memory"`
}

// UsageMetric stores utilised percentage information.
type UsageMetric struct {
	Used      float64 `json:"used"`
	Total     float64 `json:"total"`
	Unit      string  `json:"unit"`
	Timestamp string  `json:"timestamp"`
}

// Event models a cluster event entry.
type Event struct {
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	Object    string `json:"object"`
	Timestamp string `json:"timestamp"`
}

// MockOverview exposes deterministic mock data for the overview endpoint.
func MockOverview(now time.Time) ClusterOverview {
	timestamp := now.UTC().Format(time.RFC3339)

	return ClusterOverview{
		Info: ClusterInfo{
			Name:             "Mock Production Cluster",
			Version:          "v1.28.3",
			NodeCount:        8,
			NamespaceCount:   17,
			RunningPodCount:  152,
			PendingPodCount:  4,
			FailedPodCount:   2,
			TotalPodCapacity: 240,
		},
		ResourceUsage: ResourceUsage{
			CPU: UsageMetric{
				Used:      52.3,
				Total:     128,
				Unit:      "cores",
				Timestamp: timestamp,
			},
			Memory: UsageMetric{
				Used:      386.4,
				Total:     512,
				Unit:      "GiB",
				Timestamp: timestamp,
			},
		},
		RecentEvents: []Event{
			{
				Type:      "Warning",
				Reason:    "FailedScheduling",
				Message:   "0/8 nodes are available: 8 Insufficient cpu.",
				Object:    "pod/backend-7d8fdc9f7c-abc12",
				Timestamp: timestamp,
			},
			{
				Type:      "Normal",
				Reason:    "ScalingReplicaSet",
				Message:   "Scaled up deployment/frontend to 5",
				Object:    "deployment/frontend",
				Timestamp: timestamp,
			},
			{
				Type:      "Normal",
				Reason:    "NodeReady",
				Message:   "Node node-5 is Ready",
				Object:    "node/node-5",
				Timestamp: timestamp,
			},
		},
	}
}
