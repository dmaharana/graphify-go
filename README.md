# Graphify-Go

A high-performance, local-first code knowledge graph engine implemented in Go, inspired by [Graphify](https://graphify.com/). It uses Tree-Sitter AST parsing and pure Go graph algorithms to build an architectural knowledge graph you can query, explain, trace, and serve via MCP.

## Features

- **Deep AST & Relational Extraction**: Captures calls (`calls`), method receivers (`method`), struct embeddings (`embeds`), interface signatures, and imports (`imports`), with builtins filtering to avoid phantom hubs.
- **Cross-File Symbol Resolution**: Automatically resolves package-qualified calls (`pkg.Func`) and method calls across files with confidence tags (`EXTRACTED` vs `INFERRED`).
- **Graph Analysis & Queries**:
  - `explain <symbol>`: Inspects degree centrality, community assignment, callers, and dependencies.
  - `path <source> <target>`: Finds the shortest architectural call/dependency path between two components.
  - `affected <file|symbol>`: Calculates the transitive upstream blast radius.
  - `query <question>`: Extracts a focused Markdown subgraph for AI assistants.
- **Community Detection & Reports**: Pure Go Louvain modularity clustering partitions code into subsystems and generates `GRAPH_REPORT.md` (God nodes, subsystems, surprising cross-boundary links).
- **Interactive Visualizer**: Generates a self-contained, offline interactive `graph.html` visualization with deterministic layout seeding, pruning for large graphs, focus mode, path tracing, semantic shapes/edges, mini-map, and dark/light themes.
- **Multi-Root & Incremental Scanning**: Scans multiple source directories (`scan ./cmd ./pkg`), honors `.gitignore` / `.graphifyignore`, and caches file hashes in `manifest.json` for rapid `--update` scans.
- **Stdio MCP Server**: Built-in Model Context Protocol server (`graphify-go serve`) for direct tool-call integration with Claude Code, Gemini CLI, Cursor, and Codex.

## Installation

```bash
go mod tidy
go build -o graphify-go main.go
```

## Usage

### 1. Scan Codebase
Scan one or more directories:
```bash
./graphify-go scan .
# or scan multiple directories:
./graphify-go scan ./cmd ./pkg
# with visualization customization:
./graphify-go scan . --viz-max-nodes 1500 --viz-layout hierarchical --viz-theme dark
```
Artifacts are generated in `graphify-out/`:
```
graphify-out/
├── graph.json        # Complete graph representation
├── GRAPH_REPORT.md   # Architectural audit report (God nodes, communities, cross-links)
├── graph.html        # Interactive offline graph visualizer
└── manifest.json     # Incremental scan state cache
```

#### Visualizer Features & Flags
The generated `graph.html` is completely self-contained and offline (zero external CDN or network requests):
- **Deterministic Layout Seeding**: Fermat sunflower spiral community-aware layout prevents node stacking and Barnes-Hut quadtree depth overflows.
- **Intelligent Pruning for Scalability**: Graphs exceeding `--viz-max-nodes` (default: 1000) automatically retain highest-degree hubs plus at least one representative from every community, displaying a clear status banner.
- **Physics & Layout Modes**:
  - `Force-directed` (ForceAtlas2) for organic cluster exploration.
  - `Hierarchical` (Left-to-Right) for call flow and dependency direction analysis.
  - Freeze/unfreeze physics (`P`) and Stabilize & Freeze controls.
- **Rich Node & Edge Semantics**:
  - Semantic shapes: Functions (● dot), Methods (◆ diamond), Structs (■ square), Interfaces (▲ triangle), Packages (★ star), Files (⬡ hexagon).
  - Edge confidence styling: `EXTRACTED` (solid), `INFERRED` (dashed), `AMBIGUOUS` (dotted amber).
  - Architectural "God Nodes" highlighted with distinct gold border and shadow glow.
  - Degree breakdown: In-degree and Out-degree caller/callee metrics in inspector.
- **Interactive Navigation & Focus Mode**:
  - Clicking any node dims unrelated elements and highlights the 1-hop neighborhood.
  - Clicking connected neighbors centers and inspects them.
  - Built-in Shortest Path Finder (`⇄ Path` or URL query `?from=X&to=Y`) traces call routes and dims unrelated graph elements.
  - Advanced Search: Filter by node type, community, or regex (`.*`).
  - Interactive Mini-map: Bottom-right overview pane with click-to-pan.
  - Source links: "Open in VS Code" (`vscode://file/...`) and "Copy Location" (`file:line`).
- **Visual Polish**:
  - Dark / Light theme toggle (`🌓`) with high-contrast palettes and `localStorage` persistence.
  - Export view as PNG (`📷 PNG`).
  - Keyboard shortcuts: `/` (Search), `F` (Fit view), `P` (Toggle physics), `Esc` (Clear focus/path), `?` / `H` (Help).

CLI Flags:
- `--viz-max-nodes <int>`: Max nodes to include in visualizer before pruning (default: `1000`).
- `--viz-layout <force|hierarchical>`: Default layout mode (default: `force`).
- `--viz-theme <dark|light|auto>`: Default UI theme (default: `dark`).

Incremental update:
```bash
./graphify-go scan . --update
```

### 2. Architecture Queries
Query architecture for context:
```bash
./graphify-go query "how does the scanner handle ignores?"
```

Explain a symbol:
```bash
./graphify-go explain "Scanner"
```

Find the shortest path between two symbols:
```bash
./graphify-go path "ScanWithManifest" "NewResolver"
```

Determine blast radius:
```bash
./graphify-go affected "pkg/scanner/scanner.go:NewScanner"
```

### 3. Agent Integration (MCP)
Start the stdio MCP server for AI coding assistants:
```bash
./graphify-go serve
```
Available tools:
- `query_graph`: Plain-language architectural queries returning Markdown subgraphs.
- `get_node`: Inspect node properties and source lines.
- `get_neighbors`: List incoming and outgoing edges.
- `shortest_path`: Compute shortest path between two symbols.
- `get_impact`: Transitive blast-radius analysis.
