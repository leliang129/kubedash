package server

import (
	"encoding/json"
	"net/http"
	"time"

	"k8s_dashboard/internal/cluster"
)

// Server exposes HTTP handlers for the dashboard application.
type Server struct {
	mux *http.ServeMux
	now func() time.Time
}

// New constructs a server with default dependencies.
func New() *Server {
	return NewWithClock(time.Now)
}

// NewWithClock allows injection of a deterministic time source for testing.
func NewWithClock(now func() time.Time) *Server {
	s := &Server{
		mux: http.NewServeMux(),
		now: now,
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
}

func (s *Server) handleClusterOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	overview := cluster.MockOverview(s.now())

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(overview); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
