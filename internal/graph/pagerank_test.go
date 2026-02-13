package graph

import (
	"math"
	"testing"
)

func TestPageRank(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: 1, Title: "A"})
	g.AddNode(Node{ID: 2, Title: "B"})
	g.AddNode(Node{ID: 3, Title: "C"})
	g.AddEdge(1, 2, 1.0)
	g.AddEdge(2, 3, 1.0)
	g.AddEdge(3, 1, 1.0)

	ranks := PageRank(g, 0.85, 100)

	for id, rank := range ranks {
		if math.Abs(rank-1.0/3.0) > 0.01 {
			t.Errorf("node %d: expected rank ~0.333, got %.4f", id, rank)
		}
	}
}

func TestPageRankStarGraph(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: 1, Title: "A"})
	g.AddNode(Node{ID: 2, Title: "B"})
	g.AddNode(Node{ID: 3, Title: "C"})
	g.AddNode(Node{ID: 4, Title: "D"})
	g.AddEdge(2, 1, 1.0)
	g.AddEdge(3, 1, 1.0)
	g.AddEdge(4, 1, 1.0)

	ranks := PageRank(g, 0.85, 100)

	if ranks[1] <= ranks[2] || ranks[1] <= ranks[3] || ranks[1] <= ranks[4] {
		t.Errorf("expected node A to have highest rank, got: %v", ranks)
	}
}
