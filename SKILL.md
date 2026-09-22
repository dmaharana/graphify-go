# Graphify-Go Skill

Graphify-Go is a high-performance codebase structural knowledge graph engine implemented in Go. It uses Tree-Sitter AST parsing and in-memory graph algorithms to extract code architecture, cross-file relationships, and communities.

## Purpose
Use this skill when you need deep architectural understanding of a codebase. It builds a structured graph in `graphify-out/` that maps how components connect, identify hubs, compute blast radiuses, and trace dependency paths without brute-force file reading.

## Commands
```bash
# Scan codebase (produces graphify-out/graph.json, GRAPH_REPORT.md, graph.html)
graphify-go scan .
graphify-go scan ./cmd ./pkg --update

# Query architecture for contextual subgraph
graphify-go query "[question]"

# Explain a specific symbol (degree, community, callers, dependencies)
graphify-go explain "[symbol]"

# Find shortest path between two symbols
graphify-go path "[source]" "[target]"

# Blast-radius analysis (who breaks if this changes?)
graphify-go affected "[file|symbol]"

# Start stdio MCP server for agent tool calls
graphify-go serve
```

## Integration Strategy for Agents
1. Run `graphify-go scan .` on first interaction to build `graphify-out/`.
2. Consult `graphify-out/GRAPH_REPORT.md` for high-level subsystems, God nodes, and cross-boundary links.
3. Use `graphify-go query "[question]"` or MCP `query_graph` to retrieve focused context subgraphs instead of grepping.
4. Use `graphify-go affected "[symbol]"` before refactoring to verify upstream impact.
