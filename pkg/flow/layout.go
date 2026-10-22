package flow

// LayoutFlowchart arranges nodes top-to-bottom in layers (Sugiyama-style topological leveling).
func LayoutFlowchart(fc *Flowchart) {
	if fc == nil || len(fc.Nodes) == 0 {
		return
	}

	// 1. Calculate in-degrees and build adjacency lists
	inDegree := make(map[string]int)
	adj := make(map[string][]string)

	for _, n := range fc.Nodes {
		inDegree[n.ID] = 0
		adj[n.ID] = make([]string, 0)
	}

	for _, e := range fc.Edges {
		if !e.IsBackArc {
			adj[e.FromID] = append(adj[e.FromID], e.ToID)
			inDegree[e.ToID]++
		}
	}

	// 2. Assign layers using BFS/Topological traversal
	layerMap := make(map[string]int)
	queue := make([]string, 0)

	for _, n := range fc.Nodes {
		if inDegree[n.ID] == 0 {
			queue = append(queue, n.ID)
			layerMap[n.ID] = 0
		}
	}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		currLayer := layerMap[curr]

		for _, neighbor := range adj[curr] {
			if layerMap[neighbor] < currLayer+1 {
				layerMap[neighbor] = currLayer + 1
			}
			inDegree[neighbor]--
			if inDegree[neighbor] <= 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Group nodes by layer
	layerGroups := make(map[int][]*FlowNode)
	maxLayer := 0
	for _, n := range fc.Nodes {
		lvl := layerMap[n.ID]
		n.Layer = lvl
		layerGroups[lvl] = append(layerGroups[lvl], n)
		if lvl > maxLayer {
			maxLayer = lvl
		}
	}

	// 3. Coordinate placement: calculate X and Y
	verticalSpacing := 85
	horizontalSpacing := 40
	currentY := 60
	maxX := 400

	centerX := 450 // Default center axis

	for l := 0; l <= maxLayer; l++ {
		nodesInLayer := layerGroups[l]
		count := len(nodesInLayer)
		if count == 0 {
			continue
		}

		totalWidth := 0
		for _, n := range nodesInLayer {
			totalWidth += n.Width
		}
		totalWidth += (count - 1) * horizontalSpacing

		startX := centerX - totalWidth/2
		if startX < 50 {
			startX = 50
		}

		layerMaxHeight := 0
		currX := startX
		for _, n := range nodesInLayer {
			n.X = currX
			n.Y = currentY
			currX += n.Width + horizontalSpacing
			if n.X+n.Width > maxX {
				maxX = n.X + n.Width
			}
			if n.Height > layerMaxHeight {
				layerMaxHeight = n.Height
			}
		}

		currentY += layerMaxHeight + verticalSpacing
	}

	fc.Width = maxX + 80
	if fc.Width < 800 {
		fc.Width = 800
	}
	fc.Height = currentY + 40
}
