package cluster

import (
	"testing"

	"github.com/creimer/wikimap/internal/graph"
)

func TestBuildHierarchy(t *testing.T) {
	g := graph.NewGraph()
	for i := uint32(1); i <= 12; i++ {
		g.AddNode(graph.Node{ID: i})
	}
	for _, clique := range [][]uint32{{1, 2, 3, 4}, {5, 6, 7, 8}, {9, 10, 11, 12}} {
		for i := 0; i < len(clique); i++ {
			for j := i + 1; j < len(clique); j++ {
				g.AddEdge(clique[i], clique[j], 1.0)
				g.AddEdge(clique[j], clique[i], 1.0)
			}
		}
	}
	g.AddEdge(4, 5, 1.0)
	g.AddEdge(5, 4, 1.0)
	g.AddEdge(8, 9, 1.0)
	g.AddEdge(9, 8, 1.0)

	levels := BuildHierarchy(g, 3)

	if len(levels) < 2 {
		t.Fatalf("expected at least 2 hierarchy levels, got %d", len(levels))
	}

	bottomCount := countUnique(levels[len(levels)-1])
	topCount := countUnique(levels[0])
	if bottomCount <= topCount {
		t.Errorf("bottom level (%d communities) should have more than top level (%d)", bottomCount, topCount)
	}
}

func countUnique(m map[uint32]uint32) int {
	seen := make(map[uint32]bool)
	for _, v := range m {
		seen[v] = true
	}
	return len(seen)
}
