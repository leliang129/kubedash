package kubeconfig

import (
	"strings"
	"testing"
	"time"
)

const sample = `apiVersion: v1
clusters:
- cluster:
    server: https://example.com
  name: prod
contexts:
- context:
    cluster: prod
    user: admin
  name: prod-context
current-context: prod-context
`

func TestParse(t *testing.T) {
	summary, err := Parse(strings.NewReader(sample), time.Unix(0, 0))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if summary.Name != "prod-context" {
		t.Fatalf("unexpected name: %s", summary.Name)
	}

	if len(summary.Clusters) != 1 || summary.Clusters[0].Server != "https://example.com" {
		t.Fatalf("unexpected clusters: %+v", summary.Clusters)
	}

	if len(summary.Contexts) != 1 || summary.Contexts[0].Cluster != "prod" {
		t.Fatalf("unexpected contexts: %+v", summary.Contexts)
	}
}

func TestParseMissingClusters(t *testing.T) {
	_, err := Parse(strings.NewReader("apiVersion: v1"), time.Now())
	if err == nil {
		t.Fatalf("expected error for empty kubeconfig")
	}
}
