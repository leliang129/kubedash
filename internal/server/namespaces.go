package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"k8s_dashboard/internal/namespace"
)

type createNamespaceRequest struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (s *Server) handleNamespaces(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleNamespacesList(w, r)
	case http.MethodPost:
		s.handleNamespaceCreate(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleNamespaceByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/api/namespaces/")
	if name == "" {
		http.NotFound(w, r)
		return
	}

	if deleted := s.namespaces.Delete(name); !deleted {
		http.Error(w, "namespace not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleNamespacesList(w http.ResponseWriter, r *http.Request) {
	payload := s.namespaces.List(s.now())
	writeJSON(w, payload, http.StatusOK)
}

func (s *Server) handleNamespaceCreate(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "unable to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req createNamespaceRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	ns, err := s.namespaces.Create(req.Name, s.now(), req.Labels)
	if err != nil {
		switch err {
		case namespace.ErrInvalidName:
			writeJSON(w, errorResponse{Error: "命名空间名称格式不正确，请使用小写字母、数字或连字符"}, http.StatusBadRequest)
		case namespace.ErrExists:
			writeJSON(w, errorResponse{Error: "命名空间已存在"}, http.StatusConflict)
		default:
			http.Error(w, "failed to create namespace", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, ns, http.StatusCreated)
}

func writeJSON(w http.ResponseWriter, payload any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}
