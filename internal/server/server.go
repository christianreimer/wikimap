package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	tileDir string
	webDir  string
}

func NewServer(tileDir, webDir string) *Server {
	return &Server{tileDir: tileDir, webDir: webDir}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/tiles/", s.handleTile)

	if s.webDir != "" {
		mux.Handle("/", http.FileServer(http.Dir(s.webDir)))
	}

	return corsMiddleware(mux)
}

func (s *Server) handleTile(w http.ResponseWriter, r *http.Request) {
	var z, x, y int
	_, err := fmt.Sscanf(r.URL.Path, "/tiles/%d/%d/%d.pbf", &z, &x, &y)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	path := filepath.Join(s.tileDir, fmt.Sprintf("%d", z), fmt.Sprintf("%d", x), fmt.Sprintf("%d.pbf", y))
	data, err := os.ReadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Write(data)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}

		next.ServeHTTP(w, r)
	})
}
