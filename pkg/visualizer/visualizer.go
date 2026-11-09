package visualizer

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"graphify-go/pkg/graph"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed template.html
var htmlTemplate string

//go:embed vis-network.min.js
var visNetworkJS string

var CommunityColors = []string{
	"#4E79A7", "#F28E2B", "#E15759", "#76B7B2", "#59A14F",
	"#EDC948", "#B07AA1", "#FF9DA7", "#9C755F", "#BAB0AC",
	"#8CD17D", "#B6992D", "#499894", "#E15759", "#79706E",
}

const GoldenAngle = 2.399963229728653 // pi * (3 - sqrt(5))

type Config struct {
	MaxNodes int    `json:"max_nodes"`
	Layout   string `json:"layout"` // "force" or "hierarchical"
	Theme    string `json:"theme"`  // "dark", "light", "auto"
}

func DefaultConfig() Config {
	return Config{
		MaxNodes: 1000,
		Layout:   "force",
		Theme:    "dark",
	}
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
	X             float64           `json:"x"`
	Y             float64           `json:"y"`
	IsGodNode     bool              `json:"is_god_node,omitempty"`
	Line          int               `json:"line,omitempty"`
	Shape         string            `json:"shape,omitempty"`
	BorderWidth   int               `json:"borderWidth,omitempty"`
}

type VisEdge struct {
	From       string         `json:"from"`
	To         string         `json:"to"`
	Label      string         `json:"label"`
	Title      string         `json:"title"`
	Dashes     any            `json:"dashes"`
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

type VizMetadata struct {
	TotalNodes   int    `json:"total_nodes"`
	VisibleNodes int    `json:"visible_nodes"`
	TotalEdges   int    `json:"total_edges"`
	VisibleEdges int    `json:"visible_edges"`
	IsPruned     bool   `json:"is_pruned"`
	MaxNodes     int    `json:"max_nodes"`
	Layout       string `json:"layout"`
	Theme        string `json:"theme"`
}

type PruneResult struct {
	Nodes       []graph.Node
	Edges       []graph.Edge
	TotalNodes  int
	TotalEdges  int
	PrunedNodes int
	IsPruned    bool
}

// PruneGraph retains at least one representative from every community,
// and fills the remaining capacity up to maxNodes with the highest-degree nodes.
func PruneGraph(g *graph.Graph, maxNodes int) PruneResult {
	totalNodes := len(g.Nodes)
	totalEdges := len(g.Edges)

	if maxNodes <= 0 || totalNodes <= maxNodes {
		return PruneResult{
			Nodes:       g.Nodes,
			Edges:       g.Edges,
			TotalNodes:  totalNodes,
			TotalEdges:  totalEdges,
			PrunedNodes: 0,
			IsPruned:    false,
		}
	}

	nodeSet := make(map[string]bool, totalNodes)
	for _, n := range g.Nodes {
		nodeSet[n.ID] = true
	}

	degreeMap := make(map[string]int, totalNodes)
	for _, e := range g.Edges {
		if nodeSet[e.From] && nodeSet[e.To] {
			degreeMap[e.From]++
			degreeMap[e.To]++
		}
	}

	// Group nodes by community
	commNodes := make(map[int][]graph.Node)
	for _, n := range g.Nodes {
		cid := n.Community
		if cid == 0 {
			cid = 1
		}
		commNodes[cid] = append(commNodes[cid], n)
	}

	// For each community, find the highest degree node as representative
	repMap := make(map[string]bool)
	var repList []graph.Node
	for _, nodes := range commNodes {
		var bestNode graph.Node
		bestDeg := -1
		for _, n := range nodes {
			deg := degreeMap[n.ID]
			if deg > bestDeg || (deg == bestDeg && n.ID < bestNode.ID) {
				bestDeg = deg
				bestNode = n
			}
		}
		if bestDeg >= 0 {
			repMap[bestNode.ID] = true
			repList = append(repList, bestNode)
		}
	}

	selectedSet := make(map[string]bool, maxNodes)

	// If reps exceed maxNodes, take the highest degree reps
	if len(repList) >= maxNodes {
		sort.Slice(repList, func(i, j int) bool {
			degI := degreeMap[repList[i].ID]
			degJ := degreeMap[repList[j].ID]
			if degI != degJ {
				return degI > degJ
			}
			return repList[i].ID < repList[j].ID
		})
		for i := 0; i < maxNodes; i++ {
			selectedSet[repList[i].ID] = true
		}
	} else {
		// Include all community representatives first
		for id := range repMap {
			selectedSet[id] = true
		}

		// Sort all remaining nodes by degree descending
		var remaining []graph.Node
		for _, n := range g.Nodes {
			if !selectedSet[n.ID] {
				remaining = append(remaining, n)
			}
		}

		sort.Slice(remaining, func(i, j int) bool {
			degI := degreeMap[remaining[i].ID]
			degJ := degreeMap[remaining[j].ID]
			if degI != degJ {
				return degI > degJ
			}
			return remaining[i].ID < remaining[j].ID
		})

		needed := maxNodes - len(selectedSet)
		for i := 0; i < needed && i < len(remaining); i++ {
			selectedSet[remaining[i].ID] = true
		}
	}

	prunedNodes := make([]graph.Node, 0, len(selectedSet))
	for _, n := range g.Nodes {
		if selectedSet[n.ID] {
			prunedNodes = append(prunedNodes, n)
		}
	}

	prunedEdges := make([]graph.Edge, 0, len(g.Edges))
	for _, e := range g.Edges {
		if selectedSet[e.From] && selectedSet[e.To] {
			prunedEdges = append(prunedEdges, e)
		}
	}

	return PruneResult{
		Nodes:       prunedNodes,
		Edges:       prunedEdges,
		TotalNodes:  totalNodes,
		TotalEdges:  totalEdges,
		PrunedNodes: totalNodes - len(prunedNodes),
		IsPruned:    true,
	}
}

// SeedNodePositions seeds initial coordinates using a deterministic community-aware
// Fermat sunflower spiral layout.
func SeedNodePositions(nodes []VisNode) {
	if len(nodes) == 0 {
		return
	}

	commIndices := make(map[int][]int)
	for i, n := range nodes {
		commIndices[n.Community] = append(commIndices[n.Community], i)
	}

	var commIDs []int
	for cid := range commIndices {
		commIDs = append(commIDs, cid)
	}
	sort.Ints(commIDs)

	k := len(commIDs)
	commRadius := 350.0 * math.Sqrt(float64(k))
	if k <= 1 {
		commRadius = 0
	}

	commCenters := make(map[int][2]float64, k)
	for idx, cid := range commIDs {
		if k == 1 {
			commCenters[cid] = [2]float64{0, 0}
		} else {
			angle := 2.0 * math.Pi * float64(idx) / float64(k)
			commCenters[cid] = [2]float64{
				math.Round(commRadius*math.Cos(angle)*10) / 10,
				math.Round(commRadius*math.Sin(angle)*10) / 10,
			}
		}
	}

	for _, cid := range commIDs {
		indices := commIndices[cid]
		sort.Slice(indices, func(i, j int) bool {
			idxI := indices[i]
			idxJ := indices[j]
			if nodes[idxI].Degree != nodes[idxJ].Degree {
				return nodes[idxI].Degree > nodes[idxJ].Degree
			}
			return nodes[idxI].ID < nodes[idxJ].ID
		})

		center := commCenters[cid]
		spacing := 45.0

		for order, nodeIdx := range indices {
			if order == 0 {
				nodes[nodeIdx].X = center[0]
				nodes[nodeIdx].Y = center[1]
				continue
			}
			r := spacing * math.Sqrt(float64(order))
			theta := float64(order) * GoldenAngle
			nodes[nodeIdx].X = math.Round((center[0]+r*math.Cos(theta))*10) / 10
			nodes[nodeIdx].Y = math.Round((center[1]+r*math.Sin(theta))*10) / 10
		}
	}
}

func GenerateHTML(g *graph.Graph, outputPath string) error {
	return GenerateHTMLWithConfig(g, outputPath, DefaultConfig())
}

func GenerateHTMLWithConfig(g *graph.Graph, outputPath string, cfg Config) error {
	if cfg.MaxNodes <= 0 {
		cfg.MaxNodes = 1000
	}
	if cfg.Layout == "" {
		cfg.Layout = "force"
	}
	if cfg.Theme == "" {
		cfg.Theme = "dark"
	}

	pruneRes := PruneGraph(g, cfg.MaxNodes)
	renderNodes := pruneRes.Nodes
	renderEdges := pruneRes.Edges

	nodeSet := make(map[string]bool, len(renderNodes))
	degreeMap := make(map[string]int, len(renderNodes))

	for _, n := range renderNodes {
		nodeSet[n.ID] = true
	}

	for _, e := range renderEdges {
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

	// Identify god nodes for distinct visual treatment
	godNodes := g.GodNodes(8)
	godMap := make(map[string]bool, len(godNodes))
	for _, gn := range godNodes {
		godMap[gn.Node.ID] = true
	}

	// Communities mapping
	commDirCounts := make(map[int]map[string]int)
	commNodeCounts := make(map[int]int)

	for _, n := range renderNodes {
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

	visNodes := make([]VisNode, 0, len(renderNodes))
	for _, n := range renderNodes {
		cid := n.Community
		if cid == 0 {
			cid = 1
		}
		color := CommunityColors[cid%len(CommunityColors)]
		deg := degreeMap[n.ID]

		size := 12.0 + 28.0*math.Sqrt(float64(deg)/float64(maxDeg))
		if n.Type == graph.FileNode {
			size = 10.0
		}
		if godMap[n.ID] && size < 24.0 {
			size = 24.0
		}

		fontSize := 11
		if deg < maxDeg/4 && n.Type != graph.FileNode {
			fontSize = 0 // Hide small node labels until zoomed
		}

		borderColor := color
		borderWidth := 1
		isGod := godMap[n.ID]
		if isGod {
			borderColor = "#FFD700"
			borderWidth = 3
		}

		visNodes = append(visNodes, VisNode{
			ID:    n.ID,
			Label: n.Name,
			Color: map[string]string{
				"background": color,
				"border":     borderColor,
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
			IsGodNode:     isGod,
			Line:          n.Line,
			BorderWidth:   borderWidth,
		})
	}

	// Deterministic layout seeding
	SeedNodePositions(visNodes)

	visEdges := make([]VisEdge, 0, len(renderEdges))
	for _, e := range renderEdges {
		if !nodeSet[e.From] || !nodeSet[e.To] {
			continue
		}

		var dashes any = false
		width := 1
		opacity := 0.4
		edgeColor := "#3a3a5e"

		switch e.Confidence {
		case graph.ConfidenceExtracted:
			width = 2
			opacity = 0.75
		case graph.ConfidenceInferred:
			dashes = []int{5, 5}
			width = 1
			opacity = 0.5
		case graph.ConfidenceAmbiguous:
			dashes = []int{2, 4}
			width = 1
			opacity = 0.65
			edgeColor = "#E5A93C"
		default:
			dashes = true
		}

		visEdges = append(visEdges, VisEdge{
			From:       e.From,
			To:         e.To,
			Label:      string(e.Type),
			Title:      fmt.Sprintf("%s -> %s\nType: %s\nConfidence: %s", e.From, e.To, e.Type, e.Confidence),
			Dashes:     dashes,
			Width:      width,
			Color:      map[string]any{"color": edgeColor, "opacity": opacity},
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

	meta := VizMetadata{
		TotalNodes:   pruneRes.TotalNodes,
		VisibleNodes: len(visNodes),
		TotalEdges:   pruneRes.TotalEdges,
		VisibleEdges: len(visEdges),
		IsPruned:     pruneRes.IsPruned,
		MaxNodes:     cfg.MaxNodes,
		Layout:       cfg.Layout,
		Theme:        cfg.Theme,
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
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	stats := fmt.Sprintf("%d nodes &middot; %d relationships &middot; %d communities",
		len(visNodes), len(visEdges), len(legend))
	if pruneRes.IsPruned {
		stats = fmt.Sprintf("Showing %d of %d nodes &middot; %d edges &middot; %d communities",
			len(visNodes), pruneRes.TotalNodes, len(visEdges), len(legend))
	}

	htmlContent := htmlTemplate
	htmlContent = strings.Replace(htmlContent, "{{VIS_NETWORK_JS}}", visNetworkJS, 1)
	htmlContent = strings.Replace(htmlContent, "{{NODES_JSON}}", string(nodesJSON), 1)
	htmlContent = strings.Replace(htmlContent, "{{EDGES_JSON}}", string(edgesJSON), 1)
	htmlContent = strings.Replace(htmlContent, "{{LEGEND_JSON}}", string(legendJSON), 1)
	htmlContent = strings.Replace(htmlContent, "{{META_JSON}}", string(metaJSON), 1)
	htmlContent = strings.Replace(htmlContent, "{{STATS}}", stats, 1)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	return os.WriteFile(outputPath, []byte(htmlContent), 0644)
}
