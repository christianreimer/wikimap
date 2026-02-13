package parser

import "github.com/creimer/wikimap/internal/graph"

func BuildGraph(articles []Article) *graph.Graph {
	g := graph.NewGraph()

	titleToID := make(map[string]uint32, len(articles))
	for _, a := range articles {
		g.AddNode(graph.Node{ID: a.ID, Title: a.Title})
		titleToID[a.Title] = a.ID
	}

	for _, a := range articles {
		for _, link := range a.Links {
			if targetID, ok := titleToID[link]; ok && targetID != a.ID {
				g.AddEdge(a.ID, targetID, 1.0)
			}
		}
	}

	return g
}
