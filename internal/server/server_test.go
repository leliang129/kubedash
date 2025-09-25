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

func TestHandleNodesEndpoints(t *testing.T) {
	fixedTime := time.Date(2024, 7, 12, 15, 30, 0, 0, time.UTC)
	srv := NewWithClock(func() time.Time {
		return fixedTime
	})

	listReq := httptest.NewRequest(http.MethodGet, "/api/nodes", nil)
	listRR := httptest.NewRecorder()
	srv.ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", listRR.Code)
	}

	var nodes []map[string]any
	if err := json.NewDecoder(listRR.Body).Decode(&nodes); err != nil {
		t.Fatalf("decode nodes list: %v", err)
	}

	if len(nodes) == 0 {
		t.Fatalf("expected nodes in list")
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/api/nodes/node-1", nil)
	detailRR := httptest.NewRecorder()
	srv.ServeHTTP(detailRR, detailReq)

	if detailRR.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", detailRR.Code)
	}

	var detail map[string]any
	if err := json.NewDecoder(detailRR.Body).Decode(&detail); err != nil {
		t.Fatalf("decode node detail: %v", err)
	}

	if detail["name"] != "node-1" {
		t.Fatalf("expected node-1 detail, got %v", detail["name"])
	}

	notFoundReq := httptest.NewRequest(http.MethodGet, "/api/nodes/ghost", nil)
	notFoundRR := httptest.NewRecorder()
	srv.ServeHTTP(notFoundRR, notFoundReq)

	if notFoundRR.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", notFoundRR.Code)
	}
}

func TestHandlePodsEndpoints(t *testing.T) {
	fixedTime := time.Date(2024, 7, 12, 15, 30, 0, 0, time.UTC)
	srv := NewWithClock(func() time.Time {
		return fixedTime
	})

	listReq := httptest.NewRequest(http.MethodGet, "/api/pods", nil)
	listRR := httptest.NewRecorder()
	srv.ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", listRR.Code)
	}

	var pods []map[string]any
	if err := json.NewDecoder(listRR.Body).Decode(&pods); err != nil {
		t.Fatalf("decode pods list: %v", err)
	}

	if len(pods) == 0 {
		t.Fatalf("expected pods in list")
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/api/pods/frontend-7d8fdc9f7c-abc12", nil)
	detailRR := httptest.NewRecorder()
	srv.ServeHTTP(detailRR, detailReq)

	if detailRR.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", detailRR.Code)
	}

	var detail map[string]any
	if err := json.NewDecoder(detailRR.Body).Decode(&detail); err != nil {
		t.Fatalf("decode pod detail: %v", err)
	}

	if detail["name"] != "frontend-7d8fdc9f7c-abc12" {
		t.Fatalf("unexpected pod name %v", detail["name"])
	}

	notFoundReq := httptest.NewRequest(http.MethodGet, "/api/pods/missing", nil)
	notFoundRR := httptest.NewRecorder()
	srv.ServeHTTP(notFoundRR, notFoundReq)

	if notFoundRR.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", notFoundRR.Code)
	}
}

func TestHandleDeploymentsEndpoints(t *testing.T) {
	fixedTime := time.Date(2024, 7, 12, 15, 30, 0, 0, time.UTC)
	srv := NewWithClock(func() time.Time {
		return fixedTime
	})

	listReq := httptest.NewRequest(http.MethodGet, "/api/deployments", nil)
	listRR := httptest.NewRecorder()
	srv.ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", listRR.Code)
	}

	var deployments []map[string]any
	if err := json.NewDecoder(listRR.Body).Decode(&deployments); err != nil {
		t.Fatalf("decode deployments list: %v", err)
	}

	if len(deployments) == 0 {
		t.Fatalf("expected deployments in list")
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/api/deployments/frontend", nil)
	detailRR := httptest.NewRecorder()
	srv.ServeHTTP(detailRR, detailReq)

	if detailRR.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", detailRR.Code)
	}

	var detail map[string]any
	if err := json.NewDecoder(detailRR.Body).Decode(&detail); err != nil {
		t.Fatalf("decode deployment detail: %v", err)
	}

	if detail["name"] != "frontend" {
		t.Fatalf("unexpected deployment name %v", detail["name"])
	}

	scaleBody := map[string]any{"replicas": 6}
	payload, _ := json.Marshal(scaleBody)
	scaleReq := httptest.NewRequest(http.MethodPut, "/api/deployments/frontend/scale", bytes.NewReader(payload))
	scaleRR := httptest.NewRecorder()
	srv.ServeHTTP(scaleRR, scaleReq)

	if scaleRR.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", scaleRR.Code)
	}

	var scaled map[string]any
	if err := json.NewDecoder(scaleRR.Body).Decode(&scaled); err != nil {
		t.Fatalf("decode scaled detail: %v", err)
	}

	if scaled["desiredReplicas"].(float64) != 6 {
		t.Fatalf("expected desired replicas 6, got %v", scaled["desiredReplicas"])
	}

	invalidReq := httptest.NewRequest(http.MethodPut, "/api/deployments/frontend/scale", bytes.NewReader([]byte(`{"replicas":-1}`)))
	invalidRR := httptest.NewRecorder()
	srv.ServeHTTP(invalidRR, invalidReq)

	if invalidRR.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", invalidRR.Code)
	}

	notFoundReq := httptest.NewRequest(http.MethodGet, "/api/deployments/missing", nil)
	notFoundRR := httptest.NewRecorder()
	srv.ServeHTTP(notFoundRR, notFoundReq)

	if notFoundRR.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", notFoundRR.Code)
	}
}
