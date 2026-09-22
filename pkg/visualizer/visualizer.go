package visualizer

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"graphify-go/pkg/graph"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed template.html
var htmlTemplate string

var CommunityColors = []string{
	"#4E79A7", "#F28E2B", "#E15759", "#76B7B2", "#59A14F",
	"#EDC948", "#B07AA1", "#FF9DA7", "#9C755F", "#BAB0AC",
	"#8CD17D", "#B6992D", "#499894", "#E15759", "#79706E",
}

type VisNode struct {
	ID            string            `json:"id"`
	Label         string            `json:"label"`
	Color         map[string]string `json:"color"`
	Size          float64           `json:"size"`
	Font          map[string]any    `json:"font"`
	Title         string            `json:"title"`
	Community     int               `json:"community"`
	CommunityName string            `json:"community_name"`
	SourceFile    string            `json:"source_file"`
	FileType      string            `json:"file_type"`
	Degree        int               `json:"degree"`
}

type VisEdge struct {
	From       string         `json:"from"`
	To         string         `json:"to"`
	Label      string         `json:"label"`
	Title      string         `json:"title"`
	Dashes     bool           `json:"dashes"`
	Width      int            `json:"width"`
	Color      map[string]any `json:"color"`
	Confidence string         `json:"confidence"`
}

type LegendItem struct {
	CID   int    `json:"cid"`
	Color string `json:"color"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

func GenerateHTML(g *graph.Graph, outputPath string) error {
	nodeSet := make(map[string]bool)
	degreeMap := make(map[string]int)

	for _, n := range g.Nodes {
		nodeSet[n.ID] = true
	}

	for _, e := range g.Edges {
		if nodeSet[e.From] && nodeSet[e.To] {
			degreeMap[e.From]++
			degreeMap[e.To]++
		}
	}

	maxDeg := 1
	for _, deg := range degreeMap {
		if deg > maxDeg {
			maxDeg = deg
		}
	}

	// Communities mapping
	commDirCounts := make(map[int]map[string]int)
	commNodeCounts := make(map[int]int)

	for _, n := range g.Nodes {
		cid := n.Community
		if cid == 0 {
			cid = 1
		}
		commNodeCounts[cid]++
		if commDirCounts[cid] == nil {
			commDirCounts[cid] = make(map[string]int)
		}
		d := filepath.Dir(n.FilePath)
		if d == "" || d == "." {
			d = "root"
		}
		commDirCounts[cid][d]++
	}

	commNames := make(map[int]string)
	for cid, dCounts := range commDirCounts {
		topDir := "root"
		maxD := 0
		for dir, count := range dCounts {
			if count > maxD {
				maxD = count
				topDir = dir
			}
		}
		commNames[cid] = fmt.Sprintf("Community %d (%s)", cid, topDir)
	}

	visNodes := make([]VisNode, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		cid := n.Community
		if cid == 0 {
			cid = 1
		}
		color := CommunityColors[cid%len(CommunityColors)]
		deg := degreeMap[n.ID]

		size := 10.0 + 25.0*(float64(deg)/float64(maxDeg))
		if n.Type == graph.FileNode {
			size = 8.0
		}

		fontSize := 11
		if deg < maxDeg/4 && n.Type != graph.FileNode {
			fontSize = 0 // Hide small node labels until zoomed
		}

		visNodes = append(visNodes, VisNode{
			ID:    n.ID,
			Label: n.Name,
			Color: map[string]string{
				"background": color,
				"border":     color,
				"highlight":  "#ffffff",
			},
			Size: size,
			Font: map[string]any{
				"size":  fontSize,
				"color": "#ffffff",
			},
			Title:         fmt.Sprintf("%s (%s)\n%s", n.Name, n.Type, n.FilePath),
			Community:     cid,
			CommunityName: commNames[cid],
			SourceFile:    n.FilePath,
			FileType:      string(n.Type),
			Degree:        deg,
		})
	}

	visEdges := make([]VisEdge, 0, len(g.Edges))
	for _, e := range g.Edges {
		// Crucial: Only include edges where both endpoints exist in nodeSet!
		if !nodeSet[e.From] || !nodeSet[e.To] {
			continue
		}

		dashes := e.Confidence != graph.ConfidenceExtracted
		width := 1
		opacity := 0.4
		if e.Confidence == graph.ConfidenceExtracted {
			width = 2
			opacity = 0.7
		}

		visEdges = append(visEdges, VisEdge{
			From:       e.From,
			To:         e.To,
			Label:      string(e.Type),
			Title:      fmt.Sprintf("%s [%s]", e.Type, e.Confidence),
			Dashes:     dashes,
			Width:      width,
			Color:      map[string]any{"opacity": opacity},
			Confidence: string(e.Confidence),
		})
	}

	// Legend items
	var commList []int
	for cid := range commNodeCounts {
		commList = append(commList, cid)
	}
	sort.Ints(commList)

	legend := make([]LegendItem, 0, len(commList))
	for _, cid := range commList {
		legend = append(legend, LegendItem{
			CID:   cid,
			Color: CommunityColors[cid%len(CommunityColors)],
			Label: commNames[cid],
			Count: commNodeCounts[cid],
		})
	}

	nodesJSON, err := json.Marshal(visNodes)
	if err != nil {
		return err
	}
	edgesJSON, err := json.Marshal(visEdges)
	if err != nil {
		return err
	}
	legendJSON, err := json.Marshal(legend)
	if err != nil {
		return err
	}

	stats := fmt.Sprintf("%d nodes &middot; %d relationships &middot; %d communities",
		len(visNodes), len(visEdges), len(legend))

	htmlContent := htmlTemplate
	htmlContent = strings.Replace(htmlContent, "{{NODES_JSON}}", string(nodesJSON), 1)
	htmlContent = strings.Replace(htmlContent, "{{EDGES_JSON}}", string(edgesJSON), 1)
	htmlContent = strings.Replace(htmlContent, "{{LEGEND_JSON}}", string(legendJSON), 1)
	htmlContent = strings.Replace(htmlContent, "{{STATS}}", stats, 1)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	return os.WriteFile(outputPath, []byte(htmlContent), 0644)
}
