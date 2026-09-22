package report

import (
	"fmt"
	"graphify-go/pkg/graph"
	"path/filepath"
	"sort"
	"strings"
)

func GenerateReport(g *graph.Graph, projectName string) string {
	var sb strings.Builder

	fileCount := 0
	symbolCount := 0
	communitySet := make(map[int]bool)
	for _, n := range g.Nodes {
		if n.Type == graph.FileNode {
			fileCount++
		} else {
			symbolCount++
		}
		if n.Community > 0 {
			communitySet[n.Community] = true
		}
	}

	sb.WriteString(fmt.Sprintf("# Architecture Audit Report: %s\n\n", projectName))
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Files Scanned**: %d\n", fileCount))
	sb.WriteString(fmt.Sprintf("- **Symbol Nodes**: %d\n", symbolCount))
	sb.WriteString(fmt.Sprintf("- **Relationships Extracted**: %d\n", len(g.Edges)))
	sb.WriteString(fmt.Sprintf("- **Communities Detected**: %d\n\n", len(communitySet)))

	// God Nodes
	sb.WriteString("## God Nodes (Architectural Hubs)\n\n")
	sb.WriteString("Top central components that the system flows through:\n\n")
	gods := g.GodNodes(8)
	if len(gods) == 0 {
		sb.WriteString("_No high-degree hubs detected._\n\n")
	} else {
		for i, gn := range gods {
			sb.WriteString(fmt.Sprintf("%d. **%s** (`%s`) - Degree: %d\n", i+1, gn.Node.Name, gn.Node.Type, gn.Degree))
			if gn.Node.FilePath != "" {
				sb.WriteString(fmt.Sprintf("   - Location: `%s`", gn.Node.FilePath))
				if gn.Node.Line > 0 {
					sb.WriteString(fmt.Sprintf(" (L%d)", gn.Node.Line))
				}
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n")
	}

	// Communities
	sb.WriteString("## Communities (Subsystems)\n\n")
	commNodes := make(map[int][]graph.Node)
	for _, n := range g.Nodes {
		if n.Type != graph.FileNode && n.Community > 0 {
			commNodes[n.Community] = append(commNodes[n.Community], n)
		}
	}

	var commIDs []int
	for cid := range commNodes {
		commIDs = append(commIDs, cid)
	}
	sort.Ints(commIDs)

	for _, cid := range commIDs {
		nodes := commNodes[cid]
		// Find dominant directory
		dirCounts := make(map[string]int)
		for _, n := range nodes {
			d := filepath.Dir(n.FilePath)
			dirCounts[d]++
		}
		dominantDir := "root"
		maxD := 0
		for d, count := range dirCounts {
			if count > maxD {
				maxD = count
				dominantDir = d
			}
		}

		sb.WriteString(fmt.Sprintf("### Community %d: `%s` (%d symbols)\n\n", cid, dominantDir, len(nodes)))
		limit := 6
		if len(nodes) < limit {
			limit = len(nodes)
		}
		var sampleNames []string
		for i := 0; i < limit; i++ {
			sampleNames = append(sampleNames, fmt.Sprintf("`%s`", nodes[i].Name))
		}
		sb.WriteString("Key symbols: " + strings.Join(sampleNames, ", "))
		if len(nodes) > limit {
			sb.WriteString(fmt.Sprintf(" ... and %d more", len(nodes)-limit))
		}
		sb.WriteString("\n\n")
	}

	// Surprising Connections
	sb.WriteString("## Surprising Connections (Cross-Boundary Links)\n\n")
	type crossEdge struct {
		from     string
		to       string
		edgeType string
		cFrom    int
		cTo      int
	}
	var crossLinks []crossEdge
	for _, e := range g.Edges {
		if e.Type == graph.EdgeContains {
			continue
		}
		fromNode := g.GetNode(e.From)
		toNode := g.GetNode(e.To)
		if fromNode != nil && toNode != nil && fromNode.Community > 0 && toNode.Community > 0 && fromNode.Community != toNode.Community {
			crossLinks = append(crossLinks, crossEdge{
				from:     fromNode.Name,
				to:       toNode.Name,
				edgeType: string(e.Type),
				cFrom:    fromNode.Community,
				cTo:      toNode.Community,
			})
		}
	}

	if len(crossLinks) == 0 {
		sb.WriteString("_No inter-community cross-boundary calls detected._\n\n")
	} else {
		maxLinks := 8
		if len(crossLinks) < maxLinks {
			maxLinks = len(crossLinks)
		}
		for i := 0; i < maxLinks; i++ {
			cl := crossLinks[i]
			sb.WriteString(fmt.Sprintf("- **%s** (Comm %d) --[%s]--> **%s** (Comm %d)\n", cl.from, cl.cFrom, cl.edgeType, cl.to, cl.cTo))
		}
		sb.WriteString("\n")
	}

	// Suggested Questions
	sb.WriteString("## Suggested Questions\n\n")
	if len(gods) > 0 {
		sb.WriteString(fmt.Sprintf("1. `graphify-go explain \"%s\"`: Explore dependencies of the primary hub.\n", gods[0].Node.Name))
	}
	if len(gods) > 1 {
		sb.WriteString(fmt.Sprintf("2. `graphify-go path \"%s\" \"%s\"`: Trace how the top architectural hubs interact.\n", gods[0].Node.Name, gods[1].Node.Name))
	}
	if len(crossLinks) > 0 {
		sb.WriteString(fmt.Sprintf("3. `graphify-go affected \"%s\"`: Determine blast radius if the bridge symbol changes.\n", crossLinks[0].from))
	}

	return sb.String()
}
