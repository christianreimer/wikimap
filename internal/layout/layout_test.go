package layout

import (
	"math"
	"testing"

	"github.com/creimer/wikimap/internal/graph"
)

func TestForceDirectedConnectedPair(t *testing.T) {
	g := graph.NewGraph()
	g.AddNode(graph.Node{ID: 1, Title: "A"})
	g.AddNode(graph.Node{ID: 2, Title: "B"})
	g.AddEdge(1, 2, 1.0)

	positions := ForceDirected(g, 200)

	dx := positions[1].X - positions[2].X
	dy := positions[1].Y - positions[2].Y
	dist := math.Sqrt(float64(dx*dx + dy*dy))

	if dist < 0.01 {
		t.Error("connected nodes should not overlap")
	}
	if dist > 100 {
		t.Errorf("connected nodes should be relatively close, got distance %.2f", dist)
	}
}

func TestForceDirectedDisconnectedRepel(t *testing.T) {
	g := graph.NewGraph()
	g.AddNode(graph.Node{ID: 1, Title: "A"})
	g.AddNode(graph.Node{ID: 2, Title: "B"})

	positions := ForceDirected(g, 200)

	dx := positions[1].X - positions[2].X
	dy := positions[1].Y - positions[2].Y
	dist := math.Sqrt(float64(dx*dx + dy*dy))

	if dist < 1.0 {
		t.Errorf("disconnected nodes should repel, got distance %.2f", dist)
	}
}

func TestForceDirectedNormalized(t *testing.T) {
	g := graph.NewGraph()
	for i := uint32(1); i <= 5; i++ {
		g.AddNode(graph.Node{ID: i})
	}
	g.AddEdge(1, 2, 1.0)
	g.AddEdge(2, 3, 1.0)
	g.AddEdge(3, 4, 1.0)
	g.AddEdge(4, 5, 1.0)

	positions := ForceDirected(g, 300)

	for id, pos := range positions {
		if pos.X < 0 || pos.X > 1 || pos.Y < 0 || pos.Y > 1 {
			t.Errorf("node %d position (%.4f, %.4f) outside [0,1]", id, pos.X, pos.Y)
		}
	}
}
