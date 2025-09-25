package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"k8s_dashboard/internal/deploy"
)

type scaleRequest struct {
	Replicas int `json:"replicas"`
}

func (s *Server) handleDeployments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := s.deployments.List(s.now())
	writeJSON(w, payload, http.StatusOK)
}

func (s *Server) handleDeploymentByName(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/deployments/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	segments := strings.Split(path, "/")
	name := segments[0]

	switch r.Method {
	case http.MethodGet:
		if len(segments) != 1 {
			http.NotFound(w, r)
			return
		}
		detail, err := s.deployments.Get(name, s.now())
		if err != nil {
			if err == deploy.ErrNotFound {
				writeJSON(w, errorResponse{Error: "Deployment 不存在"}, http.StatusNotFound)
				return
			}
			http.Error(w, "failed to load deployment detail", http.StatusInternalServerError)
			return
		}
		writeJSON(w, detail, http.StatusOK)
	case http.MethodPut:
		if len(segments) != 2 || segments[1] != "scale" {
			http.NotFound(w, r)
			return
		}

		var req scaleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}

		detail, err := s.deployments.Scale(name, req.Replicas, s.now())
		if err != nil {
			switch err {
			case deploy.ErrInvalidReplicas:
				writeJSON(w, errorResponse{Error: "副本数无效"}, http.StatusBadRequest)
			case deploy.ErrNotFound:
				writeJSON(w, errorResponse{Error: "Deployment 不存在"}, http.StatusNotFound)
			default:
				http.Error(w, "failed to scale deployment", http.StatusInternalServerError)
			}
			return
		}

		writeJSON(w, detail, http.StatusOK)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
