package server

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"k8s_dashboard/internal/kubeconfig"
)

const maxImportSize = 5 << 20 // 5 MiB

func (s *Server) handleClusterImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(maxImportSize); err != nil {
		http.Error(w, "解析上传文件失败", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "未找到 kubeconfig 文件", http.StatusBadRequest)
		return
	}
	defer file.Close()

	summary, err := kubeconfig.Parse(limitReader(file, maxImportSize), s.now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if summary.Name == "" {
		summary.Name = strings.TrimSuffix(filepath.Base(header.Filename), filepath.Ext(header.Filename))
	}

	s.kubeconfigs.Add(summary)
	writeJSON(w, summary, http.StatusCreated)
}

func (s *Server) handleClusterImports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	items := s.kubeconfigs.List()
	writeJSON(w, items, http.StatusOK)
}

func limitReader(r io.Reader, n int64) io.Reader {
	return io.LimitReader(r, n)
}
