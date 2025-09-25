package server

import (
	"net/http"
	"time"

	"k8s_dashboard/internal/cluster"
	"k8s_dashboard/internal/namespace"
	"k8s_dashboard/internal/node"
	"k8s_dashboard/internal/pod"
)

// Server exposes HTTP handlers for the dashboard application.
type Server struct {
	mux        *http.ServeMux
	now        func() time.Time
	namespaces *namespace.Store
	nodes      *node.Store
	pods       *pod.Store
}

// New constructs a server with default dependencies.
func New() *Server {
	return NewWithClock(time.Now)
}

// NewWithClock allows injection of a deterministic time source for testing.
func NewWithClock(now func() time.Time) *Server {
	s := &Server{
		mux:        http.NewServeMux(),
		now:        now,
		namespaces: namespace.NewStore(now()),
		nodes:      node.NewStore(now()),
		pods:       pod.NewStore(now()),
	}
	s.registerRoutes()
	return s
}

// ServeHTTP makes Server implement http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/cluster/overview", s.handleClusterOverview)
	s.mux.HandleFunc("/api/namespaces", s.handleNamespaces)
	s.mux.HandleFunc("/api/namespaces/", s.handleNamespaceByName)
	s.mux.HandleFunc("/api/nodes", s.handleNodes)
	s.mux.HandleFunc("/api/nodes/", s.handleNodeByName)
	s.mux.HandleFunc("/api/pods", s.handlePods)
	s.mux.HandleFunc("/api/pods/", s.handlePodByName)
}

func (s *Server) handleClusterOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	overview := cluster.MockOverview(s.now())

	writeJSON(w, overview, http.StatusOK)
}
