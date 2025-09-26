package server

import (
	"net/http"
	"strconv"
	"strings"

	"k8s_dashboard/internal/logs"
)

func (s *Server) handleLogStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	limit := 0
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}

	filter := logs.LogFilter{
		Namespace: query.Get("namespace"),
		Pod:       query.Get("pod"),
		Level:     query.Get("level"),
		Limit:     limit,
	}

	entries := s.logs.ListLogs(s.now(), filter)
	writeJSON(w, entries, http.StatusOK)
}

func (s *Server) handleLogMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	meta := s.logs.DescribeFilters()
	writeJSON(w, meta, http.StatusOK)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	items := s.logs.ListEvents(s.now())
	writeJSON(w, items, http.StatusOK)
}
