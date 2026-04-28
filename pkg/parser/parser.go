package parser

import (
	"context"
	"fmt"
	"graphify-go/pkg/graph"
	"os"
	"path/filepath"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/javascript"
)

func ParseFile(path string) ([]graph.Node, []graph.Edge, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	ext := filepath.Ext(path)
	var lang *sitter.Language
	switch ext {
	case ".go":
		lang = golang.GetLanguage()
	case ".js":
		lang = javascript.GetLanguage()
	default:
		return nil, nil, fmt.Errorf("unsupported language: %s", ext)
	}

	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, nil, err
	}

	nodes := []graph.Node{}
	edges := []graph.Edge{}

	root := tree.RootNode()
	
	// Simple traversal to find functions and structs in Go
	walk(root, content, path, &nodes, &edges)

	return nodes, edges, nil
}

func walk(n *sitter.Node, content []byte, path string, nodes *[]graph.Node, edges *[]graph.Edge) {
	nodeType := n.Type()
	
	switch nodeType {
	case "function_declaration", "method_declaration", "method_definition":
		nameNode := n.ChildByFieldName("name")
		if nameNode != nil {
			name := string(content[nameNode.StartByte():nameNode.EndByte()])
			id := fmt.Sprintf("%s:%s", path, name)
			*nodes = append(*nodes, graph.Node{
				ID:       id,
				Type:     graph.FunctionNode,
				Name:     name,
				FilePath: path,
				Line:     int(n.StartPoint().Row) + 1,
			})
		}
	case "type_spec", "class_declaration":
		nameNode := n.ChildByFieldName("name")
		if nameNode != nil {
			name := string(content[nameNode.StartByte():nameNode.EndByte()])
			id := fmt.Sprintf("%s:%s", path, name)
			
			gType := graph.StructNode
			if nodeType == "type_spec" {
				typeNode := n.ChildByFieldName("type")
				if typeNode != nil && typeNode.Type() == "interface_type" {
					gType = graph.InterfaceNode
				}
			}

			*nodes = append(*nodes, graph.Node{
				ID:       id,
				Type:     gType,
				Name:     name,
				FilePath: path,
				Line:     int(n.StartPoint().Row) + 1,
			})
		}
	}

	for i := 0; i < int(n.ChildCount()); i++ {
		walk(n.Child(i), content, path, nodes, edges)
	}
}
