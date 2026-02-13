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
