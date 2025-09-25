package main

import (
	"log"
	"net/http"
	"os"

	"k8s_dashboard/internal/server"
)

func main() {
	addr := defaultAddr()
	srv := server.New()

	log.Printf("starting dashboard server on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func defaultAddr() string {
	if v := os.Getenv("DASHBOARD_LISTEN"); v != "" {
		return v
	}
	return ":8080"
}
