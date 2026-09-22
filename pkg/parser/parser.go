package parser

import (
	"context"
	"fmt"
	"graphify-go/pkg/graph"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/javascript"
)

type RawCall struct {
	CallerID   string `json:"caller_id"`
	CalleeName string `json:"callee_name"`
	Receiver   string `json:"receiver,omitempty"`
	ImportPath string `json:"import_path,omitempty"`
	IsMember   bool   `json:"is_member"`
	Line       int    `json:"line"`
	SourceFile string `json:"source_file"`
}

type ParseResult struct {
	PackageName string
	Nodes       []graph.Node
	Edges       []graph.Edge
	RawCalls    []RawCall
	Imports     map[string]string
}

func ParseFile(path string) (*ParseResult, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(path)
	var lang *sitter.Language
	switch ext {
	case ".go":
		lang = golang.GetLanguage()
	case ".js", ".jsx", ".ts", ".tsx":
		lang = javascript.GetLanguage()
	default:
		return nil, fmt.Errorf("unsupported language: %s", ext)
	}

	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}

	res := &ParseResult{
		Imports: make(map[string]string),
	}

	root := tree.RootNode()

	switch ext {
	case ".go":
		parseGo(root, content, path, res)
	case ".js", ".jsx", ".ts", ".tsx":
		parseJS(root, content, path, res)
	}

	return res, nil
}

func parseGo(root *sitter.Node, content []byte, path string, res *ParseResult) {
	strPath := path

	fileNID := path
	res.Nodes = append(res.Nodes, graph.Node{
		ID:       fileNID,
		Type:     graph.FileNode,
		Name:     filepath.Base(path),
		FilePath: path,
		Line:     1,
	})

	// Pass 1: find package name and imports
	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		t := child.Type()

		if t == "package_clause" {
			for j := 0; j < int(child.ChildCount()); j++ {
				c := child.Child(j)
				if c.Type() == "package_identifier" {
					res.PackageName = string(content[c.StartByte():c.EndByte()])
				}
			}
		} else if t == "import_declaration" {
			extractGoImports(child, content, fileNID, path, res)
		}
	}

	if res.PackageName == "" {
		res.PackageName = filepath.Base(filepath.Dir(path))
	}

	type funcBody struct {
		id   string
		node *sitter.Node
	}
	var bodies []funcBody
	localFuncs := make(map[string]string)

	// Pass 2: declarations
	var walkDecls func(n *sitter.Node)
	walkDecls = func(n *sitter.Node) {
		if n == nil {
			return
		}
		t := n.Type()

		switch t {
		case "function_declaration":
			nameNode := n.ChildByFieldName("name")
			if nameNode == nil {
				for i := 0; i < int(n.ChildCount()); i++ {
					if n.Child(i).Type() == "identifier" {
						nameNode = n.Child(i)
						break
					}
				}
			}
			if nameNode != nil {
				name := string(content[nameNode.StartByte():nameNode.EndByte()])
				line := int(n.StartPoint().Row) + 1
				id := fmt.Sprintf("%s:%s", path, name)
				res.Nodes = append(res.Nodes, graph.Node{
					ID:       id,
					Type:     graph.FunctionNode,
					Name:     name,
					FilePath: path,
					Line:     line,
				})
				res.Edges = append(res.Edges, graph.Edge{
					From:       fileNID,
					To:         id,
					Type:       graph.EdgeContains,
					Confidence: graph.ConfidenceExtracted,
					Line:       line,
					SourceFile: strPath,
				})
				localFuncs[name] = id

				body := n.ChildByFieldName("body")
				if body == nil {
					for i := 0; i < int(n.ChildCount()); i++ {
						if n.Child(i).Type() == "block" {
							body = n.Child(i)
							break
						}
					}
				}
				if body != nil {
					bodies = append(bodies, funcBody{id: id, node: body})
				}
			}

		case "method_declaration":
			nameNode := n.ChildByFieldName("name")
			if nameNode == nil {
				for i := 0; i < int(n.ChildCount()); i++ {
					if n.Child(i).Type() == "field_identifier" {
						nameNode = n.Child(i)
						break
					}
				}
			}
			receiverNode := n.ChildByFieldName("receiver")
			if receiverNode == nil {
				for i := 0; i < int(n.ChildCount()); i++ {
					if n.Child(i).Type() == "parameter_list" {
						receiverNode = n.Child(i)
						break
					}
				}
			}
			if nameNode != nil && receiverNode != nil {
				name := string(content[nameNode.StartByte():nameNode.EndByte()])
				receiverType := extractReceiverType(receiverNode, content)
				line := int(n.StartPoint().Row) + 1

				structID := fmt.Sprintf("%s:%s", path, receiverType)
				methodID := fmt.Sprintf("%s:%s.%s", path, receiverType, name)

				res.Nodes = append(res.Nodes, graph.Node{
					ID:       methodID,
					Type:     graph.MethodNode,
					Name:     name,
					FilePath: path,
					Line:     line,
				})
				res.Edges = append(res.Edges, graph.Edge{
					From:       structID,
					To:         methodID,
					Type:       graph.EdgeMethod,
					Confidence: graph.ConfidenceExtracted,
					Line:       line,
					SourceFile: strPath,
				})
				localFuncs[name] = methodID

				body := n.ChildByFieldName("body")
				if body == nil {
					for i := 0; i < int(n.ChildCount()); i++ {
						if n.Child(i).Type() == "block" {
							body = n.Child(i)
							break
						}
					}
				}
				if body != nil {
					bodies = append(bodies, funcBody{id: methodID, node: body})
				}
			}

		case "type_declaration":
			for i := 0; i < int(n.ChildCount()); i++ {
				ts := n.Child(i)
				if ts.Type() != "type_spec" {
					continue
				}
				nameNode := ts.ChildByFieldName("name")
				if nameNode == nil {
					for j := 0; j < int(ts.ChildCount()); j++ {
						if ts.Child(j).Type() == "type_identifier" {
							nameNode = ts.Child(j)
							break
						}
					}
				}
				if nameNode == nil {
					continue
				}
				typeName := string(content[nameNode.StartByte():nameNode.EndByte()])
				line := int(ts.StartPoint().Row) + 1
				typeID := fmt.Sprintf("%s:%s", path, typeName)

				typeType := ts.ChildByFieldName("type")
				if typeType == nil {
					for j := 0; j < int(ts.ChildCount()); j++ {
						cType := ts.Child(j).Type()
						if cType == "struct_type" || cType == "interface_type" {
							typeType = ts.Child(j)
							break
						}
					}
				}

				gType := graph.StructNode
				if typeType != nil && typeType.Type() == "interface_type" {
					gType = graph.InterfaceNode
				}

				res.Nodes = append(res.Nodes, graph.Node{
					ID:       typeID,
					Type:     gType,
					Name:     typeName,
					FilePath: path,
					Line:     line,
				})
				res.Edges = append(res.Edges, graph.Edge{
					From:       fileNID,
					To:         typeID,
					Type:       graph.EdgeContains,
					Confidence: graph.ConfidenceExtracted,
					Line:       line,
					SourceFile: strPath,
				})

				if typeType != nil && typeType.Type() == "struct_type" {
					extractStructFields(typeType, content, typeID, path, res)
				} else if typeType != nil && typeType.Type() == "interface_type" {
					extractInterfaceMethods(typeType, content, typeID, path, res)
				}
			}
		}

		for i := 0; i < int(n.ChildCount()); i++ {
			c := n.Child(i)
			if t == "source_file" || (t != "function_declaration" && t != "method_declaration" && t != "type_declaration") {
				walkDecls(c)
			}
		}
	}
	walkDecls(root)

	// Pass 3: calls within function bodies
	for _, b := range bodies {
		walkCalls(b.node, content, b.id, path, res, localFuncs)
	}
}

func extractReceiverType(receiverNode *sitter.Node, content []byte) string {
	for i := 0; i < int(receiverNode.ChildCount()); i++ {
		p := receiverNode.Child(i)
		if p.Type() == "parameter_declaration" {
			tNode := p.ChildByFieldName("type")
			if tNode == nil {
				for j := 0; j < int(p.ChildCount()); j++ {
					c := p.Child(j)
					if c.Type() == "pointer_type" || c.Type() == "type_identifier" {
						tNode = c
						break
					}
				}
			}
			if tNode != nil {
				raw := string(content[tNode.StartByte():tNode.EndByte()])
				raw = strings.TrimPrefix(raw, "*")
				return strings.TrimSpace(raw)
			}
		}
	}
	return ""
}

func extractGoImports(importDecl *sitter.Node, content []byte, fileNID, path string, res *ParseResult) {
	var processSpec func(spec *sitter.Node)
	processSpec = func(spec *sitter.Node) {
		if spec.Type() != "import_spec" {
			return
		}
		pathNode := spec.ChildByFieldName("path")
		if pathNode == nil {
			for i := 0; i < int(spec.ChildCount()); i++ {
				if spec.Child(i).Type() == "interpreted_string_literal" {
					pathNode = spec.Child(i)
					break
				}
			}
		}
		if pathNode == nil {
			return
		}
		importPath := strings.Trim(string(content[pathNode.StartByte():pathNode.EndByte()]), `"`)
		aliasNode := spec.ChildByFieldName("name")
		var alias string
		if aliasNode != nil {
			alias = string(content[aliasNode.StartByte():aliasNode.EndByte()])
		} else {
			parts := strings.Split(importPath, "/")
			alias = parts[len(parts)-1]
		}
		if alias != "_" && alias != "." {
			res.Imports[alias] = importPath
		}

		line := int(spec.StartPoint().Row) + 1
		res.Edges = append(res.Edges, graph.Edge{
			From:       fileNID,
			To:         "pkg:" + importPath,
			Type:       graph.EdgeImports,
			Confidence: graph.ConfidenceExtracted,
			Line:       line,
			SourceFile: path,
		})
	}

	for i := 0; i < int(importDecl.ChildCount()); i++ {
		c := importDecl.Child(i)
		if c.Type() == "import_spec" {
			processSpec(c)
		} else if c.Type() == "import_spec_list" {
			for j := 0; j < int(c.ChildCount()); j++ {
				processSpec(c.Child(j))
			}
		}
	}
}

func extractStructFields(structType *sitter.Node, content []byte, structID, path string, res *ParseResult) {
	for i := 0; i < int(structType.ChildCount()); i++ {
		fieldList := structType.Child(i)
		if fieldList.Type() != "field_declaration_list" {
			continue
		}
		for j := 0; j < int(fieldList.ChildCount()); j++ {
			field := fieldList.Child(j)
			if field.Type() != "field_declaration" {
				continue
			}
			hasName := false
			var typeNode *sitter.Node
			for k := 0; k < int(field.ChildCount()); k++ {
				c := field.Child(k)
				if c.Type() == "field_identifier" {
					hasName = true
				} else if c.Type() == "type_identifier" || c.Type() == "pointer_type" || c.Type() == "qualified_type" {
					typeNode = c
				}
			}
			if typeNode != nil {
				typeName := strings.TrimPrefix(string(content[typeNode.StartByte():typeNode.EndByte()]), "*")
				line := int(field.StartPoint().Row) + 1
				if !hasName {
					// Anonymous field = embedded struct
					res.Edges = append(res.Edges, graph.Edge{
						From:       structID,
						To:         typeName,
						Type:       graph.EdgeEmbeds,
						Confidence: graph.ConfidenceExtracted,
						Line:       line,
						SourceFile: path,
					})
				} else if !IsGoPredeclaredType(typeName) {
					res.Edges = append(res.Edges, graph.Edge{
						From:       structID,
						To:         typeName,
						Type:       graph.EdgeReferences,
						Confidence: graph.ConfidenceExtracted,
						Line:       line,
						SourceFile: path,
					})
				}
			}
		}
	}
}

func extractInterfaceMethods(ifaceType *sitter.Node, content []byte, ifaceID, path string, res *ParseResult) {
	for i := 0; i < int(ifaceType.ChildCount()); i++ {
		elem := ifaceType.Child(i)
		if elem.Type() == "method_elem" {
			nameNode := elem.ChildByFieldName("name")
			if nameNode == nil {
				for j := 0; j < int(elem.ChildCount()); j++ {
					if elem.Child(j).Type() == "field_identifier" {
						nameNode = elem.Child(j)
						break
					}
				}
			}
			if nameNode != nil {
				methodName := string(content[nameNode.StartByte():nameNode.EndByte()])
				line := int(elem.StartPoint().Row) + 1
				methodID := fmt.Sprintf("%s.%s", ifaceID, methodName)

				res.Nodes = append(res.Nodes, graph.Node{
					ID:       methodID,
					Type:     graph.MethodNode,
					Name:     methodName,
					FilePath: path,
					Line:     line,
				})
				res.Edges = append(res.Edges, graph.Edge{
					From:       ifaceID,
					To:         methodID,
					Type:       graph.EdgeMethod,
					Confidence: graph.ConfidenceExtracted,
					Line:       line,
					SourceFile: path,
				})
			}
		}
	}
}

func walkCalls(node *sitter.Node, content []byte, callerID, path string, res *ParseResult, localFuncs map[string]string) {
	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if n.Type() == "call_expression" {
			fnNode := n.ChildByFieldName("function")
			if fnNode == nil && n.ChildCount() > 0 {
				fnNode = n.Child(0)
			}
			line := int(n.StartPoint().Row) + 1
			if fnNode != nil {
				if fnNode.Type() == "identifier" {
					callee := string(content[fnNode.StartByte():fnNode.EndByte()])
					if !IsGoPredeclaredFunc(callee) {
						if targetID, ok := localFuncs[callee]; ok {
							res.Edges = append(res.Edges, graph.Edge{
								From:       callerID,
								To:         targetID,
								Type:       graph.EdgeCalls,
								Confidence: graph.ConfidenceExtracted,
								Line:       line,
								SourceFile: path,
							})
						} else {
							res.RawCalls = append(res.RawCalls, RawCall{
								CallerID:   callerID,
								CalleeName: callee,
								Line:       line,
								SourceFile: path,
							})
						}
					}
				} else if fnNode.Type() == "selector_expression" {
					field := fnNode.ChildByFieldName("field")
					operand := fnNode.ChildByFieldName("operand")
					if field == nil || operand == nil {
						for j := 0; j < int(fnNode.ChildCount()); j++ {
							c := fnNode.Child(j)
							if c.Type() == "identifier" {
								operand = c
							} else if c.Type() == "field_identifier" {
								field = c
							}
						}
					}
					if field != nil && operand != nil {
						callee := string(content[field.StartByte():field.EndByte()])
						receiver := string(content[operand.StartByte():operand.EndByte()])
						importPath := res.Imports[receiver]
						isMember := importPath == ""

						res.RawCalls = append(res.RawCalls, RawCall{
							CallerID:   callerID,
							CalleeName: callee,
							Receiver:   receiver,
							ImportPath: importPath,
							IsMember:   isMember,
							Line:       line,
							SourceFile: path,
						})
					}
				}
			}
		}

		for i := 0; i < int(n.ChildCount()); i++ {
			walk(n.Child(i))
		}
	}
	walk(node)
}

func parseJS(root *sitter.Node, content []byte, path string, res *ParseResult) {
	fileNID := path
	res.Nodes = append(res.Nodes, graph.Node{
		ID:       fileNID,
		Type:     graph.FileNode,
		Name:     filepath.Base(path),
		FilePath: path,
		Line:     1,
	})

	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		t := n.Type()
		line := int(n.StartPoint().Row) + 1

		switch t {
		case "function_declaration":
			nameNode := n.ChildByFieldName("name")
			if nameNode != nil {
				name := string(content[nameNode.StartByte():nameNode.EndByte()])
				id := fmt.Sprintf("%s:%s", path, name)
				res.Nodes = append(res.Nodes, graph.Node{
					ID:       id,
					Type:     graph.FunctionNode,
					Name:     name,
					FilePath: path,
					Line:     line,
				})
				res.Edges = append(res.Edges, graph.Edge{
					From:       fileNID,
					To:         id,
					Type:       graph.EdgeContains,
					Confidence: graph.ConfidenceExtracted,
					Line:       line,
					SourceFile: path,
				})
			}
		case "class_declaration":
			nameNode := n.ChildByFieldName("name")
			if nameNode != nil {
				name := string(content[nameNode.StartByte():nameNode.EndByte()])
				id := fmt.Sprintf("%s:%s", path, name)
				res.Nodes = append(res.Nodes, graph.Node{
					ID:       id,
					Type:     graph.StructNode,
					Name:     name,
					FilePath: path,
					Line:     line,
				})
				res.Edges = append(res.Edges, graph.Edge{
					From:       fileNID,
					To:         id,
					Type:       graph.EdgeContains,
					Confidence: graph.ConfidenceExtracted,
					Line:       line,
					SourceFile: path,
				})
			}
		}

		for i := 0; i < int(n.ChildCount()); i++ {
			walk(n.Child(i))
		}
	}
	walk(root)
}
