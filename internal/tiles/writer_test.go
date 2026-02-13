package tiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndReadTile(t *testing.T) {
	dir := t.TempDir()

	tile := GeneratedTile{
		Z: 2, X: 1, Y: 3,
		Nodes: []TileNode{
			{ID: 42, X: 0.5, Y: 0.5, Label: "Test", Importance: 0.9, ClusterID: 1},
		},
		Edges: []TileEdge{
			{FromID: 42, ToID: 99, Weight: 0.7, ToX: 0.8, ToY: 0.2},
		},
	}

	err := WriteTile(dir, tile)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, "2", "1", "3.pbf")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("tile file not written at %s", path)
	}

	loaded, err := ReadTile(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(loaded.Nodes))
	}
	if loaded.Nodes[0].Label != "Test" {
		t.Errorf("expected label 'Test', got %q", loaded.Nodes[0].Label)
	}
	if loaded.Nodes[0].ID != 42 {
		t.Errorf("expected ID 42, got %d", loaded.Nodes[0].ID)
	}
}
