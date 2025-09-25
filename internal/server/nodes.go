package server

import (
	"net/http"
	"strings"

	"k8s_dashboard/internal/node"
)

func (s *Server) handleNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := s.nodes.List(s.now())
	writeJSON(w, payload, http.StatusOK)
}

func (s *Server) handleNodeByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/api/nodes/")
	if name == "" {
		http.NotFound(w, r)
		return
	}

	detail, err := s.nodes.Get(name, s.now())
	if err != nil {
		if err == node.ErrNotFound {
			writeJSON(w, errorResponse{Error: "节点不存在"}, http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load node detail", http.StatusInternalServerError)
		return
	}

	writeJSON(w, detail, http.StatusOK)
}
