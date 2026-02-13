package graph

func PageRank(g *Graph, damping float64, iterations int) map[uint32]float64 {
	n := float64(g.NodeCount())
	if n == 0 {
		return nil
	}

	rank := make(map[uint32]float64, g.NodeCount())
	for id := range g.nodes {
		rank[id] = 1.0 / n
	}

	for iter := 0; iter < iterations; iter++ {
		newRank := make(map[uint32]float64, g.NodeCount())
		for id := range g.nodes {
			newRank[id] = (1.0 - damping) / n
		}

		for id := range g.nodes {
			outDeg := len(g.outEdges[id])
			if outDeg == 0 {
				share := rank[id] / n
				for other := range g.nodes {
					newRank[other] += damping * share
				}
			} else {
				share := rank[id] / float64(outDeg)
				for _, neighbor := range g.outEdges[id] {
					newRank[neighbor] += damping * share
				}
			}
		}

		rank = newRank
	}

	return rank
}
