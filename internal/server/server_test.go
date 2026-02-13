package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/creimer/wikimap/internal/tiles"
)

func TestServeTile(t *testing.T) {
	dir := t.TempDir()

	tile := tiles.GeneratedTile{
		Z: 0, X: 0, Y: 0,
		Nodes: []tiles.TileNode{
			{ID: 1, X: 0.5, Y: 0.5, Label: "Test", Importance: 1.0},
		},
	}
	if err := tiles.WriteTile(dir, tile); err != nil {
		t.Fatal(err)
	}

	srv := NewServer(dir, "")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/tiles/0/0/0.pbf")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServeTile404(t *testing.T) {
	dir := t.TempDir()

	srv := NewServer(dir, "")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/tiles/99/99/99.pbf")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestCORSHeaders(t *testing.T) {
	dir := t.TempDir()

	tile := tiles.GeneratedTile{Z: 0, X: 0, Y: 0}
	tiles.WriteTile(dir, tile)

	srv := NewServer(dir, "")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/tiles/0/0/0.pbf")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	cors := resp.Header.Get("Access-Control-Allow-Origin")
	if cors != "*" {
		t.Errorf("expected CORS header '*', got %q", cors)
	}

	cacheControl := resp.Header.Get("Cache-Control")
	if cacheControl == "" {
		t.Error("expected Cache-Control header")
	}
}
