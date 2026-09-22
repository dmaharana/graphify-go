package resolver

import (
	"graphify-go/pkg/graph"
	"graphify-go/pkg/parser"
	"strings"
)

type Resolver struct {
	results []*parser.ParseResult

	packageSymbols map[string]map[string][]string
	globalSymbols  map[string][]string
	nodeMap        map[string]graph.Node
	filePackage    map[string]string
}

func NewResolver() *Resolver {
	return &Resolver{
		results:        []*parser.ParseResult{},
		packageSymbols: make(map[string]map[string][]string),
		globalSymbols:  make(map[string][]string),
		nodeMap:        make(map[string]graph.Node),
		filePackage:    make(map[string]string),
	}
}

func (r *Resolver) AddParseResult(res *parser.ParseResult) {
	r.results = append(r.results, res)

	pkg := res.PackageName
	if r.packageSymbols[pkg] == nil {
		r.packageSymbols[pkg] = make(map[string][]string)
	}

	for _, n := range res.Nodes {
		r.nodeMap[n.ID] = n
		if n.FilePath != "" {
			r.filePackage[n.FilePath] = pkg
		}
		if n.Type != graph.FileNode {
			r.packageSymbols[pkg][n.Name] = append(r.packageSymbols[pkg][n.Name], n.ID)
			r.globalSymbols[n.Name] = append(r.globalSymbols[n.Name], n.ID)
		}
	}
}

func (r *Resolver) Resolve() *graph.Graph {
	g := graph.NewGraph()

	// Add all nodes
	for _, n := range r.nodeMap {
		g.AddNode(n)
	}

	// Add existing edges (rewiring bare embeds / references if needed)
	seenEdges := make(map[string]bool)
	for _, res := range r.results {
		pkg := res.PackageName
		for _, e := range res.Edges {
			target := e.To
			if e.Type == graph.EdgeEmbeds || e.Type == graph.EdgeReferences {
				if _, ok := r.nodeMap[target]; !ok {
					if ids, ok := r.packageSymbols[pkg][target]; ok && len(ids) > 0 {
						target = ids[0]
					} else if ids, ok := r.globalSymbols[target]; ok && len(ids) > 0 {
						target = ids[0]
					}
				}
			}
			edgeKey := e.From + "->" + target + ":" + string(e.Type)
			if !seenEdges[edgeKey] {
				seenEdges[edgeKey] = true
				g.AddEdge(e.From, target, e.Type, e.Confidence, e.Line, e.SourceFile)
			}
		}
	}

	// Resolve RawCalls
	for _, res := range r.results {
		callerPkg := res.PackageName
		for _, rc := range res.RawCalls {
			var targetID string

			if !rc.IsMember && rc.Receiver != "" {
				targetPkg := rc.Receiver
				if syms, ok := r.packageSymbols[targetPkg]; ok {
					if ids, ok := syms[rc.CalleeName]; ok && len(ids) > 0 {
						targetID = ids[0]
					}
				}
				if targetID == "" && rc.ImportPath != "" {
					parts := strings.Split(rc.ImportPath, "/")
					lastPart := parts[len(parts)-1]
					if syms, ok := r.packageSymbols[lastPart]; ok {
						if ids, ok := syms[rc.CalleeName]; ok && len(ids) > 0 {
							targetID = ids[0]
						}
					}
				}
			} else if rc.Receiver == "" {
				if syms, ok := r.packageSymbols[callerPkg]; ok {
					if ids, ok := syms[rc.CalleeName]; ok && len(ids) > 0 {
						targetID = ids[0]
					}
				}
			} else if rc.IsMember {
				if ids, ok := r.packageSymbols[callerPkg][rc.CalleeName]; ok && len(ids) > 0 {
					targetID = ids[0]
				} else if ids, ok := r.globalSymbols[rc.CalleeName]; ok && len(ids) == 1 {
					targetID = ids[0]
				}
			}

			if targetID != "" && targetID != rc.CallerID {
				edgeKey := rc.CallerID + "->" + targetID + ":" + string(graph.EdgeCalls)
				if !seenEdges[edgeKey] {
					seenEdges[edgeKey] = true
					g.AddEdge(rc.CallerID, targetID, graph.EdgeCalls, graph.ConfidenceInferred, rc.Line, rc.SourceFile)
				}
			}
		}
	}

	return g
}
