package server

import (
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
