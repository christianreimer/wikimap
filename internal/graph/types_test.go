package graph

import "testing"

func TestNewGraph(t *testing.T) {
	g := NewGraph()
	if g.NodeCount() != 0 {
		t.Fatalf("expected 0 nodes, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 0 {
		t.Fatalf("expected 0 edges, got %d", g.EdgeCount())
	}
}

func TestAddNode(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: 1, Title: "Go (programming language)", PageViews: 50000})
	g.AddNode(Node{ID: 2, Title: "Rust (programming language)", PageViews: 30000})

	if g.NodeCount() != 2 {
		t.Fatalf("expected 2 nodes, got %d", g.NodeCount())
	}

	n, ok := g.Node(1)
	if !ok {
		t.Fatal("expected to find node 1")
	}
	if n.Title != "Go (programming language)" {
		t.Fatalf("expected title 'Go (programming language)', got %q", n.Title)
	}
}

func TestAddEdge(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: 1, Title: "A"})
	g.AddNode(Node{ID: 2, Title: "B"})
	g.AddEdge(1, 2, 1.0)

	if g.EdgeCount() != 1 {
		t.Fatalf("expected 1 edge, got %d", g.EdgeCount())
	}

	neighbors := g.OutNeighbors(1)
	if len(neighbors) != 1 || neighbors[0] != 2 {
		t.Fatalf("expected out-neighbors [2], got %v", neighbors)
	}

	inNeighbors := g.InNeighbors(2)
	if len(inNeighbors) != 1 || inNeighbors[0] != 1 {
		t.Fatalf("expected in-neighbors [1], got %v", inNeighbors)
	}
}
