package server

import (
	"net/http"
	"strings"

	"k8s_dashboard/internal/service"
)

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := s.services.List(s.now())
	writeJSON(w, payload, http.StatusOK)
}

func (s *Server) handleServiceByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/api/services/")
	if name == "" {
		http.NotFound(w, r)
		return
	}

	detail, err := s.services.Get(name, s.now())
	if err != nil {
		if err == service.ErrNotFound {
			writeJSON(w, errorResponse{Error: "Service 不存在"}, http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load service detail", http.StatusInternalServerError)
		return
	}

	writeJSON(w, detail, http.StatusOK)
}
