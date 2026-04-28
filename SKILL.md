# Graphify-Go Skill

Graphify-Go is a high-performance codebase structural extractor implemented in Go. It uses Tree-Sitter to build a knowledge graph of your code's architecture, identifying functions, structs, interfaces, and their relationships across multiple languages.

## Purpose
Use this skill when you need a deep architectural understanding of a codebase. It provides a structured `graph.json` that maps out how different parts of the system connect, which is far more efficient than brute-force file reading for large projects.

## Commands
The primary commands are:
```bash
graphify-go scan [directory]
graphify-go query "[question]" --graph graph.json
```
`scan` will:
1. Walk the directory (respecting common ignores).
2. Parse supported files (Go, Javascript) using Tree-Sitter.
3. Generate a `graph.json` mapping the architecture.

`query` will:
1. Extract a relevant subgraph based on the question provided.
2. Format the output in Markdown, optimized for AI consumption.
3. Include source locations and relationships to provide context.

## Integration Strategy
When using this skill, the agent should:
1. Run `graphify-go scan .` to generate the latest architectural graph.
2. Use `graphify-go query "[question]"` to pull a targeted context block when investigating specific areas (e.g., "how is auth handled?").
3. Use the graph data to navigate the codebase more effectively when answering complex questions.

## Supported Languages
- Go (.go)
- Javascript (.js)
- *Extendable to Python, TypeScript, Java, C++, etc.*
