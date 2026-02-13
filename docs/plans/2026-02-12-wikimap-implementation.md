# WikiMap Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Google Maps-style interactive visualization of Wikipedia's link graph, where articles are cities, links are roads, and topic clusters form continents/countries/states.

**Architecture:** Go offline pipeline parses Wikipedia dump, builds a directed graph, detects communities, computes spatial layout, and generates vector tiles. Go HTTP server serves tiles. Vanilla TypeScript + WebGL frontend renders the map with pan/zoom/search.

**Tech Stack:** Go (pipeline + server), Vanilla TypeScript + WebGL (frontend), Protocol Buffers (tile format), Wikipedia dump (data source)

---

## Phase 1: Project Scaffolding

### Task 1: Initialize Go module and directory structure

**Files:**
- Create: `go.mod`
- Create: `cmd/pipeline/main.go`
- Create: `cmd/server/main.go`

**Step 1: Initialize Go module**

Run: `cd /Users/creimer/code/wikimap && go mod init github.com/creimer/wikimap`

**Step 2: Create directory structure**

```bash
mkdir -p cmd/pipeline cmd/server
mkdir -p internal/parser internal/graph internal/cluster internal/layout internal/tiles internal/search internal/server
mkdir -p proto
```

**Step 3: Create placeholder main files**

`cmd/pipeline/main.go`:
```go
package main

import "fmt"

func main() {
	fmt.Println("wikimap pipeline")
}
```

`cmd/server/main.go`:
```go
package main

import "fmt"

func main() {
	fmt.Println("wikimap server")
}
```

**Step 4: Verify builds**

Run: `go build ./cmd/pipeline && go build ./cmd/server`
Expected: No errors

**Step 5: Commit**

```bash
git add cmd/ internal/ proto/ go.mod
git commit -m "scaffold: initialize Go module and directory structure"
```

---

### Task 2: Initialize TypeScript frontend project

**Files:**
- Create: `web/package.json`
- Create: `web/tsconfig.json`
- Create: `web/index.html`
- Create: `web/src/main.ts`
- Create: `web/vite.config.ts`

**Step 1: Create package.json**

`web/package.json`:
```json
{
  "name": "wikimap-web",
  "private": true,
  "version": "0.0.1",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "test": "vitest run",
    "test:watch": "vitest"
  },
  "devDependencies": {
    "typescript": "^5.7.0",
    "vite": "^6.0.0",
    "vitest": "^3.0.0"
  }
}
```

**Step 2: Create tsconfig.json**

`web/tsconfig.json`:
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "isolatedModules": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "sourceMap": true
  },
  "include": ["src"]
}
```

**Step 3: Create vite.config.ts**

`web/vite.config.ts`:
```typescript
import { defineConfig } from 'vite';

export default defineConfig({
  root: '.',
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/tiles': 'http://localhost:8080',
    },
  },
});
```

**Step 4: Create index.html**

`web/index.html`:
```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>WikiMap</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    html, body { width: 100%; height: 100%; overflow: hidden; }
    canvas { display: block; width: 100%; height: 100%; }
  </style>
</head>
<body>
  <canvas id="map"></canvas>
  <script type="module" src="/src/main.ts"></script>
</body>
</html>
```

**Step 5: Create main.ts**

`web/src/main.ts`:
```typescript
const canvas = document.getElementById('map') as HTMLCanvasElement;
canvas.width = window.innerWidth;
canvas.height = window.innerHeight;
console.log('WikiMap initialized', canvas.width, canvas.height);
```

**Step 6: Install dependencies and verify**

Run: `cd /Users/creimer/code/wikimap/web && npm install`
Run: `cd /Users/creimer/code/wikimap/web && npx tsc --noEmit`
Expected: No errors

**Step 7: Create .gitignore**

`web/.gitignore`:
```
node_modules/
dist/
```

**Step 8: Commit**

```bash
git add web/
git commit -m "scaffold: initialize TypeScript frontend with Vite"
```

---

## Phase 2: Wikipedia Dump Parser

### Task 3: Define graph data types

**Files:**
- Create: `internal/graph/types.go`
- Create: `internal/graph/types_test.go`

**Step 1: Write test for basic types**

`internal/graph/types_test.go`:
```go
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
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/graph/...`
Expected: FAIL (types not defined)

**Step 3: Implement graph types**

`internal/graph/types.go`:
```go
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
	nodes    map[uint32]Node
	outEdges map[uint32][]uint32
	inEdges  map[uint32][]uint32
	weights  map[[2]uint32]float32
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
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/graph/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/graph/
git commit -m "feat: add graph data types with adjacency list"
```

---

### Task 4: Wikipedia XML dump parser

**Files:**
- Create: `internal/parser/parser.go`
- Create: `internal/parser/parser_test.go`
- Create: `internal/parser/testdata/small.xml`

**Step 1: Create test fixture**

`internal/parser/testdata/small.xml`:
```xml
<mediawiki>
  <page>
    <title>Cat</title>
    <ns>0</ns>
    <id>100</id>
    <revision>
      <text bytes="1234">The '''cat''' is a [[domestication|domestic]] [[species]] of small [[carnivore|carnivorous]] [[mammal]]. See also [[Dog]].</text>
    </revision>
  </page>
  <page>
    <title>Dog</title>
    <ns>0</ns>
    <id>200</id>
    <revision>
      <text bytes="999">The '''dog''' is a [[domestication|domesticated]] descendant of the [[wolf]]. See also [[Cat]].</text>
    </revision>
  </page>
  <page>
    <title>Wikipedia:Manual of Style</title>
    <ns>4</ns>
    <id>300</id>
    <revision>
      <text bytes="500">This is a Wikipedia namespace page that should be skipped.</text>
    </revision>
  </page>
  <page>
    <title>Mammal</title>
    <ns>0</ns>
    <id>400</id>
    <revision>
      <text bytes="800">'''Mammals''' are a group of [[vertebrate]] animals. Includes [[Cat]] and [[Dog]].</text>
    </revision>
  </page>
</mediawiki>
```

**Step 2: Write parser test**

`internal/parser/parser_test.go`:
```go
package parser

import (
	"os"
	"testing"
)

func TestParseArticles(t *testing.T) {
	f, err := os.Open("testdata/small.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	articles, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	// Should skip ns=4 page (Wikipedia namespace)
	if len(articles) != 3 {
		t.Fatalf("expected 3 articles, got %d", len(articles))
	}

	cat := findByTitle(articles, "Cat")
	if cat == nil {
		t.Fatal("expected to find article 'Cat'")
	}
	if cat.ID != 100 {
		t.Fatalf("expected Cat ID=100, got %d", cat.ID)
	}

	// Cat should link to: Domestication, Species, Carnivore, Mammal, Dog
	expectedLinks := []string{"Domestication", "Species", "Carnivore", "Mammal", "Dog"}
	if len(cat.Links) != len(expectedLinks) {
		t.Fatalf("expected %d links from Cat, got %d: %v", len(expectedLinks), len(cat.Links), cat.Links)
	}
	for _, want := range expectedLinks {
		found := false
		for _, got := range cat.Links {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected Cat to link to %q, links: %v", want, cat.Links)
		}
	}
}

func findByTitle(articles []Article, title string) *Article {
	for i := range articles {
		if articles[i].Title == title {
			return &articles[i]
		}
	}
	return nil
}
```

**Step 3: Run test to verify it fails**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/parser/...`
Expected: FAIL

**Step 4: Implement parser**

`internal/parser/parser.go`:
```go
package parser

import (
	"encoding/xml"
	"io"
	"regexp"
	"strings"
)

type Article struct {
	ID    uint32
	Title string
	Links []string
}

var linkRegex = regexp.MustCompile(`\[\[([^\]|#]+)(?:[|#][^\]]*)?\]\]`)

func Parse(r io.Reader) ([]Article, error) {
	decoder := xml.NewDecoder(r)
	var articles []Article

	var inPage, inRevision bool
	var current struct {
		title string
		ns    string
		id    uint32
		text  string
	}

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "page":
				inPage = true
				current.title = ""
				current.ns = ""
				current.id = 0
				current.text = ""
			case "revision":
				inRevision = true
			case "title", "ns", "id", "text":
				if !inPage {
					continue
				}
				var content string
				if err := decoder.DecodeElement(&content, &t); err != nil {
					return nil, err
				}
				switch t.Name.Local {
				case "title":
					current.title = content
				case "ns":
					current.ns = content
				case "id":
					if !inRevision {
						var id uint32
						for _, c := range content {
							id = id*10 + uint32(c-'0')
						}
						current.id = id
					}
				case "text":
					current.text = content
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "revision":
				inRevision = false
			case "page":
				inPage = false
				if current.ns == "0" {
					links := extractLinks(current.text)
					articles = append(articles, Article{
						ID:    current.id,
						Title: current.title,
						Links: links,
					})
				}
			}
		}
	}
	return articles, nil
}

func extractLinks(text string) []string {
	matches := linkRegex.FindAllStringSubmatch(text, -1)
	seen := make(map[string]bool)
	var links []string
	for _, m := range matches {
		target := strings.TrimSpace(m[1])
		// Capitalize first letter to normalize
		if len(target) > 0 {
			target = strings.ToUpper(target[:1]) + target[1:]
		}
		if !seen[target] {
			seen[target] = true
			links = append(links, target)
		}
	}
	return links
}
```

**Step 5: Run test to verify it passes**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/parser/...`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/parser/
git commit -m "feat: add Wikipedia XML dump parser with wikilink extraction"
```

---

### Task 5: Build graph from parsed articles

**Files:**
- Create: `internal/parser/graphbuilder.go`
- Create: `internal/parser/graphbuilder_test.go`

**Step 1: Write test**

`internal/parser/graphbuilder_test.go`:
```go
package parser

import (
	"os"
	"testing"
)

func TestBuildGraph(t *testing.T) {
	f, err := os.Open("testdata/small.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	articles, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	g := BuildGraph(articles)

	// 3 articles: Cat(100), Dog(200), Mammal(400)
	if g.NodeCount() != 3 {
		t.Fatalf("expected 3 nodes, got %d", g.NodeCount())
	}

	// Cat links to Dog and Mammal (other targets don't exist as articles)
	outCat := g.OutNeighbors(100)
	if len(outCat) != 2 {
		t.Fatalf("expected Cat to have 2 out-edges (to existing articles), got %d", len(outCat))
	}

	// Dog links to Cat
	outDog := g.OutNeighbors(200)
	if len(outDog) != 1 {
		t.Fatalf("expected Dog to have 1 out-edge, got %d", len(outDog))
	}

	// Mammal links to Cat and Dog
	outMammal := g.OutNeighbors(400)
	if len(outMammal) != 2 {
		t.Fatalf("expected Mammal to have 2 out-edges, got %d", len(outMammal))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/parser/...`
Expected: FAIL

**Step 3: Implement graph builder**

`internal/parser/graphbuilder.go`:
```go
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
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/parser/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/parser/graphbuilder.go internal/parser/graphbuilder_test.go
git commit -m "feat: build graph from parsed Wikipedia articles"
```

---

## Phase 3: Graph Metrics

### Task 6: PageRank computation

**Files:**
- Create: `internal/graph/pagerank.go`
- Create: `internal/graph/pagerank_test.go`

**Step 1: Write test**

`internal/graph/pagerank_test.go`:
```go
package graph

import (
	"math"
	"testing"
)

func TestPageRank(t *testing.T) {
	// Simple triangle: A->B->C->A
	g := NewGraph()
	g.AddNode(Node{ID: 1, Title: "A"})
	g.AddNode(Node{ID: 2, Title: "B"})
	g.AddNode(Node{ID: 3, Title: "C"})
	g.AddEdge(1, 2, 1.0)
	g.AddEdge(2, 3, 1.0)
	g.AddEdge(3, 1, 1.0)

	ranks := PageRank(g, 0.85, 100)

	// In a symmetric cycle, all ranks should be equal (~0.333)
	for id, rank := range ranks {
		if math.Abs(rank-1.0/3.0) > 0.01 {
			t.Errorf("node %d: expected rank ~0.333, got %.4f", id, rank)
		}
	}
}

func TestPageRankStarGraph(t *testing.T) {
	// Star: B->A, C->A, D->A (A should have highest rank)
	g := NewGraph()
	g.AddNode(Node{ID: 1, Title: "A"})
	g.AddNode(Node{ID: 2, Title: "B"})
	g.AddNode(Node{ID: 3, Title: "C"})
	g.AddNode(Node{ID: 4, Title: "D"})
	g.AddEdge(2, 1, 1.0)
	g.AddEdge(3, 1, 1.0)
	g.AddEdge(4, 1, 1.0)

	ranks := PageRank(g, 0.85, 100)

	if ranks[1] <= ranks[2] || ranks[1] <= ranks[3] || ranks[1] <= ranks[4] {
		t.Errorf("expected node A to have highest rank, got: %v", ranks)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/graph/...`
Expected: FAIL

**Step 3: Implement PageRank**

`internal/graph/pagerank.go`:
```go
package graph

// PageRank computes PageRank scores using power iteration.
// damping is typically 0.85, iterations ~100.
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
				// Dangling node: distribute rank evenly
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
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/graph/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/graph/pagerank.go internal/graph/pagerank_test.go
git commit -m "feat: add PageRank computation via power iteration"
```

---

## Phase 4: Community Detection

### Task 7: Louvain community detection

**Files:**
- Create: `internal/cluster/louvain.go`
- Create: `internal/cluster/louvain_test.go`

**Step 1: Write test**

`internal/cluster/louvain_test.go`:
```go
package cluster

import (
	"testing"

	"github.com/creimer/wikimap/internal/graph"
)

func TestLouvainTwoClusters(t *testing.T) {
	// Two dense cliques connected by a single edge
	g := graph.NewGraph()
	// Clique 1: nodes 1,2,3
	for _, id := range []uint32{1, 2, 3} {
		g.AddNode(graph.Node{ID: id, Title: ""})
	}
	// Clique 2: nodes 4,5,6
	for _, id := range []uint32{4, 5, 6} {
		g.AddNode(graph.Node{ID: id, Title: ""})
	}
	// Dense connections within clique 1
	for _, e := range [][2]uint32{{1, 2}, {2, 1}, {1, 3}, {3, 1}, {2, 3}, {3, 2}} {
		g.AddEdge(e[0], e[1], 1.0)
	}
	// Dense connections within clique 2
	for _, e := range [][2]uint32{{4, 5}, {5, 4}, {4, 6}, {6, 4}, {5, 6}, {6, 5}} {
		g.AddEdge(e[0], e[1], 1.0)
	}
	// Single bridge
	g.AddEdge(3, 4, 1.0)
	g.AddEdge(4, 3, 1.0)

	communities := Louvain(g)

	// Nodes in the same clique should be in the same community
	if communities[1] != communities[2] || communities[1] != communities[3] {
		t.Errorf("clique 1 should be one community: %v", communities)
	}
	if communities[4] != communities[5] || communities[4] != communities[6] {
		t.Errorf("clique 2 should be one community: %v", communities)
	}
	// The two cliques should be in different communities
	if communities[1] == communities[4] {
		t.Errorf("cliques should be in different communities: %v", communities)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/cluster/...`
Expected: FAIL

**Step 3: Implement Louvain**

`internal/cluster/louvain.go`:
```go
package cluster

import "github.com/creimer/wikimap/internal/graph"

// Louvain runs the Louvain modularity optimization algorithm.
// Returns a map from node ID to community ID.
func Louvain(g *graph.Graph) map[uint32]uint32 {
	// Initialize: each node in its own community
	community := make(map[uint32]uint32)
	for id := range g.Nodes() {
		community[id] = id
	}

	totalWeight := float64(g.EdgeCount())
	if totalWeight == 0 {
		return community
	}

	// Compute weighted degree for each node (sum of edge weights)
	degree := make(map[uint32]float64)
	for id := range g.Nodes() {
		degree[id] = float64(len(g.OutNeighbors(id)) + len(g.InNeighbors(id)))
	}

	// Sum of weights inside each community
	communityInternalWeight := make(map[uint32]float64)
	// Sum of degrees of nodes in each community
	communityTotalDegree := make(map[uint32]float64)
	for id := range g.Nodes() {
		communityTotalDegree[community[id]] += degree[id]
	}

	improved := true
	for improved {
		improved = false
		for id := range g.Nodes() {
			currentCom := community[id]

			// Calculate weights to neighboring communities
			neighborComWeight := make(map[uint32]float64)
			for _, nb := range g.OutNeighbors(id) {
				neighborComWeight[community[nb]] += float64(g.EdgeWeight(id, nb))
			}
			for _, nb := range g.InNeighbors(id) {
				neighborComWeight[community[nb]] += float64(g.EdgeWeight(nb, id))
			}

			// Remove node from current community
			communityTotalDegree[currentCom] -= degree[id]

			bestCom := currentCom
			bestGain := 0.0

			for com, wc := range neighborComWeight {
				// Modularity gain from moving node to community com
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

	// Normalize community IDs to sequential values
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
	_ = communityInternalWeight

	return result
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/cluster/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/cluster/
git commit -m "feat: add Louvain community detection"
```

---

### Task 8: Hierarchical clustering

**Files:**
- Create: `internal/cluster/hierarchy.go`
- Create: `internal/cluster/hierarchy_test.go`

**Step 1: Write test**

`internal/cluster/hierarchy_test.go`:
```go
package cluster

import (
	"testing"

	"github.com/creimer/wikimap/internal/graph"
)

type HierarchyLevel struct {
	NodeToCommunity map[uint32]uint32
	NumCommunities  int
}

func TestBuildHierarchy(t *testing.T) {
	g := graph.NewGraph()
	for i := uint32(1); i <= 12; i++ {
		g.AddNode(graph.Node{ID: i})
	}
	// 3 cliques of 4 nodes each
	for _, clique := range [][]uint32{{1, 2, 3, 4}, {5, 6, 7, 8}, {9, 10, 11, 12}} {
		for i := 0; i < len(clique); i++ {
			for j := i + 1; j < len(clique); j++ {
				g.AddEdge(clique[i], clique[j], 1.0)
				g.AddEdge(clique[j], clique[i], 1.0)
			}
		}
	}
	// Bridges
	g.AddEdge(4, 5, 1.0)
	g.AddEdge(5, 4, 1.0)
	g.AddEdge(8, 9, 1.0)
	g.AddEdge(9, 8, 1.0)

	levels := BuildHierarchy(g, 3)

	if len(levels) < 2 {
		t.Fatalf("expected at least 2 hierarchy levels, got %d", len(levels))
	}

	// Bottom level should have more communities than top level
	bottomCount := countUnique(levels[len(levels)-1])
	topCount := countUnique(levels[0])
	if bottomCount <= topCount {
		t.Errorf("bottom level (%d communities) should have more than top level (%d)", bottomCount, topCount)
	}
}

func countUnique(m map[uint32]uint32) int {
	seen := make(map[uint32]bool)
	for _, v := range m {
		seen[v] = true
	}
	return len(seen)
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/cluster/...`
Expected: FAIL

**Step 3: Implement hierarchical clustering**

`internal/cluster/hierarchy.go`:
```go
package cluster

import "github.com/creimer/wikimap/internal/graph"

// BuildHierarchy runs Louvain repeatedly on coarsened graphs to produce
// multiple levels of clustering. Returns levels from coarsest (top) to finest (bottom).
func BuildHierarchy(g *graph.Graph, maxLevels int) []map[uint32]uint32 {
	var levels []map[uint32]uint32
	currentGraph := g

	for level := 0; level < maxLevels; level++ {
		communities := Louvain(currentGraph)
		numCommunities := countUniqueCommunities(communities)

		if numCommunities >= currentGraph.NodeCount() {
			break // No further compression possible
		}

		levels = append(levels, communities)

		if numCommunities <= 1 {
			break
		}

		// Coarsen: create a new graph where each community is a node
		currentGraph = coarsen(currentGraph, communities)
	}

	// Expand community assignments back to original node IDs
	if len(levels) > 1 {
		for i := len(levels) - 2; i >= 0; i-- {
			expanded := make(map[uint32]uint32)
			bottom := levels[len(levels)-1]
			for nodeID := range bottom {
				// Walk up the chain
				comID := nodeID
				for j := len(levels) - 1; j > i; j-- {
					comID = levels[j][comID]
				}
				expanded[nodeID] = levels[i][comID]
			}
			levels[i] = expanded
		}
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
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/cluster/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/cluster/hierarchy.go internal/cluster/hierarchy_test.go
git commit -m "feat: add hierarchical clustering via repeated Louvain coarsening"
```

---

## Phase 5: Spatial Layout

### Task 9: Force-directed layout engine

**Files:**
- Create: `internal/layout/layout.go`
- Create: `internal/layout/layout_test.go`

**Step 1: Write test**

`internal/layout/layout_test.go`:
```go
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

	// Connected nodes should end up at a finite distance
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
	// No edges -- should repel

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

	// All positions should be in [0, 1] after normalization
	for id, pos := range positions {
		if pos.X < 0 || pos.X > 1 || pos.Y < 0 || pos.Y > 1 {
			t.Errorf("node %d position (%.4f, %.4f) outside [0,1]", id, pos.X, pos.Y)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/layout/...`
Expected: FAIL

**Step 3: Implement force-directed layout**

`internal/layout/layout.go`:
```go
package layout

import (
	"math"
	"math/rand"

	"github.com/creimer/wikimap/internal/graph"
)

type Point struct {
	X, Y float32
}

// ForceDirected computes 2D positions for all graph nodes using
// Fruchterman-Reingold force-directed placement.
func ForceDirected(g *graph.Graph, iterations int) map[uint32]Point {
	nodes := g.Nodes()
	n := len(nodes)
	if n == 0 {
		return nil
	}

	area := float64(n)
	k := math.Sqrt(area / float64(n)) // ideal edge length

	// Initialize random positions
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

		// Repulsive forces between all pairs
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

		// Attractive forces along edges
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

		// Apply displacements with temperature limiting
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

		// Cool
		temp *= 0.95
	}

	// Normalize to [0, 1]
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
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/layout/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/layout/
git commit -m "feat: add Fruchterman-Reingold force-directed layout"
```

---

## Phase 6: Tile Generation

### Task 10: Define tile protobuf schema

**Files:**
- Create: `proto/tile.proto`

**Step 1: Create protobuf schema**

`proto/tile.proto`:
```protobuf
syntax = "proto3";
package wikimap;
option go_package = "github.com/creimer/wikimap/internal/tiles/pb";

message Tile {
  repeated TileNode nodes = 1;
  repeated TileEdge edges = 2;
  repeated TileCluster clusters = 3;
}

message TileNode {
  uint32 id = 1;
  float x = 2;        // position within tile [0, 1]
  float y = 3;
  string label = 4;
  float importance = 5; // 0-1 normalized
  uint32 cluster_id = 6;
}

message TileEdge {
  uint32 from_id = 1;
  uint32 to_id = 2;
  float weight = 3;   // 0-1 normalized
  // Target position if edge crosses tile boundary
  float to_x = 4;
  float to_y = 5;
}

message TileCluster {
  uint32 id = 1;
  string label = 2;
  float center_x = 3;
  float center_y = 4;
  uint32 color = 5;   // RGB packed into uint32
}
```

**Step 2: Install protoc-gen-go and generate**

Run: `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`
Run: `cd /Users/creimer/code/wikimap && protoc --go_out=. --go_opt=paths=source_relative proto/tile.proto`

If protoc is not installed:
Run: `brew install protobuf`

Alternative: use `go get google.golang.org/protobuf` and define the types manually in Go if protoc setup is cumbersome.

**Step 3: Add protobuf dependency**

Run: `cd /Users/creimer/code/wikimap && go get google.golang.org/protobuf`

**Step 4: Commit**

```bash
git add proto/ internal/tiles/pb/ go.mod go.sum
git commit -m "feat: define tile protobuf schema"
```

---

### Task 11: Tile generator

**Files:**
- Create: `internal/tiles/generator.go`
- Create: `internal/tiles/generator_test.go`

**Step 1: Write test**

`internal/tiles/generator_test.go`:
```go
package tiles

import (
	"testing"

	"github.com/creimer/wikimap/internal/graph"
	"github.com/creimer/wikimap/internal/layout"
)

func TestTileCoord(t *testing.T) {
	// At zoom 0, entire world is one tile (0,0)
	tx, ty := WorldToTile(0.5, 0.5, 0)
	if tx != 0 || ty != 0 {
		t.Errorf("zoom 0: expected tile (0,0), got (%d,%d)", tx, ty)
	}

	// At zoom 1, world is 2x2 tiles
	tx, ty = WorldToTile(0.75, 0.75, 1)
	if tx != 1 || ty != 1 {
		t.Errorf("zoom 1: expected tile (1,1), got (%d,%d)", tx, ty)
	}
}

func TestGenerateTiles(t *testing.T) {
	g := graph.NewGraph()
	g.AddNode(graph.Node{ID: 1, Title: "A", PageViews: 1000})
	g.AddNode(graph.Node{ID: 2, Title: "B", PageViews: 500})
	g.AddNode(graph.Node{ID: 3, Title: "C", PageViews: 100})
	g.AddEdge(1, 2, 1.0)
	g.AddEdge(2, 3, 1.0)

	positions := map[uint32]layout.Point{
		1: {X: 0.2, Y: 0.3},
		2: {X: 0.7, Y: 0.8},
		3: {X: 0.1, Y: 0.9},
	}

	importance := map[uint32]float64{
		1: 1.0,
		2: 0.5,
		3: 0.1,
	}

	communities := map[uint32]uint32{
		1: 0, 2: 0, 3: 1,
	}

	tiles := GenerateTiles(g, positions, importance, communities, 0, 2)

	// At zoom 0 we should have 1 tile
	zoom0 := tilesAtZoom(tiles, 0)
	if len(zoom0) != 1 {
		t.Fatalf("expected 1 tile at zoom 0, got %d", len(zoom0))
	}

	// The single zoom-0 tile should contain all 3 nodes (small graph, all important)
	tile := zoom0[0]
	if len(tile.Nodes) != 3 {
		t.Errorf("expected 3 nodes in zoom-0 tile, got %d", len(tile.Nodes))
	}
}

func tilesAtZoom(tiles []GeneratedTile, zoom int) []GeneratedTile {
	var result []GeneratedTile
	for _, t := range tiles {
		if t.Z == zoom {
			result = append(result, t)
		}
	}
	return result
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/tiles/...`
Expected: FAIL

**Step 3: Implement tile generator**

`internal/tiles/generator.go`:
```go
package tiles

import (
	"math"

	"github.com/creimer/wikimap/internal/graph"
	"github.com/creimer/wikimap/internal/layout"
)

type TileNode struct {
	ID         uint32
	X, Y       float32 // position within tile [0,1]
	Label      string
	Importance float32
	ClusterID  uint32
}

type TileEdge struct {
	FromID uint32
	ToID   uint32
	Weight float32
	ToX    float32 // target position in tile coords
	ToY    float32
}

type GeneratedTile struct {
	Z, X, Y int
	Nodes   []TileNode
	Edges   []TileEdge
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

// GenerateTiles creates vector tiles for all zoom levels from minZoom to maxZoom.
func GenerateTiles(
	g *graph.Graph,
	positions map[uint32]layout.Point,
	importance map[uint32]float64,
	communities map[uint32]uint32,
	minZoom, maxZoom int,
) []GeneratedTile {
	var result []GeneratedTile

	for zoom := minZoom; zoom <= maxZoom; zoom++ {
		// Importance threshold decreases with zoom (show more at higher zoom)
		threshold := importanceThreshold(zoom, maxZoom)
		tileMap := make(map[[3]int]*GeneratedTile)

		// Assign visible nodes to tiles
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

			// Convert world coords to tile-local coords
			size := float32(1 << zoom)
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

		// Add edges where both endpoints are visible
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
				size := float32(1 << zoom)
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
	// Exponential decay: zoom 0 shows top nodes, max zoom shows everything
	return math.Pow(10, -float64(zoom)/float64(maxZoom)*3)
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/tiles/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tiles/
git commit -m "feat: add tile generator with zoom-based importance filtering"
```

---

### Task 12: Tile serialization and disk writer

**Files:**
- Create: `internal/tiles/writer.go`
- Create: `internal/tiles/writer_test.go`

**Step 1: Write test**

`internal/tiles/writer_test.go`:
```go
package tiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndReadTile(t *testing.T) {
	dir := t.TempDir()

	tile := GeneratedTile{
		Z: 2, X: 1, Y: 3,
		Nodes: []TileNode{
			{ID: 42, X: 0.5, Y: 0.5, Label: "Test", Importance: 0.9, ClusterID: 1},
		},
		Edges: []TileEdge{
			{FromID: 42, ToID: 99, Weight: 0.7, ToX: 0.8, ToY: 0.2},
		},
	}

	err := WriteTile(dir, tile)
	if err != nil {
		t.Fatal(err)
	}

	// Verify file exists at expected path
	path := filepath.Join(dir, "2", "1", "3.pbf")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("tile file not written at %s", path)
	}

	// Read it back
	loaded, err := ReadTile(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(loaded.Nodes))
	}
	if loaded.Nodes[0].Label != "Test" {
		t.Errorf("expected label 'Test', got %q", loaded.Nodes[0].Label)
	}
	if loaded.Nodes[0].ID != 42 {
		t.Errorf("expected ID 42, got %d", loaded.Nodes[0].ID)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/tiles/...`
Expected: FAIL

**Step 3: Implement tile writer/reader using encoding/json as initial format**

Use JSON initially for simplicity. Switch to protobuf later when optimizing.

`internal/tiles/writer.go`:
```go
package tiles

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteTile writes a tile to disk at {dir}/{z}/{x}/{y}.pbf
func WriteTile(dir string, tile GeneratedTile) error {
	tileDir := filepath.Join(dir, fmt.Sprintf("%d", tile.Z), fmt.Sprintf("%d", tile.X))
	if err := os.MkdirAll(tileDir, 0o755); err != nil {
		return err
	}

	path := filepath.Join(tileDir, fmt.Sprintf("%d.pbf", tile.Y))
	data, err := json.Marshal(tile)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// ReadTile reads a tile from disk.
func ReadTile(path string) (GeneratedTile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return GeneratedTile{}, err
	}

	var tile GeneratedTile
	err = json.Unmarshal(data, &tile)
	return tile, err
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/tiles/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tiles/writer.go internal/tiles/writer_test.go
git commit -m "feat: add tile disk writer/reader with JSON serialization"
```

---

## Phase 7: Tile Server

### Task 13: HTTP tile server

**Files:**
- Create: `internal/server/server.go`
- Create: `internal/server/server_test.go`

**Step 1: Write test**

`internal/server/server_test.go`:
```go
package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/creimer/wikimap/internal/tiles"
)

func TestServeTile(t *testing.T) {
	dir := t.TempDir()

	// Write a test tile
	tile := tiles.GeneratedTile{
		Z: 0, X: 0, Y: 0,
		Nodes: []tiles.TileNode{
			{ID: 1, X: 0.5, Y: 0.5, Label: "Test", Importance: 1.0},
		},
	}
	if err := tiles.WriteTile(dir, tile); err != nil {
		t.Fatal(err)
	}

	srv := NewServer(dir, "")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/tiles/0/0/0.pbf")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServeTile404(t *testing.T) {
	dir := t.TempDir()

	srv := NewServer(dir, "")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/tiles/99/99/99.pbf")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestCORSHeaders(t *testing.T) {
	dir := t.TempDir()

	// Write a tile so we get a 200
	tile := tiles.GeneratedTile{Z: 0, X: 0, Y: 0}
	tiles.WriteTile(dir, tile)

	srv := NewServer(dir, "")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/tiles/0/0/0.pbf")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	cors := resp.Header.Get("Access-Control-Allow-Origin")
	if cors != "*" {
		t.Errorf("expected CORS header '*', got %q", cors)
	}

	cacheControl := resp.Header.Get("Cache-Control")
	if cacheControl == "" {
		t.Error("expected Cache-Control header")
	}
}

// Ensure test fixture directory exists
func init() {
	_ = os.MkdirAll(filepath.Join(os.TempDir(), "wikimap-test"), 0o755)
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/server/...`
Expected: FAIL

**Step 3: Implement server**

`internal/server/server.go`:
```go
package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	tileDir  string
	webDir   string
}

func NewServer(tileDir, webDir string) *Server {
	return &Server{tileDir: tileDir, webDir: webDir}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/tiles/", s.handleTile)

	if s.webDir != "" {
		mux.Handle("/", http.FileServer(http.Dir(s.webDir)))
	}

	return corsMiddleware(mux)
}

func (s *Server) handleTile(w http.ResponseWriter, r *http.Request) {
	// Parse /tiles/{z}/{x}/{y}.pbf
	var z, x, y int
	_, err := fmt.Sscanf(r.URL.Path, "/tiles/%d/%d/%d.pbf", &z, &x, &y)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	path := filepath.Join(s.tileDir, fmt.Sprintf("%d", z), fmt.Sprintf("%d", x), fmt.Sprintf("%d.pbf", y))
	data, err := os.ReadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/x-protobuf")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Write(data)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/creimer/code/wikimap && go test ./internal/server/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/server/
git commit -m "feat: add HTTP tile server with CORS and caching"
```

---

### Task 14: Wire up CLI commands

**Files:**
- Modify: `cmd/pipeline/main.go`
- Modify: `cmd/server/main.go`

**Step 1: Implement pipeline CLI**

`cmd/pipeline/main.go`:
```go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/creimer/wikimap/internal/cluster"
	"github.com/creimer/wikimap/internal/graph"
	"github.com/creimer/wikimap/internal/layout"
	"github.com/creimer/wikimap/internal/parser"
	"github.com/creimer/wikimap/internal/tiles"
)

func main() {
	dumpPath := flag.String("dump", "", "path to Wikipedia XML dump")
	outDir := flag.String("out", "data/tiles", "output directory for tiles")
	maxZoom := flag.Int("max-zoom", 8, "maximum zoom level to generate")
	flag.Parse()

	if *dumpPath == "" {
		fmt.Fprintln(os.Stderr, "usage: pipeline -dump <path-to-dump.xml>")
		os.Exit(1)
	}

	// Stage 1: Parse
	log.Println("Parsing Wikipedia dump...")
	f, err := os.Open(*dumpPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	articles, err := parser.Parse(f)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Parsed %d articles", len(articles))

	// Stage 2: Build graph
	log.Println("Building graph...")
	g := parser.BuildGraph(articles)
	log.Printf("Graph: %d nodes, %d edges", g.NodeCount(), g.EdgeCount())

	// Stage 3: Compute importance
	log.Println("Computing PageRank...")
	importance := graph.PageRank(g, 0.85, 50)

	// Stage 4: Community detection
	log.Println("Detecting communities...")
	communities := cluster.Louvain(g)
	comCount := 0
	seen := make(map[uint32]bool)
	for _, c := range communities {
		if !seen[c] {
			seen[c] = true
			comCount++
		}
	}
	log.Printf("Found %d communities", comCount)

	// Stage 5: Layout
	log.Println("Computing layout...")
	positions := layout.ForceDirected(g, 500)

	// Stage 6: Generate tiles
	log.Println("Generating tiles...")
	allTiles := tiles.GenerateTiles(g, positions, importance, communities, 0, *maxZoom)
	log.Printf("Generated %d tiles", len(allTiles))

	// Stage 7: Write to disk
	log.Println("Writing tiles...")
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatal(err)
	}
	for _, tile := range allTiles {
		if err := tiles.WriteTile(*outDir, tile); err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("Done. Tiles written to %s", *outDir)
}
```

**Step 2: Implement server CLI**

`cmd/server/main.go`:
```go
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/creimer/wikimap/internal/server"
)

func main() {
	tileDir := flag.String("tiles", "data/tiles", "directory containing generated tiles")
	webDir := flag.String("web", "web/dist", "directory containing frontend build")
	port := flag.Int("port", 8080, "port to listen on")
	flag.Parse()

	srv := server.NewServer(*tileDir, *webDir)
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Serving tiles from %s on http://localhost%s", *tileDir, addr)
	log.Fatal(http.ListenAndServe(addr, srv.Handler()))
}
```

**Step 3: Verify both build**

Run: `cd /Users/creimer/code/wikimap && go build ./cmd/pipeline && go build ./cmd/server`
Expected: No errors

**Step 4: Commit**

```bash
git add cmd/
git commit -m "feat: wire up pipeline and server CLI commands"
```

---

## Phase 8: Frontend Map Engine

### Task 15: Camera and viewport math

**Files:**
- Create: `web/src/engine/camera.ts`
- Create: `web/src/engine/camera.test.ts`

**Step 1: Write test**

`web/src/engine/camera.test.ts`:
```typescript
import { describe, it, expect } from 'vitest';
import { Camera, visibleTiles } from './camera';

describe('Camera', () => {
  it('starts at default position', () => {
    const cam = new Camera();
    expect(cam.x).toBe(0.5);
    expect(cam.y).toBe(0.5);
    expect(cam.zoom).toBe(0);
  });

  it('pans by screen delta', () => {
    const cam = new Camera();
    cam.pan(100, 0, 800, 600); // move right by 100px on 800x600 screen
    expect(cam.x).toBeGreaterThan(0.5);
  });

  it('zooms in', () => {
    const cam = new Camera();
    cam.zoomBy(1, 400, 300, 800, 600);
    expect(cam.zoom).toBe(1);
  });

  it('clamps zoom to valid range', () => {
    const cam = new Camera();
    cam.zoomBy(-5, 400, 300, 800, 600);
    expect(cam.zoom).toBe(0);
    cam.zoomBy(100, 400, 300, 800, 600);
    expect(cam.zoom).toBeLessThanOrEqual(18);
  });
});

describe('visibleTiles', () => {
  it('returns single tile at zoom 0', () => {
    const cam = new Camera();
    const result = visibleTiles(cam, 800, 600);
    expect(result).toEqual([{ z: 0, x: 0, y: 0 }]);
  });

  it('returns multiple tiles at higher zoom', () => {
    const cam = new Camera();
    cam.zoom = 2;
    const result = visibleTiles(cam, 800, 600);
    expect(result.length).toBeGreaterThan(1);
    result.forEach(t => {
      expect(t.z).toBe(2);
      expect(t.x).toBeGreaterThanOrEqual(0);
      expect(t.y).toBeGreaterThanOrEqual(0);
    });
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/creimer/code/wikimap/web && npx vitest run src/engine/camera.test.ts`
Expected: FAIL

**Step 3: Implement camera**

`web/src/engine/camera.ts`:
```typescript
export class Camera {
  x = 0.5;  // world coordinate [0,1]
  y = 0.5;
  zoom = 0;

  private static MIN_ZOOM = 0;
  private static MAX_ZOOM = 18;

  pan(dx: number, dy: number, screenW: number, screenH: number): void {
    const scale = Math.pow(2, this.zoom);
    const tileSize = Math.min(screenW, screenH);
    this.x -= dx / (tileSize * scale);
    this.y -= dy / (tileSize * scale);
  }

  zoomBy(delta: number, screenX: number, screenY: number, screenW: number, screenH: number): void {
    const newZoom = Math.max(Camera.MIN_ZOOM, Math.min(Camera.MAX_ZOOM, this.zoom + delta));
    if (newZoom === this.zoom) return;

    // Zoom toward the cursor position
    const tileSize = Math.min(screenW, screenH);
    const scale = Math.pow(2, this.zoom);
    const worldX = this.x + (screenX - screenW / 2) / (tileSize * scale);
    const worldY = this.y + (screenY - screenH / 2) / (tileSize * scale);

    const newScale = Math.pow(2, newZoom);
    this.x = worldX - (screenX - screenW / 2) / (tileSize * newScale);
    this.y = worldY - (screenY - screenH / 2) / (tileSize * newScale);
    this.zoom = newZoom;
  }
}

export interface TileCoord {
  z: number;
  x: number;
  y: number;
}

export function visibleTiles(cam: Camera, screenW: number, screenH: number): TileCoord[] {
  const z = Math.floor(cam.zoom);
  const size = 1 << z;
  const tileSize = Math.min(screenW, screenH);
  const scale = Math.pow(2, cam.zoom);
  const worldPerPixel = 1 / (tileSize * scale);

  const halfW = (screenW / 2) * worldPerPixel;
  const halfH = (screenH / 2) * worldPerPixel;

  const minX = Math.max(0, Math.floor((cam.x - halfW) * size));
  const maxX = Math.min(size - 1, Math.floor((cam.x + halfW) * size));
  const minY = Math.max(0, Math.floor((cam.y - halfH) * size));
  const maxY = Math.min(size - 1, Math.floor((cam.y + halfH) * size));

  const tiles: TileCoord[] = [];
  for (let x = minX; x <= maxX; x++) {
    for (let y = minY; y <= maxY; y++) {
      tiles.push({ z, x, y });
    }
  }
  return tiles;
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/creimer/code/wikimap/web && npx vitest run src/engine/camera.test.ts`
Expected: PASS

**Step 5: Commit**

```bash
git add web/src/engine/
git commit -m "feat: add camera and viewport tile computation"
```

---

### Task 16: Tile loader

**Files:**
- Create: `web/src/engine/tile-loader.ts`
- Create: `web/src/engine/tile-loader.test.ts`
- Create: `web/src/types/tile.ts`

**Step 1: Define tile types**

`web/src/types/tile.ts`:
```typescript
export interface TileNode {
  id: number;
  x: number;
  y: number;
  label: string;
  importance: number;
  clusterId: number;
}

export interface TileEdge {
  fromId: number;
  toId: number;
  weight: number;
  toX: number;
  toY: number;
}

export interface TileData {
  nodes: TileNode[];
  edges: TileEdge[];
}
```

**Step 2: Write tile loader test**

`web/src/engine/tile-loader.test.ts`:
```typescript
import { describe, it, expect, vi } from 'vitest';
import { TileLoader } from './tile-loader';
import type { TileData } from '../types/tile';

describe('TileLoader', () => {
  it('fetches and caches tiles', async () => {
    const mockData: TileData = {
      nodes: [{ id: 1, x: 0.5, y: 0.5, label: 'Test', importance: 1, clusterId: 0 }],
      edges: [],
    };

    const fetcher = vi.fn().mockResolvedValue(mockData);
    const loader = new TileLoader(fetcher);

    const tile = await loader.load(0, 0, 0);
    expect(tile).toEqual(mockData);
    expect(fetcher).toHaveBeenCalledOnce();

    // Second load should use cache
    const tile2 = await loader.load(0, 0, 0);
    expect(tile2).toEqual(mockData);
    expect(fetcher).toHaveBeenCalledOnce(); // still once
  });

  it('evicts tiles beyond cache limit', async () => {
    const fetcher = vi.fn().mockResolvedValue({ nodes: [], edges: [] });
    const loader = new TileLoader(fetcher, 2); // cache limit = 2

    await loader.load(0, 0, 0);
    await loader.load(1, 0, 0);
    await loader.load(1, 1, 0); // should evict (0,0,0)

    expect(fetcher).toHaveBeenCalledTimes(3);

    // Re-fetch evicted tile
    await loader.load(0, 0, 0);
    expect(fetcher).toHaveBeenCalledTimes(4);
  });
});
```

**Step 3: Run test to verify it fails**

Run: `cd /Users/creimer/code/wikimap/web && npx vitest run src/engine/tile-loader.test.ts`
Expected: FAIL

**Step 4: Implement tile loader**

`web/src/engine/tile-loader.ts`:
```typescript
import type { TileData } from '../types/tile';

export type TileFetcher = (z: number, x: number, y: number) => Promise<TileData>;

export class TileLoader {
  private cache = new Map<string, TileData>();
  private pending = new Map<string, Promise<TileData>>();
  private accessOrder: string[] = [];

  constructor(
    private fetcher: TileFetcher,
    private maxCached = 256,
  ) {}

  async load(z: number, x: number, y: number): Promise<TileData> {
    const key = `${z}/${x}/${y}`;

    const cached = this.cache.get(key);
    if (cached) {
      // Move to end of access order
      this.accessOrder = this.accessOrder.filter(k => k !== key);
      this.accessOrder.push(key);
      return cached;
    }

    let pending = this.pending.get(key);
    if (pending) return pending;

    pending = this.fetcher(z, x, y).then(data => {
      this.pending.delete(key);
      this.cache.set(key, data);
      this.accessOrder.push(key);
      this.evict();
      return data;
    });

    this.pending.set(key, pending);
    return pending;
  }

  private evict(): void {
    while (this.cache.size > this.maxCached && this.accessOrder.length > 0) {
      const oldest = this.accessOrder.shift()!;
      this.cache.delete(oldest);
    }
  }
}

export function createFetcher(baseUrl = ''): TileFetcher {
  return async (z, x, y) => {
    const resp = await fetch(`${baseUrl}/tiles/${z}/${x}/${y}.pbf`);
    if (!resp.ok) return { nodes: [], edges: [] };
    return resp.json();
  };
}
```

**Step 5: Run test to verify it passes**

Run: `cd /Users/creimer/code/wikimap/web && npx vitest run src/engine/tile-loader.test.ts`
Expected: PASS

**Step 6: Commit**

```bash
git add web/src/types/ web/src/engine/tile-loader.ts web/src/engine/tile-loader.test.ts
git commit -m "feat: add tile loader with LRU cache"
```

---

## Phase 9: WebGL Renderer

### Task 17: WebGL context and node renderer

**Files:**
- Create: `web/src/renderer/webgl.ts`
- Create: `web/src/renderer/shaders.ts`

**Step 1: Create shader source**

`web/src/renderer/shaders.ts`:
```typescript
export const NODE_VERTEX = `
  attribute vec2 a_position;
  attribute float a_size;
  attribute vec3 a_color;
  attribute float a_opacity;

  uniform mat3 u_projection;

  varying vec3 v_color;
  varying float v_opacity;

  void main() {
    vec3 pos = u_projection * vec3(a_position, 1.0);
    gl_Position = vec4(pos.xy, 0.0, 1.0);
    gl_PointSize = a_size;
    v_color = a_color;
    v_opacity = a_opacity;
  }
`;

export const NODE_FRAGMENT = `
  precision mediump float;
  varying vec3 v_color;
  varying float v_opacity;

  void main() {
    // Circle shape via distance from center
    vec2 coord = gl_PointCoord - vec2(0.5);
    float dist = length(coord);
    if (dist > 0.5) discard;

    float edge = smoothstep(0.45, 0.5, dist);
    gl_FragColor = vec4(v_color, v_opacity * (1.0 - edge));
  }
`;

export const EDGE_VERTEX = `
  attribute vec2 a_position;
  attribute float a_opacity;

  uniform mat3 u_projection;

  varying float v_opacity;

  void main() {
    vec3 pos = u_projection * vec3(a_position, 1.0);
    gl_Position = vec4(pos.xy, 0.0, 1.0);
    v_opacity = a_opacity;
  }
`;

export const EDGE_FRAGMENT = `
  precision mediump float;
  varying float v_opacity;

  void main() {
    gl_FragColor = vec4(0.5, 0.5, 0.5, v_opacity * 0.3);
  }
`;
```

**Step 2: Create WebGL renderer**

`web/src/renderer/webgl.ts`:
```typescript
import { NODE_VERTEX, NODE_FRAGMENT, EDGE_VERTEX, EDGE_FRAGMENT } from './shaders';
import type { Camera } from '../engine/camera';
import type { TileData } from '../types/tile';
import type { TileCoord } from '../engine/camera';

// Cluster colors (10 distinct hues)
const CLUSTER_COLORS: [number, number, number][] = [
  [0.90, 0.30, 0.30], [0.30, 0.70, 0.90], [0.40, 0.80, 0.40],
  [0.95, 0.70, 0.20], [0.70, 0.40, 0.90], [0.90, 0.50, 0.70],
  [0.30, 0.80, 0.75], [0.85, 0.85, 0.30], [0.60, 0.60, 0.60],
  [1.00, 0.60, 0.40],
];

export class MapRenderer {
  private gl: WebGLRenderingContext;
  private nodeProgram: WebGLProgram;
  private edgeProgram: WebGLProgram;

  constructor(private canvas: HTMLCanvasElement) {
    const gl = canvas.getContext('webgl', { alpha: false, antialias: true });
    if (!gl) throw new Error('WebGL not supported');
    this.gl = gl;
    gl.enable(gl.BLEND);
    gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);

    this.nodeProgram = createProgram(gl, NODE_VERTEX, NODE_FRAGMENT);
    this.edgeProgram = createProgram(gl, EDGE_VERTEX, EDGE_FRAGMENT);
  }

  resize(): void {
    this.canvas.width = this.canvas.clientWidth * devicePixelRatio;
    this.canvas.height = this.canvas.clientHeight * devicePixelRatio;
    this.gl.viewport(0, 0, this.canvas.width, this.canvas.height);
  }

  render(camera: Camera, tileEntries: { coord: TileCoord; data: TileData }[]): void {
    const gl = this.gl;
    gl.clearColor(0.97, 0.97, 0.97, 1.0);
    gl.clear(gl.COLOR_BUFFER_BIT);

    const proj = this.projectionMatrix(camera);

    this.renderEdges(gl, proj, tileEntries, camera);
    this.renderNodes(gl, proj, tileEntries, camera);
  }

  private projectionMatrix(cam: Camera): number[] {
    const scale = Math.pow(2, cam.zoom);
    const tileSize = Math.min(this.canvas.width, this.canvas.height);
    const sx = (tileSize * scale * 2) / this.canvas.width;
    const sy = (tileSize * scale * 2) / this.canvas.height;
    const tx = -cam.x * sx;
    const ty = -cam.y * sy;
    // Column-major 3x3
    return [sx, 0, 0, 0, -sy, 0, tx, ty, 1];
  }

  private renderNodes(
    gl: WebGLRenderingContext,
    proj: number[],
    tileEntries: { coord: TileCoord; data: TileData }[],
    camera: Camera,
  ): void {
    gl.useProgram(this.nodeProgram);

    const uProj = gl.getUniformLocation(this.nodeProgram, 'u_projection');
    gl.uniformMatrix3fv(uProj, false, proj);

    const positions: number[] = [];
    const sizes: number[] = [];
    const colors: number[] = [];
    const opacities: number[] = [];

    for (const { coord, data } of tileEntries) {
      const tileScale = 1 / (1 << coord.z);
      for (const node of data.nodes) {
        const wx = (coord.x + node.x) * tileScale;
        const wy = (coord.y + node.y) * tileScale;
        positions.push(wx, wy);

        const baseSize = 2 + node.importance * 12;
        sizes.push(baseSize * devicePixelRatio);

        const c = CLUSTER_COLORS[node.clusterId % CLUSTER_COLORS.length];
        colors.push(c[0], c[1], c[2]);
        opacities.push(1.0);
      }
    }

    if (positions.length === 0) return;

    this.bindAttribute(gl, this.nodeProgram, 'a_position', new Float32Array(positions), 2);
    this.bindAttribute(gl, this.nodeProgram, 'a_size', new Float32Array(sizes), 1);
    this.bindAttribute(gl, this.nodeProgram, 'a_color', new Float32Array(colors), 3);
    this.bindAttribute(gl, this.nodeProgram, 'a_opacity', new Float32Array(opacities), 1);

    gl.drawArrays(gl.POINTS, 0, positions.length / 2);
  }

  private renderEdges(
    gl: WebGLRenderingContext,
    proj: number[],
    tileEntries: { coord: TileCoord; data: TileData }[],
    camera: Camera,
  ): void {
    gl.useProgram(this.edgeProgram);

    const uProj = gl.getUniformLocation(this.edgeProgram, 'u_projection');
    gl.uniformMatrix3fv(uProj, false, proj);

    const positions: number[] = [];
    const opacities: number[] = [];

    for (const { coord, data } of tileEntries) {
      const tileScale = 1 / (1 << coord.z);
      // Build node position lookup for this tile
      const nodePos = new Map<number, [number, number]>();
      for (const node of data.nodes) {
        nodePos.set(node.id, [
          (coord.x + node.x) * tileScale,
          (coord.y + node.y) * tileScale,
        ]);
      }

      for (const edge of data.edges) {
        const from = nodePos.get(edge.fromId);
        if (!from) continue;
        const toX = (coord.x + edge.toX) * tileScale;
        const toY = (coord.y + edge.toY) * tileScale;

        positions.push(from[0], from[1], toX, toY);
        opacities.push(edge.weight, edge.weight);
      }
    }

    if (positions.length === 0) return;

    this.bindAttribute(gl, this.edgeProgram, 'a_position', new Float32Array(positions), 2);
    this.bindAttribute(gl, this.edgeProgram, 'a_opacity', new Float32Array(opacities), 1);

    gl.drawArrays(gl.LINES, 0, positions.length / 2);
  }

  private bindAttribute(
    gl: WebGLRenderingContext,
    program: WebGLProgram,
    name: string,
    data: Float32Array,
    size: number,
  ): void {
    const loc = gl.getAttribLocation(program, name);
    if (loc < 0) return;
    const buf = gl.createBuffer()!;
    gl.bindBuffer(gl.ARRAY_BUFFER, buf);
    gl.bufferData(gl.ARRAY_BUFFER, data, gl.DYNAMIC_DRAW);
    gl.enableVertexAttribArray(loc);
    gl.vertexAttribPointer(loc, size, gl.FLOAT, false, 0, 0);
  }
}

function createProgram(gl: WebGLRenderingContext, vSrc: string, fSrc: string): WebGLProgram {
  const vs = compileShader(gl, gl.VERTEX_SHADER, vSrc);
  const fs = compileShader(gl, gl.FRAGMENT_SHADER, fSrc);
  const program = gl.createProgram()!;
  gl.attachShader(program, vs);
  gl.attachShader(program, fs);
  gl.linkProgram(program);
  if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
    throw new Error('Program link failed: ' + gl.getProgramInfoLog(program));
  }
  return program;
}

function compileShader(gl: WebGLRenderingContext, type: number, src: string): WebGLShader {
  const shader = gl.createShader(type)!;
  gl.shaderSource(shader, src);
  gl.compileShader(shader);
  if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
    throw new Error('Shader compile failed: ' + gl.getShaderInfoLog(shader));
  }
  return shader;
}
```

**Step 3: Verify TypeScript compiles**

Run: `cd /Users/creimer/code/wikimap/web && npx tsc --noEmit`
Expected: No errors

**Step 4: Commit**

```bash
git add web/src/renderer/
git commit -m "feat: add WebGL renderer with node and edge shaders"
```

---

## Phase 10: Frontend Interaction

### Task 18: Pan and zoom handlers

**Files:**
- Create: `web/src/interaction/controls.ts`
- Create: `web/src/interaction/controls.test.ts`

**Step 1: Write test**

`web/src/interaction/controls.test.ts`:
```typescript
import { describe, it, expect, vi } from 'vitest';
import { PanHandler, ZoomHandler } from './controls';
import { Camera } from '../engine/camera';

describe('PanHandler', () => {
  it('updates camera on drag', () => {
    const cam = new Camera();
    const startX = cam.x;
    const handler = new PanHandler(cam);
    handler.onPointerDown(100, 100);
    handler.onPointerMove(150, 100, 800, 600); // drag right by 50px
    expect(cam.x).not.toBe(startX);
  });

  it('does nothing without drag start', () => {
    const cam = new Camera();
    const startX = cam.x;
    const handler = new PanHandler(cam);
    handler.onPointerMove(150, 100, 800, 600);
    expect(cam.x).toBe(startX);
  });
});

describe('ZoomHandler', () => {
  it('zooms in on positive wheel delta', () => {
    const cam = new Camera();
    const handler = new ZoomHandler(cam);
    handler.onWheel(-100, 400, 300, 800, 600); // scroll up = zoom in
    expect(cam.zoom).toBeGreaterThan(0);
  });

  it('zooms out on negative wheel delta', () => {
    const cam = new Camera();
    cam.zoom = 5;
    const handler = new ZoomHandler(cam);
    handler.onWheel(100, 400, 300, 800, 600); // scroll down = zoom out
    expect(cam.zoom).toBeLessThan(5);
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/creimer/code/wikimap/web && npx vitest run src/interaction/controls.test.ts`
Expected: FAIL

**Step 3: Implement controls**

`web/src/interaction/controls.ts`:
```typescript
import { Camera } from '../engine/camera';

export class PanHandler {
  private dragging = false;
  private lastX = 0;
  private lastY = 0;

  constructor(private camera: Camera) {}

  onPointerDown(x: number, y: number): void {
    this.dragging = true;
    this.lastX = x;
    this.lastY = y;
  }

  onPointerMove(x: number, y: number, screenW: number, screenH: number): void {
    if (!this.dragging) return;
    const dx = x - this.lastX;
    const dy = y - this.lastY;
    this.camera.pan(dx, dy, screenW, screenH);
    this.lastX = x;
    this.lastY = y;
  }

  onPointerUp(): void {
    this.dragging = false;
  }
}

export class ZoomHandler {
  constructor(private camera: Camera) {}

  onWheel(deltaY: number, screenX: number, screenY: number, screenW: number, screenH: number): void {
    const zoomDelta = deltaY > 0 ? -0.5 : 0.5;
    this.camera.zoomBy(zoomDelta, screenX, screenY, screenW, screenH);
  }
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/creimer/code/wikimap/web && npx vitest run src/interaction/controls.test.ts`
Expected: PASS

**Step 5: Commit**

```bash
git add web/src/interaction/
git commit -m "feat: add pan and zoom interaction handlers"
```

---

## Phase 11: Integration

### Task 19: Wire up main.ts

**Files:**
- Modify: `web/src/main.ts`

**Step 1: Implement main app loop**

`web/src/main.ts`:
```typescript
import { Camera, visibleTiles, type TileCoord } from './engine/camera';
import { TileLoader, createFetcher } from './engine/tile-loader';
import { MapRenderer } from './renderer/webgl';
import { PanHandler, ZoomHandler } from './interaction/controls';
import type { TileData } from './types/tile';

const canvas = document.getElementById('map') as HTMLCanvasElement;

const camera = new Camera();
const loader = new TileLoader(createFetcher());
const renderer = new MapRenderer(canvas);
const panHandler = new PanHandler(camera);
const zoomHandler = new ZoomHandler(camera);

// Resize
function resize() {
  canvas.style.width = window.innerWidth + 'px';
  canvas.style.height = window.innerHeight + 'px';
  renderer.resize();
  requestRender();
}
window.addEventListener('resize', resize);
resize();

// Interaction
canvas.addEventListener('pointerdown', e => {
  panHandler.onPointerDown(e.clientX, e.clientY);
  canvas.setPointerCapture(e.pointerId);
});

canvas.addEventListener('pointermove', e => {
  panHandler.onPointerMove(e.clientX, e.clientY, canvas.clientWidth, canvas.clientHeight);
  requestRender();
});

canvas.addEventListener('pointerup', () => {
  panHandler.onPointerUp();
});

canvas.addEventListener('wheel', e => {
  e.preventDefault();
  zoomHandler.onWheel(e.deltaY, e.clientX, e.clientY, canvas.clientWidth, canvas.clientHeight);
  requestRender();
}, { passive: false });

// Render loop
let renderPending = false;
const loadedTiles = new Map<string, { coord: TileCoord; data: TileData }>();

function requestRender() {
  if (renderPending) return;
  renderPending = true;
  requestAnimationFrame(renderFrame);
}

function renderFrame() {
  renderPending = false;

  const tiles = visibleTiles(camera, canvas.clientWidth, canvas.clientHeight);

  // Load any missing tiles
  for (const coord of tiles) {
    const key = `${coord.z}/${coord.x}/${coord.y}`;
    if (!loadedTiles.has(key)) {
      loader.load(coord.z, coord.x, coord.y).then(data => {
        loadedTiles.set(key, { coord, data });
        requestRender();
      });
    }
  }

  // Collect loaded tile data for visible tiles
  const entries: { coord: TileCoord; data: TileData }[] = [];
  for (const coord of tiles) {
    const key = `${coord.z}/${coord.x}/${coord.y}`;
    const entry = loadedTiles.get(key);
    if (entry) entries.push(entry);
  }

  renderer.render(camera, entries);
}

requestRender();
```

**Step 2: Verify build**

Run: `cd /Users/creimer/code/wikimap/web && npx tsc --noEmit`
Expected: No errors

**Step 3: Commit**

```bash
git add web/src/main.ts
git commit -m "feat: wire up main app loop with tile loading and rendering"
```

---

### Task 20: End-to-end test with small dataset

**Files:**
- Create: `testdata/small-wiki.xml` (copy of parser test fixture)

**Step 1: Create a small test dataset with more articles**

Create `testdata/small-wiki.xml` with ~20 articles covering a few topic clusters (animals, countries, sciences). This gives enough nodes to verify the full pipeline produces tiles with visible structure.

**Step 2: Run the pipeline**

Run:
```bash
cd /Users/creimer/code/wikimap
go run ./cmd/pipeline -dump testdata/small-wiki.xml -out data/tiles -max-zoom 4
```
Expected: Output showing parsed articles, graph stats, communities, layout, tile generation.

**Step 3: Start the server**

Run:
```bash
cd /Users/creimer/code/wikimap
go run ./cmd/server -tiles data/tiles -web "" -port 8080 &
```

**Step 4: Start the frontend dev server**

Run: `cd /Users/creimer/code/wikimap/web && npm run dev`

**Step 5: Verify in browser**

Open http://localhost:5173. You should see colored dots (nodes) on a light background. Drag to pan, scroll to zoom. The nodes should be clustered by community.

**Step 6: Commit test data**

```bash
git add testdata/
git commit -m "test: add small wiki dataset for end-to-end testing"
```

---

## Next Steps (Future Plans)

These are out of scope for this plan but will be needed:

1. **Protobuf serialization** -- replace JSON tile format with protobuf for smaller payloads
2. **Search** -- add search endpoint and UI search bar with fly-to animation
3. **Labels** -- render article title text in WebGL (SDF text or canvas overlay)
4. **Article detail panel** -- click a node to see article summary
5. **Wikipedia clickstream data** -- import click data for edge weights
6. **Pageview data** -- import pageview dumps for node importance
7. **Large-scale layout** -- Barnes-Hut optimization for force-directed layout at scale
8. **Smooth zoom cross-fade** -- interpolate node opacity across fractional zoom levels
9. **Cluster region rendering** -- filled polygons/voronoi for topic areas
10. **Full English Wikipedia** -- process the complete dump (~22GB)
