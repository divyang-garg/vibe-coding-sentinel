// Package architecture_cycles - Cycle detection on ModuleGraph for architecture analysis
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package services

// findCyclesInModuleGraph returns all cycles in the module graph (list of file paths per cycle).
func findCyclesInModuleGraph(graph ModuleGraph) [][]string {
	adj := make(map[string][]string)
	for _, e := range graph.Edges {
		if e.From != "" && e.To != "" {
			adj[e.From] = append(adj[e.From], e.To)
		}
	}

	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := make([]string, 0, 32)
	pathIdx := make(map[string]int)
	var cycles [][]string
	seenCycle := make(map[string]bool)

	var dfs func(node string)
	dfs = func(node string) {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)
		pathIdx[node] = len(path) - 1

		for _, next := range adj[node] {
			if !visited[next] {
				dfs(next)
			} else if recStack[next] {
				start := pathIdx[next]
				cycle := make([]string, len(path)-start)
				copy(cycle, path[start:])
				key := cycleKey(cycle)
				if !seenCycle[key] {
					seenCycle[key] = true
					cycles = append(cycles, cycle)
				}
			}
		}

		path = path[:len(path)-1]
		delete(pathIdx, node)
		recStack[node] = false
	}

	for _, n := range graph.Nodes {
		if !visited[n.Path] {
			dfs(n.Path)
		}
	}
	// Also run from any node that has edges but might not be in Nodes
	for from := range adj {
		if !visited[from] {
			dfs(from)
		}
	}
	return cycles
}

func cycleKey(cycle []string) string {
	if len(cycle) == 0 {
		return ""
	}
	min := cycle[0]
	for _, s := range cycle[1:] {
		if s < min {
			min = s
		}
	}
	// Normalize so we don't duplicate same cycle with different start
	idx := 0
	for i, s := range cycle {
		if s == min {
			idx = i
			break
		}
	}
	var b []byte
	for i := 0; i < len(cycle); i++ {
		p := (idx + i) % len(cycle)
		if len(b) > 0 {
			b = append(b, '|')
		}
		b = append(b, cycle[p]...)
	}
	return string(b)
}
