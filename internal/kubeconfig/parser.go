package kubeconfig

import (
	"fmt"
	"io"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Parse decodes kubeconfig bytes and returns a Summary.
func Parse(r io.Reader, now time.Time) (Summary, error) {
	var cfg struct {
		Clusters []struct {
			Name    string `yaml:"name"`
			Cluster struct {
				Server string `yaml:"server"`
			} `yaml:"cluster"`
		} `yaml:"clusters"`
		Contexts []struct {
			Name    string `yaml:"name"`
			Context struct {
				Cluster string `yaml:"cluster"`
				User    string `yaml:"user"`
			} `yaml:"context"`
		} `yaml:"contexts"`
		CurrentContext string `yaml:"current-context"`
	}

	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&cfg); err != nil {
		return Summary{}, fmt.Errorf("解析 kubeconfig 失败: %w", err)
	}

	if len(cfg.Clusters) == 0 {
		return Summary{}, fmt.Errorf("kubeconfig 中缺少 clusters 配置")
	}

	clusters := make([]Cluster, 0, len(cfg.Clusters))
	for _, c := range cfg.Clusters {
		clusters = append(clusters, Cluster{Name: strings.TrimSpace(c.Name), Server: strings.TrimSpace(c.Cluster.Server)})
	}

	contexts := make([]Context, 0, len(cfg.Contexts))
	for _, c := range cfg.Contexts {
		contexts = append(contexts, Context{Name: strings.TrimSpace(c.Name), Cluster: strings.TrimSpace(c.Context.Cluster), User: strings.TrimSpace(c.Context.User)})
	}

	name := cfg.CurrentContext
	if name == "" && len(contexts) > 0 {
		name = contexts[0].Name
	}

	return Summary{
		Name:           strings.TrimSpace(name),
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: strings.TrimSpace(cfg.CurrentContext),
		ImportedAt:     now,
	}, nil
}
