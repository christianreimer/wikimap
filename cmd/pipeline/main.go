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

	log.Println("Building graph...")
	g := parser.BuildGraph(articles)
	log.Printf("Graph: %d nodes, %d edges", g.NodeCount(), g.EdgeCount())

	log.Println("Computing PageRank...")
	importance := graph.PageRank(g, 0.85, 50)

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

	log.Println("Computing layout...")
	positions := layout.ForceDirected(g, 500)

	log.Println("Generating tiles...")
	allTiles := tiles.GenerateTiles(g, positions, importance, communities, 0, *maxZoom)
	log.Printf("Generated %d tiles", len(allTiles))

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
