package tiles

import (
	"testing"

	"github.com/creimer/wikimap/internal/graph"
	"github.com/creimer/wikimap/internal/layout"
)

func TestTileCoord(t *testing.T) {
	tx, ty := WorldToTile(0.5, 0.5, 0)
	if tx != 0 || ty != 0 {
		t.Errorf("zoom 0: expected tile (0,0), got (%d,%d)", tx, ty)
	}

	tx, ty = WorldToTile(0.75, 0.75, 1)
	if tx != 1 || ty != 1 {
		t.Errorf("zoom 1: expected tile (1,1), got (%d,%d)", tx, ty)
	}
}

func TestGenerateTiles(t *testing.T) {
	g := graph.NewGraph()
	g.AddNode(graph.Node{ID: 1, Title: "A", PageViews: 1000})
	g.AddNode(graph.Node{ID: 2, Title: "B", PageViews: 500})
	g.AddNode(graph.Node{ID: 3, Title: "C", PageViews: 100})
	g.AddEdge(1, 2, 1.0)
	g.AddEdge(2, 3, 1.0)

	positions := map[uint32]layout.Point{
		1: {X: 0.2, Y: 0.3},
		2: {X: 0.7, Y: 0.8},
		3: {X: 0.1, Y: 0.9},
	}

	importance := map[uint32]float64{
		1: 1.0,
		2: 0.5,
		3: 0.1,
	}

	communities := map[uint32]uint32{
		1: 0, 2: 0, 3: 1,
	}

	tiles := GenerateTiles(g, positions, importance, communities, 0, 2)

	zoom0 := tilesAtZoom(tiles, 0)
	if len(zoom0) != 1 {
		t.Fatalf("expected 1 tile at zoom 0, got %d", len(zoom0))
	}

	tile := zoom0[0]
	if len(tile.Nodes) != 3 {
		t.Errorf("expected 3 nodes in zoom-0 tile, got %d", len(tile.Nodes))
	}
}

func tilesAtZoom(tiles []GeneratedTile, zoom int) []GeneratedTile {
	var result []GeneratedTile
	for _, t := range tiles {
		if t.Z == zoom {
			result = append(result, t)
		}
	}
	return result
}
