# WikiMap Design Document

**Date:** 2026-02-12
**Status:** Approved

## Overview

WikiMap is a Google Maps-style interactive visualization of Wikipedia's link graph. Users explore Wikipedia as a navigable map where articles are cities, links are roads, and topic clusters form states, countries, and continents. The experience uses semantic zooming with progressive disclosure -- zooming in reveals more detail, just like Google Maps.

Based on the smooth-zoom tile approach described in [Prototyping a Smoother Map](https://medium.com/google-design/google-maps-cb0326d165f5) by Antin Harasymiv, adapted from raster image tiles to vector tiles rendered with WebGL.

## Architecture

```
Wikipedia Dump (XML)
        |
        v
  Data Pipeline (Go)
  - Parse dump
  - Build graph
  - Community detection
  - Spatial layout
  - Tile generation
        |
        v
  Vector Tiles (disk / object storage)
        |
        v
  Tile Server (Go HTTP)
  - GET /tiles/{z}/{x}/{y}.pbf
  - GET /search?q={query}
  - GET /article/{id}
        |
        v
  Frontend (Vanilla TypeScript + WebGL)
  - Tile loader + viewport manager
  - WebGL renderer (clusters, edges, nodes, labels)
  - Interaction (pan, zoom, hover, click, search)
```

## Tech Stack

- **Backend / Pipeline:** Go
- **Frontend:** Vanilla TypeScript + WebGL
- **Data Source:** English Wikipedia dump (~6.8M articles)
- **Tile Format:** Protocol Buffers (or flat binary)
- **Rendering:** WebGL with instanced rendering

## Data Pipeline

### Stage 1: Parse Wikipedia Dump

- Download `enwiki-*-pages-articles.xml.bz2` (~22GB compressed)
- Stream-parse XML to extract: article title, article ID, page views (from separate pageview dump), outgoing wikilinks
- Output: compact adjacency list + metadata file

### Stage 2: Graph Construction & Metrics

- Build directed graph in memory (~16-32GB RAM for 6.8M nodes)
- Compute node importance via PageRank or combined page views + in-degree
- Compute edge weights from Wikipedia clickstream dataset

### Stage 3: Community Detection & Hierarchy

- Run hierarchical community detection (Louvain/Leiden) to create cluster hierarchy
- Produces 3-5 levels mapping to zoom ranges:
  - Level 0: ~20-50 top-level clusters (Science, History, Geography, etc.)
  - Level 1: ~200-500 sub-clusters
  - Level 2: ~2000-5000 fine-grained clusters
  - Level 3+: individual articles

### Stage 4: Spatial Layout

- **Inter-cluster:** Position cluster centroids using force-directed placement (clusters repel, connected clusters attract)
- **Intra-cluster:** Position nodes within clusters using force-directed layout or Barnes-Hut approximation
- Hierarchical approach avoids laying out all 6.8M nodes simultaneously
- Output: (x, y) coordinates for every node, normalized to world coordinate space

### Stage 5: Tile Generation

- Divide coordinate space into quadtree-based tile grid (z/x/y scheme)
- For each zoom level, include nodes and edges above the importance threshold for that level
- Serialize tiles as compact binary (Protocol Buffers) containing:
  - Node positions, labels, importance scores
  - Edges with weights
  - Cluster boundaries
- Write tiles to disk

## Tile Server

Lightweight Go HTTP server, essentially a static file server with search.

### Endpoints

- `GET /tiles/{z}/{x}/{y}.pbf` -- vector tile at given coordinates
- `GET /search?q={query}` -- article ID + coordinates for navigation
- `GET /article/{id}` -- article metadata (title, summary, links, coordinates)

### Implementation

- Go net/http or chi router
- Tiles from disk or embedded KV store (bbolt/pebble)
- Search via pre-built index (bleve or prefix trie)
- Aggressive caching headers (tiles are immutable)
- CORS for local dev; static file serving for frontend in production

## Frontend

### Map Engine

- Custom tile-loading engine managing a viewport into world coordinate space
- Camera state: (x, y, zoom)
- Computes visible tiles, loads from server, unloads off-screen tiles
- Smooth zoom cross-fading: tiles from adjacent zoom levels rendered simultaneously with interpolated opacity during transitions

### WebGL Rendering Layers (bottom to top)

1. **Cluster regions** -- filled polygons/voronoi cells for topic boundaries, colored by category
2. **Edges** -- lines between connected nodes, importance-thresholded per zoom. Aggregated "highways" at low zoom.
3. **Nodes** -- circles/points, size encodes importance (page views). Progressive disclosure by zoom.
4. **Labels** -- article titles, progressive disclosure matching nodes

### Interaction

- **Pan:** Click-drag / touch-drag
- **Zoom:** Scroll wheel / pinch-to-zoom with smooth animation
- **Hover:** Highlight node, tooltip with title and view count
- **Click:** Select node, show article summary panel, highlight connections
- **Search:** Text input, camera flies to matching article

### Performance

- WebGL instanced rendering for nodes (one draw call per layer)
- Edge bundling/simplification at low zoom
- Render to offscreen texture at rest, re-render only on camera movement
- LOD: only process data above importance threshold per zoom level

### UI Chrome

- Search bar (top)
- Zoom controls (+/-)
- Article detail panel (slides in from right)
- Legend (cluster/topic colors)

## Zoom Levels & Progressive Disclosure

| Zoom | Map Analogy | Visible Content | Node Threshold |
|------|-------------|----------------|----------------|
| 0-2 | World | ~20-50 top-level topic clusters as colored regions | Top ~50 articles |
| 3-5 | Continent | Sub-clusters, ~500-2000 labeled articles | Top ~2000 |
| 6-9 | Country/State | Individual clusters, ~10K-50K articles, major link highways | Top ~50K |
| 10-13 | City | All articles in viewport, individual links, article labels dominate | Top ~500K |
| 14-18 | Street | Full detail -- every article, every link | All 6.8M |

### Transition Behavior

- Nodes fade in (opacity 0 to 1) over ~0.5 zoom levels as importance threshold is crossed
- Cluster boundaries fade inversely with zoom
- Labels appear with slight delay to avoid flicker
- Edges transition from aggregated highways to individual links

### Visual Hierarchy

- **Node size:** logarithmic scale of page views, clamped per zoom level
- **Node color:** top-level cluster membership
- **Edge opacity:** edge weight (click-through frequency)
- **Label font size:** scales with importance, clamped per zoom level

## Project Structure

```
wikimap/
  cmd/
    pipeline/       # Data processing CLI
    server/         # Tile server
  internal/
    parser/         # Wikipedia dump parser
    graph/          # Graph construction, PageRank, metrics
    cluster/        # Community detection, hierarchy
    layout/         # Force-directed layout, coordinate assignment
    tiles/          # Tile generation, serialization
    search/         # Search index building
    server/         # HTTP server, tile serving
  proto/
    tile.proto      # Tile protobuf definition
  web/
    src/
      main.ts       # Entry point
      engine/       # Map engine (viewport, tile loading, camera)
      renderer/     # WebGL rendering (nodes, edges, labels, clusters)
      interaction/  # Pan, zoom, hover, click handlers
      ui/           # Search bar, detail panel, controls
      types/        # TypeScript type definitions
    index.html
    tsconfig.json
  docs/
    plans/
  go.mod
  go.sum
```
