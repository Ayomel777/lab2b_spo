package analyzer

type SubgraphInfo struct {
	SelectVertex  int   `json:"selectVertex"`
	Transitions   []int `json:"transitions"`
	SelectCost    int   `json:"selectCost"`
	SubgraphVerts []int `json:"subgraphVerts"`
	LowerBound    int   `json:"lowerBound"`
}

type VertexComplexity struct {
	VertexID   int    `json:"vertexId"`
	Complexity int    `json:"complexity"`
	Kind       string `json:"kind"`
}

type BoundaryResult struct {
	Sa       int                `json:"sa"`
	So       float64            `json:"so"`
	Subs     []SubgraphInfo     `json:"subs"`
	Vertices []VertexComplexity `json:"vertices"`
}

func AnalyzeBoundary(cfg *CFG) BoundaryResult {
	if cfg == nil || len(cfg.Nodes) == 0 {
		return BoundaryResult{}
	}

	res := BoundaryResult{
		Subs:     make([]SubgraphInfo, 0),
		Vertices: make([]VertexComplexity, 0, len(cfg.Nodes)),
	}

	for _, n := range cfg.Nodes {
		if n == cfg.End {
			res.Vertices = append(res.Vertices, VertexComplexity{
				VertexID: n.ID, Complexity: 0, Kind: "end",
			})
			continue
		}
		if len(n.Out) <= 1 {
			kind := "accept"
			if n == cfg.Start {
				kind = "start"
			}
			res.Vertices = append(res.Vertices, VertexComplexity{
				VertexID: n.ID, Complexity: 1, Kind: kind,
			})
			res.Sa += 1
			continue
		}

		merge, complexity := boundaryForSelect(cfg, n)
		subgraphVerts := collectIntermediate(n, merge)

		transitions := make([]int, 0, len(n.Out))
		for _, s := range n.Out {
			transitions = append(transitions, s.ID)
		}

		lowerID := -1
		if merge != nil {
			lowerID = merge.ID
		}

		res.Subs = append(res.Subs, SubgraphInfo{
			SelectVertex:  n.ID,
			Transitions:   transitions,
			SelectCost:    complexity,
			SubgraphVerts: subgraphVerts,
			LowerBound:    lowerID,
		})
		res.Vertices = append(res.Vertices, VertexComplexity{
			VertexID: n.ID, Complexity: complexity, Kind: "select",
		})
		res.Sa += complexity
	}

	if res.Sa > 0 {
		res.So = 1.0 - float64(len(cfg.Nodes)-1)/float64(res.Sa)
	}

	return res
}

func boundaryForSelect(cfg *CFG, v *CFGNode) (*CFGNode, int) {
	if len(v.Out) < 2 {
		return nil, 1
	}

	var loopSucc, exitSucc *CFGNode
	for _, s := range v.Out {
		if canReach(s, v) {
			loopSucc = s
		} else {
			exitSucc = s
		}
	}

	if loopSucc != nil && exitSucc != nil {
		return exitSucc, computeLoopComplexity(v, exitSucc)
	}

	merge := findMergePoint(v.Out)
	if merge == nil {
		merge = cfg.End
	}
	return merge, computeBranchComplexity(v, merge)
}

func computeLoopComplexity(v, exit *CFGNode) int {
	visited := map[int]bool{}
	queue := []*CFGNode{}
	for _, s := range v.Out {
		if s != exit {
			queue = append(queue, s)
		}
	}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		if visited[n.ID] || n == exit {
			continue
		}
		visited[n.ID] = true
		for _, s := range n.Out {
			if s != exit && !visited[s.ID] {
				queue = append(queue, s)
			}
		}
	}
	return len(visited) + 1
}

func computeBranchComplexity(v, merge *CFGNode) int {
	visited := map[int]bool{}
	queue := []*CFGNode{}
	for _, s := range v.Out {
		if s != merge {
			queue = append(queue, s)
		}
	}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		if visited[n.ID] || n == merge {
			continue
		}
		visited[n.ID] = true
		for _, s := range n.Out {
			if s != merge && !visited[s.ID] {
				queue = append(queue, s)
			}
		}
	}
	return len(visited) + 1
}

func collectIntermediate(v, merge *CFGNode) []int {
	if merge == nil {
		return nil
	}
	visited := map[int]bool{}
	queue := []*CFGNode{}
	for _, s := range v.Out {
		if s != merge {
			queue = append(queue, s)
		}
	}
	result := []int{}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		if visited[n.ID] || n == merge {
			continue
		}
		visited[n.ID] = true
		result = append(result, n.ID)
		for _, s := range n.Out {
			if s != merge && !visited[s.ID] {
				queue = append(queue, s)
			}
		}
	}
	return result
}
