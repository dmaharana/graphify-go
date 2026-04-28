# Graphify-Go Skill

Graphify-Go is a high-performance codebase structural extractor implemented in Go. It uses Tree-Sitter to build a knowledge graph of your code's architecture, identifying functions, structs, interfaces, and their relationships across multiple languages.

## Purpose
Use this skill when you need a deep architectural understanding of a codebase. It provides a structured `graph.json` that maps out how different parts of the system connect, which is far more efficient than brute-force file reading for large projects.

## Commands
The primary command is:
```bash
graphify-go scan [directory]
```
This will:
1. Walk the directory (respecting common ignores).
2. Parse supported files (Go, Javascript) using Tree-Sitter.
3. Generate a `graph.json` mapping the architecture.

## Integration Strategy
When using this skill, the agent should:
1. Run `./graphify-go scan .` to generate the latest architectural graph.
2. Read the resulting `graph.json` to identify key entry points and structural patterns.
3. Use the graph data to navigate the codebase more effectively when answering complex questions.

## Supported Languages
- Go (.go)
- Javascript (.js)
- *Extendable to Python, TypeScript, Java, C++, etc.*
