package server

import (
	"net/http"
	"strings"

	"k8s_dashboard/internal/pod"
)

func (s *Server) handlePods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := s.pods.List(s.now())
	writeJSON(w, payload, http.StatusOK)
}

func (s *Server) handlePodByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/api/pods/")
	if name == "" {
		http.NotFound(w, r)
		return
	}

	detail, err := s.pods.Get(name, s.now())
	if err != nil {
		if err == pod.ErrNotFound {
			writeJSON(w, errorResponse{Error: "Pod 不存在"}, http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load pod detail", http.StatusInternalServerError)
		return
	}

	writeJSON(w, detail, http.StatusOK)
}
