package cluster

import (
	"graphify-go/pkg/graph"
	"path/filepath"
	"sort"
)

type Detector struct {
	graph *graph.Graph
}

func NewDetector(g *graph.Graph) *Detector {
	return &Detector{graph: g}
}

func (d *Detector) Detect() map[int][]string {
	nodes := d.graph.Nodes
	if len(nodes) == 0 {
		return nil
	}

	// Filter out File nodes for clustering, but keep index
	symbolNodes := make([]string, 0)
	nodeIndex := make(map[string]int)
	for _, n := range nodes {
		if n.Type != graph.FileNode {
			nodeIndex[n.ID] = len(symbolNodes)
			symbolNodes = append(symbolNodes, n.ID)
		}
	}

	if len(symbolNodes) == 0 {
		return nil
	}

	// Adjacency and weights
	numNodes := len(symbolNodes)
	adj := make([]map[int]float64, numNodes)
	degrees := make([]float64, numNodes)
	for i := range adj {
		adj[i] = make(map[int]float64)
	}

	var m float64 // total edge weight
	for _, e := range d.graph.Edges {
		if e.Type == graph.EdgeContains {
			continue
		}
		u, okU := nodeIndex[e.From]
		v, okV := nodeIndex[e.To]
		if okU && okV && u != v {
			w := e.Weight
			if w == 0 {
				w = 1.0
			}
			adj[u][v] += w
			adj[v][u] += w
			degrees[u] += w
			degrees[v] += w
			m += w
		}
	}

	// If no edges, fallback to directory-based clustering
	if m == 0 {
		return d.fallbackDirectoryClustering()
	}

	// Community assignment
	comm := make([]int, numNodes)
	commTot := make(map[int]float64)
	for i := range comm {
		comm[i] = i
		commTot[i] = degrees[i]
	}

	twoM := 2.0 * m

	// Run Louvain passes
	improved := true
	for pass := 0; pass < 15 && improved; pass++ {
		improved = false
		for i := 0; i < numNodes; i++ {
			cOld := comm[i]
			kI := degrees[i]
			if kI == 0 {
				continue
			}

			// Remove i from its community
			commTot[cOld] -= kI

			// Calculate weights to neighbor communities
			neighborComms := make(map[int]float64)
			for j, w := range adj[i] {
				neighborComms[comm[j]] += w
			}

			// Find best community
			bestComm := cOld
			bestGain := 0.0

			for c, kIin := range neighborComms {
				tot := commTot[c]
				gain := kIin - (tot * kI / twoM)
				if gain > bestGain {
					bestGain = gain
					bestComm = c
				}
			}

			// Re-insert into best community
			comm[i] = bestComm
			commTot[bestComm] += kI
			if bestComm != cOld {
				improved = true
			}
		}
	}

	// Renumber communities 1..K
	commMap := make(map[int]int)
	nextComm := 1
	for _, c := range comm {
		if _, ok := commMap[c]; !ok {
			commMap[c] = nextComm
			nextComm++
		}
	}

	communities := make(map[int][]string)
	for i, id := range symbolNodes {
		cid := commMap[comm[i]]
		communities[cid] = append(communities[cid], id)
		// Update node in graph
		for j := range d.graph.Nodes {
			if d.graph.Nodes[j].ID == id {
				d.graph.Nodes[j].Community = cid
				break
			}
		}
	}

	// Assign file nodes to community of their contained symbols
	d.assignFileCommunities(communities)

	return communities
}

func (d *Detector) fallbackDirectoryClustering() map[int][]string {
	dirMap := make(map[string]int)
	communities := make(map[int][]string)
	nextComm := 1

	for i := range d.graph.Nodes {
		n := &d.graph.Nodes[i]
		dir := filepath.Dir(n.FilePath)
		if dir == "" || dir == "." {
			dir = "root"
		}
		cid, ok := dirMap[dir]
		if !ok {
			cid = nextComm
			dirMap[dir] = cid
			nextComm++
		}
		n.Community = cid
		communities[cid] = append(communities[cid], n.ID)
	}

	return communities
}

func (d *Detector) assignFileCommunities(communities map[int][]string) {
	for i := range d.graph.Nodes {
		n := &d.graph.Nodes[i]
		if n.Type == graph.FileNode {
			// Find most frequent community among its contained children
			commCounts := make(map[int]int)
			for _, e := range d.graph.Edges {
				if e.From == n.ID && e.Type == graph.EdgeContains {
					child := d.graph.GetNode(e.To)
					if child != nil && child.Community > 0 {
						commCounts[child.Community]++
					}
				}
			}
			maxC := 1
			maxCount := 0
			for c, count := range commCounts {
				if count > maxCount {
					maxCount = count
					maxC = c
				}
			}
			n.Community = maxC
			communities[maxC] = append(communities[maxC], n.ID)
		}
	}

	// Sort community slices for stability
	for c := range communities {
		sort.Strings(communities[c])
	}
}
