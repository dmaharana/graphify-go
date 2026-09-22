# Pure Local-First Deterministic Code Intelligence

We decided to keep `graphify-go` strictly offline and deterministic, using AST parsing (Tree-Sitter) and in-memory graph algorithms without external LLM dependencies for code extraction and graph analysis. This ensures sub-second scan times, zero API costs, zero data residency concerns, and a single portable Go binary, leaving higher-level LLM reasoning to the AI agent consuming the graph.
