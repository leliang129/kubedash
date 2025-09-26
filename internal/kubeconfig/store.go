package kubeconfig

import (
	"sync"
	"time"
)

// Summary describes an imported kubeconfig resource.
type Summary struct {
	Name           string    `json:"name"`
	Clusters       []Cluster `json:"clusters"`
	Contexts       []Context `json:"contexts"`
	CurrentContext string    `json:"currentContext"`
	ImportedAt     time.Time `json:"importedAt"`
}

// Cluster captures minimal cluster information from kubeconfig.
type Cluster struct {
	Name   string `json:"name"`
	Server string `json:"server"`
}

// Context captures context associations.
type Context struct {
	Name    string `json:"name"`
	Cluster string `json:"cluster"`
	User    string `json:"user"`
}

// Store keeps track of imported kubeconfig summaries.
type Store struct {
	mu    sync.RWMutex
	items []Summary
}

// NewStore returns an empty kubeconfig store.
func NewStore() *Store {
	return &Store{}
}

// List returns imported kubeconfigs in reverse chronological order.
func (s *Store) List() []Summary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Summary, len(s.items))
	copy(result, s.items)
	return result
}

// Add stores a new kubeconfig summary.
func (s *Store) Add(summary Summary) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// prepend to keep newest first
	s.items = append([]Summary{summary}, s.items...)
}
