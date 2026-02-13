package cluster

import (
	"testing"

	"github.com/creimer/wikimap/internal/graph"
)

func TestLouvainTwoClusters(t *testing.T) {
	g := graph.NewGraph()
	for _, id := range []uint32{1, 2, 3} {
		g.AddNode(graph.Node{ID: id, Title: ""})
	}
	for _, id := range []uint32{4, 5, 6} {
		g.AddNode(graph.Node{ID: id, Title: ""})
	}
	for _, e := range [][2]uint32{{1, 2}, {2, 1}, {1, 3}, {3, 1}, {2, 3}, {3, 2}} {
		g.AddEdge(e[0], e[1], 1.0)
	}
	for _, e := range [][2]uint32{{4, 5}, {5, 4}, {4, 6}, {6, 4}, {5, 6}, {6, 5}} {
		g.AddEdge(e[0], e[1], 1.0)
	}
	g.AddEdge(3, 4, 1.0)
	g.AddEdge(4, 3, 1.0)

	communities := Louvain(g)

	if communities[1] != communities[2] || communities[1] != communities[3] {
		t.Errorf("clique 1 should be one community: %v", communities)
	}
	if communities[4] != communities[5] || communities[4] != communities[6] {
		t.Errorf("clique 2 should be one community: %v", communities)
	}
	if communities[1] == communities[4] {
		t.Errorf("cliques should be in different communities: %v", communities)
	}
}
