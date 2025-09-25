package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"k8s_dashboard/internal/cluster"
)

func TestHandleClusterOverview(t *testing.T) {
	fixedTime := time.Date(2024, 7, 12, 15, 30, 0, 0, time.UTC)
	srv := NewWithClock(func() time.Time {
		return fixedTime
	})

	req := httptest.NewRequest(http.MethodGet, "/api/cluster/overview", nil)
	rr := httptest.NewRecorder()

	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content type: %s", ct)
	}

	var payload cluster.ClusterOverview
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Info.Version != "v1.28.3" {
		t.Errorf("unexpected version %s", payload.Info.Version)
	}

	if payload.ResourceUsage.Memory.Timestamp != "2024-07-12T15:30:00Z" {
		t.Errorf("unexpected timestamp %s", payload.ResourceUsage.Memory.Timestamp)
	}
}

func TestHandleIndex(t *testing.T) {
	srv := New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("unexpected content type: %s", ct)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "集群概览") {
		t.Fatalf("expected body to contain overview heading")
	}
}

func TestHandleNamespacesList(t *testing.T) {
	fixedTime := time.Date(2024, 7, 12, 15, 30, 0, 0, time.UTC)
	srv := NewWithClock(func() time.Time {
		return fixedTime
	})

	req := httptest.NewRequest(http.MethodGet, "/api/namespaces", nil)
	rr := httptest.NewRecorder()

	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload []map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(payload) == 0 {
		t.Fatalf("expected namespaces list")
	}

	if payload[0]["name"] != "default" {
		t.Fatalf("expected default namespace first")
	}
}

func TestHandleNamespaceCreateAndDelete(t *testing.T) {
	fixedTime := time.Date(2024, 7, 12, 15, 30, 0, 0, time.UTC)
	srv := NewWithClock(func() time.Time {
		return fixedTime
	})

	body := map[string]any{"name": "staging"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/namespaces", bytes.NewReader(raw))
	rr := httptest.NewRecorder()

	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	var created map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	if created["name"] != "staging" {
		t.Fatalf("unexpected namespace name %v", created["name"])
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/api/namespaces/staging", nil)
	delRR := httptest.NewRecorder()
	srv.ServeHTTP(delRR, delReq)

	if delRR.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", delRR.Code)
	}

	delReq2 := httptest.NewRequest(http.MethodDelete, "/api/namespaces/staging", nil)
	delRR2 := httptest.NewRecorder()
	srv.ServeHTTP(delRR2, delReq2)

	if delRR2.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", delRR2.Code)
	}
}
