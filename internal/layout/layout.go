package layout

import (
	"math"
	"math/rand"

	"github.com/creimer/wikimap/internal/graph"
)

type Point struct {
	X, Y float32
}

func ForceDirected(g *graph.Graph, iterations int) map[uint32]Point {
	nodes := g.Nodes()
	n := len(nodes)
	if n == 0 {
		return nil
	}

	area := float64(n)
	k := math.Sqrt(area / float64(n))

	pos := make(map[uint32][2]float64, n)
	rng := rand.New(rand.NewSource(42))
	for id := range nodes {
		pos[id] = [2]float64{rng.Float64() * math.Sqrt(area), rng.Float64() * math.Sqrt(area)}
	}

	ids := make([]uint32, 0, n)
	for id := range nodes {
		ids = append(ids, id)
	}

	temp := math.Sqrt(area) / 2

	for iter := 0; iter < iterations; iter++ {
		disp := make(map[uint32][2]float64, n)

		for i := 0; i < len(ids); i++ {
			for j := i + 1; j < len(ids); j++ {
				u, v := ids[i], ids[j]
				dx := pos[u][0] - pos[v][0]
				dy := pos[u][1] - pos[v][1]
				dist := math.Max(math.Sqrt(dx*dx+dy*dy), 0.001)
				force := k * k / dist
				fx := dx / dist * force
				fy := dy / dist * force
				du := disp[u]
				du[0] += fx
				du[1] += fy
				disp[u] = du
				dv := disp[v]
				dv[0] -= fx
				dv[1] -= fy
				disp[v] = dv
			}
		}

		for _, u := range ids {
			for _, v := range g.OutNeighbors(u) {
				dx := pos[u][0] - pos[v][0]
				dy := pos[u][1] - pos[v][1]
				dist := math.Max(math.Sqrt(dx*dx+dy*dy), 0.001)
				force := dist * dist / k
				fx := dx / dist * force
				fy := dy / dist * force
				du := disp[u]
				du[0] -= fx
				du[1] -= fy
				disp[u] = du
				dv := disp[v]
				dv[0] += fx
				dv[1] += fy
				disp[v] = dv
			}
		}

		for _, id := range ids {
			dx := disp[id][0]
			dy := disp[id][1]
			dist := math.Max(math.Sqrt(dx*dx+dy*dy), 0.001)
			scale := math.Min(dist, temp) / dist
			p := pos[id]
			p[0] += dx * scale
			p[1] += dy * scale
			pos[id] = p
		}

		temp *= 0.95
	}

	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, p := range pos {
		minX = math.Min(minX, p[0])
		minY = math.Min(minY, p[1])
		maxX = math.Max(maxX, p[0])
		maxY = math.Max(maxY, p[1])
	}

	rangeX := maxX - minX
	rangeY := maxY - minY
	if rangeX < 0.001 {
		rangeX = 1
	}
	if rangeY < 0.001 {
		rangeY = 1
	}

	result := make(map[uint32]Point, n)
	for id, p := range pos {
		result[id] = Point{
			X: float32((p[0] - minX) / rangeX),
			Y: float32((p[1] - minY) / rangeY),
		}
	}
	return result
}
