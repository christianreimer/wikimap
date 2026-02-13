package cluster

import "github.com/creimer/wikimap/internal/graph"

func Louvain(g *graph.Graph) map[uint32]uint32 {
	community := make(map[uint32]uint32)
	for id := range g.Nodes() {
		community[id] = id
	}

	totalWeight := float64(g.EdgeCount())
	if totalWeight == 0 {
		return community
	}

	degree := make(map[uint32]float64)
	for id := range g.Nodes() {
		degree[id] = float64(len(g.OutNeighbors(id)) + len(g.InNeighbors(id)))
	}

	communityTotalDegree := make(map[uint32]float64)
	for id := range g.Nodes() {
		communityTotalDegree[community[id]] += degree[id]
	}

	improved := true
	for improved {
		improved = false
		for id := range g.Nodes() {
			currentCom := community[id]

			neighborComWeight := make(map[uint32]float64)
			for _, nb := range g.OutNeighbors(id) {
				neighborComWeight[community[nb]] += float64(g.EdgeWeight(id, nb))
			}
			for _, nb := range g.InNeighbors(id) {
				neighborComWeight[community[nb]] += float64(g.EdgeWeight(nb, id))
			}

			communityTotalDegree[currentCom] -= degree[id]

			bestCom := currentCom
			bestGain := 0.0

			for com, wc := range neighborComWeight {
				gain := wc/totalWeight - degree[id]*communityTotalDegree[com]/(2.0*totalWeight*totalWeight)
				if gain > bestGain {
					bestGain = gain
					bestCom = com
				}
			}

			community[id] = bestCom
			communityTotalDegree[bestCom] += degree[id]

			if bestCom != currentCom {
				improved = true
			}
		}
	}

	remap := make(map[uint32]uint32)
	var nextID uint32
	result := make(map[uint32]uint32)
	for id, com := range community {
		if _, ok := remap[com]; !ok {
			remap[com] = nextID
			nextID++
		}
		result[id] = remap[com]
	}

	return result
}
