package aiengine

// GraphEngine is a faithful Go port of app/core/graph_engine.py.
//
// Every *_graph.json under the embedded data/domains tree is parsed. Files
// without a valid top-level domain_id are skipped (like Python's
// _extract_domain_id raise-and-continue). A topological order is computed per
// domain with a deterministic Kahn's algorithm using a lexicographic min-heap,
// preserving the strict dependency ordering the Python networkx topo sort
// guarantees (networkx is deterministic for the same insertion order).

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// GraphNode is a decoded domain-graph node. Raw preserves every original field
// so personalized-path responses match the FastAPI payloads exactly.
type GraphNode struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	AssessmentRubric map[string]any `json:"-"`
	Raw              map[string]any `json:"-"`
}

// DomainGraph carries a parsed domain graph plus its metadata payload.
type DomainGraph struct {
	DomainID    string
	DomainName  string
	SourceFile  string
	Nodes       []*GraphNode
	NodeByID    map[string]*GraphNode
	hardPrereqs map[string][]string
	metadata    map[string]any
}

type GraphEngine struct {
	DomainGraphs map[string]*DomainGraph
	DomainList   []string
}

// ParseGraphEngine mirrors GraphEngine.__init__() + _load_domains().
func ParseGraphEngine() (*GraphEngine, error) {
	engine := &GraphEngine{
		DomainGraphs: map[string]*DomainGraph{},
	}

	graphFiles, err := listGraphFiles()
	if err != nil {
		return nil, err
	}
	if len(graphFiles) == 0 {
		return nil, fmt.Errorf("no domain graph files found under '%s'", domainDataRoot)
	}

	validCount := 0
	for _, file := range graphFiles {
		payload, err := loadGraphPayload(file)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON in domain file '%s': %w", file, err)
		}

		domainID, err := extractDomainID(payload, file)
		if err != nil {
			continue // skip, matching Python behaviour
		}
		validCount++

		graph := &DomainGraph{
			DomainID:    domainID,
			DomainName:  stringAny(payload["domain_name"]),
			SourceFile:  file,
			NodeByID:    map[string]*GraphNode{},
			hardPrereqs: map[string][]string{},
			metadata:    payload,
		}

		nodesRaw, ok := payload["nodes"].([]any)
		if !ok {
			return nil, fmt.Errorf("domain '%s' has an invalid 'nodes' payload; expected a list", domainID)
		}
		seen := map[string]bool{}
		for _, n := range nodesRaw {
			nodeMap, ok := n.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("domain '%s' contains a non-object node entry", domainID)
			}
			nodeID, _ := nodeMap["id"].(string)
			if nodeID == "" {
				return nil, fmt.Errorf("domain '%s' contains a node without a valid string 'id' field", domainID)
			}
			if seen[nodeID] {
				return nil, fmt.Errorf("duplicate node ID '%s' found in domain '%s'", nodeID, domainID)
			}
			seen[nodeID] = true
			gn := &GraphNode{
				ID:               nodeID,
				Name:             stringAny(nodeMap["name"]),
				AssessmentRubric: rubricOrNil(nodeMap),
				Raw:              nodeMap,
			}
			graph.Nodes = append(graph.Nodes, gn)
			graph.NodeByID[nodeID] = gn
		}
		engine.DomainGraphs[domainID] = graph

		// Edges from prerequisites.hard (same validation as Python).
		for _, nodeMapRaw := range nodesRaw {
			nodeMap := nodeMapRaw.(map[string]any)
			nodeID, _ := nodeMap["id"].(string)
			prereqsRaw, _ := nodeMap["prerequisites"].(map[string]any)
			if prereqsRaw == nil {
				continue
			}
			hardRaw, _ := prereqsRaw["hard"].([]any)
			for _, depRaw := range hardRaw {
				dep, _ := depRaw.(string)
				if dep == "" {
					return nil, fmt.Errorf("node '%s' in domain '%s' contains a non-string prerequisite ID", nodeID, domainID)
				}
				if _, ok := graph.NodeByID[dep]; !ok {
					return nil, fmt.Errorf("node '%s' in domain '%s' references missing prerequisite '%s'", nodeID, domainID, dep)
				}
				if dep == nodeID {
					return nil, fmt.Errorf("node '%s' in domain '%s' is self-referencing as a prerequisite", nodeID, domainID)
				}
				graph.hardPrereqs[nodeID] = append(graph.hardPrereqs[nodeID], dep)
			}
		}

		if !isDAG(graph.hardPrereqs, graph.NodeByID) {
			return nil, fmt.Errorf("circular dependency detected in domain '%s'; the graph must be a valid DAG", domainID)
		}
	}

	if validCount == 0 {
		return nil, fmt.Errorf("no valid domain graph payloads were found under '%s'; each graph file must contain a top-level 'domain_id' field", domainDataRoot)
	}

	ids := make([]string, 0, len(engine.DomainGraphs))
	for id := range engine.DomainGraphs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	engine.DomainList = ids
	return engine, nil
}

func listGraphFiles() ([]string, error) {
	var files []string
	dirs, err := dataFS.ReadDir(domainDataRoot)
	if err != nil {
		return nil, err
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		sub, err := dataFS.ReadDir(domainDataRoot + "/" + dir.Name())
		if err != nil {
			continue
		}
		for _, f := range sub {
			if strings.HasSuffix(f.Name(), "_graph.json") {
				files = append(files, domainDataRoot+"/"+dir.Name()+"/"+f.Name())
			}
		}
	}
	sort.Strings(files)
	return files, nil
}

func loadGraphPayload(path string) (map[string]any, error) {
	raw, err := dataFS.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func extractDomainID(payload map[string]any, file string) (string, error) {
	domainID, _ := payload["domain_id"].(string)
	domainID = strings.TrimSpace(domainID)
	if domainID == "" {
		return "", fmt.Errorf("missing or invalid 'domain_id' in file '%s'", file)
	}
	return domainID, nil
}

// isDAG runs Kahn's algorithm; returns false when a cycle is present.
func isDAG(prereqs map[string][]string, nodes map[string]*GraphNode) bool {
	indegree := map[string]int{}
	for id := range nodes {
		indegree[id] = 0
	}
	adj := map[string][]string{}
	for child, deps := range prereqs {
		for _, dep := range deps {
			adj[dep] = append(adj[dep], child)
			indegree[child]++
		}
	}
	var queue []string
	for id, deg := range indegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)
	count := 0
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		count++
		for _, child := range sortedStrings(adj[id]) {
			indegree[child]--
			if indegree[child] == 0 {
				queue = append(queue, child)
				sort.Strings(queue)
			}
		}
	}
	return count == len(nodes)
}

func sortedStrings(s []string) []string {
	out := append([]string(nil), s...)
	sort.Strings(out)
	return out
}

// HardPrereqs returns the hard prerequisite node ids for a node.
func (g *DomainGraph) HardPrereqs(nodeID string) []string {
	return append([]string(nil), g.hardPrereqs[nodeID]...)
}

// TopoOrder returns a deterministic topological ordering for a domain graph.
func (g *DomainGraph) TopoOrder() []string {
	indegree := map[string]int{}
	for _, n := range g.Nodes {
		indegree[n.ID] = 0
	}
	adj := map[string][]string{}
	for child, deps := range g.hardPrereqs {
		for _, dep := range deps {
			adj[dep] = append(adj[dep], child)
			indegree[child]++
		}
	}
	heap := &stringHeap{}
	for id, deg := range indegree {
		if deg == 0 {
			heap.Push(id)
		}
	}
	var order []string
	for heap.Len() > 0 {
		id := heap.Pop()
		order = append(order, id)
		for _, child := range sortedStrings(adj[id]) {
			indegree[child]--
			if indegree[child] == 0 {
				heap.Push(child)
			}
		}
	}
	return order
}

// GetPersonalizedPath mirrors get_personalized_path(): topo order minus
// completed nodes, returning full node dicts.
func (g *GraphEngine) GetPersonalizedPath(domainID string, completedNodes []string) ([]map[string]any, error) {
	graph, ok := g.DomainGraphs[domainID]
	if !ok {
		valid := make([]string, 0, len(g.DomainGraphs))
		for id := range g.DomainGraphs {
			valid = append(valid, id)
		}
		sort.Strings(valid)
		list := "none"
		if len(valid) > 0 {
			list = strings.Join(valid, ", ")
		}
		return nil, fmt.Errorf("domain '%s' does not exist. Available domains: %s", domainID, list)
	}

	completed := map[string]bool{}
	for _, id := range completedNodes {
		completed[id] = true
	}

	var ordered []map[string]any
	for _, id := range graph.TopoOrder() {
		if completed[id] {
			continue
		}
		if node, ok := graph.NodeByID[id]; ok {
			ordered = append(ordered, cloneMap(node.Raw))
		}
	}
	return ordered, nil
}

func (g *GraphEngine) GetNode(domainID, nodeID string) (*GraphNode, error) {
	graph, ok := g.DomainGraphs[domainID]
	if !ok {
		return nil, fmt.Errorf("node '%s' not found in domain '%s'", nodeID, domainID)
	}
	node, ok := graph.NodeByID[nodeID]
	if !ok {
		return nil, fmt.Errorf("node '%s' not found in domain '%s'", nodeID, domainID)
	}
	return node, nil
}

func (g *GraphEngine) DomainExists(domainID string) bool {
	_, ok := g.DomainGraphs[domainID]
	return ok
}

// NodeResources returns the typed resources slice from a node's raw payload.
func (n *GraphNode) NodeResources() []map[string]any {
	resources, _ := n.Raw["resources"].([]any)
	var out []map[string]any
	for _, r := range resources {
		if m, ok := r.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func rubricOrNil(node map[string]any) map[string]any {
	if r, ok := node["assessment_rubric"].(map[string]any); ok {
		return r
	}
	return map[string]any{}
}

func stringAny(v any) string {
	s, _ := v.(string)
	return s
}

func cloneMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// stringHeap is a min-heap of strings for lexicographic Kahn ordering.
type stringHeap struct{ items []string }

func (h *stringHeap) Push(s string) { h.items = append(h.items, s) }
func (h *stringHeap) Pop() string {
	best := 0
	for i := 1; i < len(h.items); i++ {
		if h.items[i] < h.items[best] {
			best = i
		}
	}
	s := h.items[best]
	h.items = append(h.items[:best], h.items[best+1:]...)
	return s
}
func (h *stringHeap) Len() int { return len(h.items) }
