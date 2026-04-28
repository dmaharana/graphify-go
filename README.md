# Graphify-Go

A Golang-based structural codebase extractor inspired by [Graphify](https://graphify.net/). It uses Tree-Sitter to generate a knowledge graph of code architecture.

## Features
- **Fast AST Parsing**: Uses Tree-Sitter for high-performance, multi-language support.
- **Multi-language**: Currently supports Go and Javascript (extensible).
- **Graph Generation**: Outputs a structured `graph.json` mapping files, functions, structs, and interfaces.
- **Gemini CLI Ready**: Includes a `SKILL.md` for seamless integration as an AI assistant skill.

## Installation
Ensure you have Go installed, then:
```bash
go mod tidy
go build -o graphify-go main.go
```

## Usage
Scan your codebase:
```bash
./graphify-go scan .
```
This generates a `graph.json` file.

Query the architecture:
```bash
./graphify-go query "how does the scanner work?" --graph graph.json
```
The query command extracts a relevant subgraph and formats it in Markdown, perfect for providing context to an AI assistant.

## Extending
To add support for more languages (e.g., Python):
1. Add the grammar: `go get github.com/smacker/go-tree-sitter/python`
2. Update `pkg/parser/parser.go` to handle the `.py` extension.
3. Update the `walk` function if the language uses different node types for symbols.
