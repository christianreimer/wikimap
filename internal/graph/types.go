package graph

type Node struct {
	ID        uint32
	Title     string
	PageViews uint32
}

type Edge struct {
	From   uint32
	To     uint32
	Weight float32
}

type Graph struct {
	nodes     map[uint32]Node
	outEdges  map[uint32][]uint32
	inEdges   map[uint32][]uint32
	weights   map[[2]uint32]float32
	edgeCount int
}

func NewGraph() *Graph {
	return &Graph{
		nodes:    make(map[uint32]Node),
		outEdges: make(map[uint32][]uint32),
		inEdges:  make(map[uint32][]uint32),
		weights:  make(map[[2]uint32]float32),
	}
}

func (g *Graph) AddNode(n Node) {
	g.nodes[n.ID] = n
}

func (g *Graph) Node(id uint32) (Node, bool) {
	n, ok := g.nodes[id]
	return n, ok
}

func (g *Graph) NodeCount() int {
	return len(g.nodes)
}

func (g *Graph) EdgeCount() int {
	return g.edgeCount
}

func (g *Graph) AddEdge(from, to uint32, weight float32) {
	g.outEdges[from] = append(g.outEdges[from], to)
	g.inEdges[to] = append(g.inEdges[to], from)
	g.weights[[2]uint32{from, to}] = weight
	g.edgeCount++
}

func (g *Graph) OutNeighbors(id uint32) []uint32 {
	return g.outEdges[id]
}

func (g *Graph) InNeighbors(id uint32) []uint32 {
	return g.inEdges[id]
}

func (g *Graph) EdgeWeight(from, to uint32) float32 {
	return g.weights[[2]uint32{from, to}]
}

func (g *Graph) Nodes() map[uint32]Node {
	return g.nodes
}
