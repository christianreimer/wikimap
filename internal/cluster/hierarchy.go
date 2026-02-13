package cluster

import "github.com/creimer/wikimap/internal/graph"

func BuildHierarchy(g *graph.Graph, maxLevels int) []map[uint32]uint32 {
	var rawLevels []map[uint32]uint32
	currentGraph := g

	for level := 0; level < maxLevels; level++ {
		communities := Louvain(currentGraph)
		numCommunities := countUniqueCommunities(communities)

		if numCommunities >= currentGraph.NodeCount() {
			break
		}

		rawLevels = append(rawLevels, communities)

		if numCommunities <= 1 {
			break
		}

		currentGraph = coarsen(currentGraph, communities)
	}

	if len(rawLevels) == 0 {
		return nil
	}

	// Build output levels where each level maps original node IDs to community IDs.
	// levels[0] is the coarsest (fewest communities), levels[last] is the finest.
	// rawLevels[0] maps original node IDs -> fine community IDs
	// rawLevels[1] maps fine community IDs -> coarser community IDs
	// etc.
	// We reverse so that index 0 = coarsest, last = finest.
	levels := make([]map[uint32]uint32, len(rawLevels))

	// The finest level is rawLevels[0] which already maps original node IDs.
	levels[len(levels)-1] = rawLevels[0]

	// For coarser levels, compose the mappings.
	for i := len(rawLevels) - 1; i >= 1; i-- {
		composed := make(map[uint32]uint32)
		for nodeID, comID := range rawLevels[0] {
			// Walk up through rawLevels[1..i] to get the coarser community
			cur := comID
			for j := 1; j <= i; j++ {
				cur = rawLevels[j][cur]
			}
			composed[nodeID] = cur
		}
		levels[len(rawLevels)-1-i] = composed
	}

	return levels
}

func coarsen(g *graph.Graph, communities map[uint32]uint32) *graph.Graph {
	cg := graph.NewGraph()
	comSet := make(map[uint32]bool)
	for _, com := range communities {
		if !comSet[com] {
			cg.AddNode(graph.Node{ID: com})
			comSet[com] = true
		}
	}

	edgeWeights := make(map[[2]uint32]float32)
	for id := range g.Nodes() {
		fromCom := communities[id]
		for _, nb := range g.OutNeighbors(id) {
			toCom := communities[nb]
			if fromCom != toCom {
				key := [2]uint32{fromCom, toCom}
				edgeWeights[key] += g.EdgeWeight(id, nb)
			}
		}
	}

	for key, w := range edgeWeights {
		cg.AddEdge(key[0], key[1], w)
	}

	return cg
}

func countUniqueCommunities(m map[uint32]uint32) int {
	seen := make(map[uint32]bool)
	for _, v := range m {
		seen[v] = true
	}
	return len(seen)
}
