package tiles

import (
	"math"

	"github.com/creimer/wikimap/internal/graph"
	"github.com/creimer/wikimap/internal/layout"
)

type TileNode struct {
	ID         uint32  `json:"id"`
	X          float32 `json:"x"`
	Y          float32 `json:"y"`
	Label      string  `json:"label"`
	Importance float32 `json:"importance"`
	ClusterID  uint32  `json:"clusterId"`
}

type TileEdge struct {
	FromID uint32  `json:"fromId"`
	ToID   uint32  `json:"toId"`
	Weight float32 `json:"weight"`
	ToX    float32 `json:"toX"`
	ToY    float32 `json:"toY"`
}

type GeneratedTile struct {
	Z     int        `json:"z"`
	X     int        `json:"x"`
	Y     int        `json:"y"`
	Nodes []TileNode `json:"nodes"`
	Edges []TileEdge `json:"edges"`
}

func WorldToTile(wx, wy float32, zoom int) (int, int) {
	size := 1 << zoom
	tx := int(wx * float32(size))
	ty := int(wy * float32(size))
	if tx >= size {
		tx = size - 1
	}
	if ty >= size {
		ty = size - 1
	}
	return tx, ty
}

func GenerateTiles(
	g *graph.Graph,
	positions map[uint32]layout.Point,
	importance map[uint32]float64,
	communities map[uint32]uint32,
	minZoom, maxZoom int,
) []GeneratedTile {
	var result []GeneratedTile

	for zoom := minZoom; zoom <= maxZoom; zoom++ {
		threshold := importanceThreshold(zoom, maxZoom)
		tileMap := make(map[[3]int]*GeneratedTile)

		for id, pos := range positions {
			imp := importance[id]
			if imp < threshold {
				continue
			}

			tx, ty := WorldToTile(pos.X, pos.Y, zoom)
			key := [3]int{zoom, tx, ty}
			tile := tileMap[key]
			if tile == nil {
				tile = &GeneratedTile{Z: zoom, X: tx, Y: ty}
				tileMap[key] = tile
			}

			sizeInt := 1 << zoom
			size := float32(sizeInt)
			localX := pos.X*size - float32(tx)
			localY := pos.Y*size - float32(ty)

			node, _ := g.Node(id)
			tile.Nodes = append(tile.Nodes, TileNode{
				ID:         id,
				X:          localX,
				Y:          localY,
				Label:      node.Title,
				Importance: float32(imp),
				ClusterID:  communities[id],
			})
		}

		visibleNodes := make(map[uint32]bool)
		for id, imp := range importance {
			if imp >= threshold {
				visibleNodes[id] = true
			}
		}

		for fromID := range visibleNodes {
			fromPos := positions[fromID]
			fromTX, fromTY := WorldToTile(fromPos.X, fromPos.Y, zoom)
			key := [3]int{zoom, fromTX, fromTY}
			tile := tileMap[key]
			if tile == nil {
				continue
			}

			for _, toID := range g.OutNeighbors(fromID) {
				if !visibleNodes[toID] {
					continue
				}
				toPos := positions[toID]
				sizeInt := 1 << zoom
				size := float32(sizeInt)
				toLocalX := toPos.X*size - float32(fromTX)
				toLocalY := toPos.Y*size - float32(fromTY)

				tile.Edges = append(tile.Edges, TileEdge{
					FromID: fromID,
					ToID:   toID,
					Weight: g.EdgeWeight(fromID, toID),
					ToX:    toLocalX,
					ToY:    toLocalY,
				})
			}
		}

		for _, tile := range tileMap {
			result = append(result, *tile)
		}
	}

	return result
}

func importanceThreshold(zoom, maxZoom int) float64 {
	if maxZoom == 0 {
		return 0
	}
	return math.Pow(10, -float64(zoom+1)/float64(maxZoom+1)*3)
}
