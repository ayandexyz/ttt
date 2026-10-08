package commitgraph

// Cell is one terminal column of a graph row. Lane picks the color; it is -1
// for blank cells.
type Cell struct {
	Ch     rune
	Lane   int
	Commit bool
}

// Graph lays out commits one row at a time, newest first, so a paged history
// can keep extending the same graph.
type Graph struct {
	lanes []string
}

func (g *Graph) Reset() {
	g.lanes = nil
}

// Next returns the row for a commit. Each lane occupies two columns: the lane
// glyph and the gap after it, which carries horizontal connectors.
func (g *Graph) Next(hash string, parents []string) []Cell {
	before := append([]string(nil), g.lanes...)

	commit := -1
	for i, h := range g.lanes {
		if h == hash {
			commit = i
			break
		}
	}
	if commit < 0 {
		commit = g.alloc()
	}

	kind := make(map[int]rune)
	for i, h := range before {
		if i == commit || h != hash {
			continue
		}
		if i > commit {
			kind[i] = '╯'
		} else {
			kind[i] = '╰'
		}
	}

	if len(parents) > 0 {
		g.lanes[commit] = parents[0]
	} else {
		g.lanes[commit] = ""
	}
	// New lanes are allocated while converging lanes still hold hash, so a
	// slot never has to draw a lane ending and another starting at once.
	for _, p := range parents[min(1, len(parents)):] {
		k := g.alloc()
		g.lanes[k] = p
		if k > commit {
			kind[k] = '╮'
		} else {
			kind[k] = '╭'
		}
	}
	for i := range kind {
		if g.lanes[i] == hash {
			g.lanes[i] = ""
		}
	}

	width := max(len(before), len(g.lanes))
	g.trim()

	lo, hi := commit, commit
	for i := range kind {
		lo, hi = min(lo, i), max(hi, i)
	}
	for i, ch := range kind {
		if i == lo || i == hi {
			continue
		}
		if ch == '╯' || ch == '╰' {
			kind[i] = '┴'
		} else {
			kind[i] = '┬'
		}
	}

	cells := make([]Cell, 0, width*2-1)
	for i := range width {
		cell := Cell{Ch: ' ', Lane: -1}
		passing := i < len(before) && before[i] != "" && i < len(g.lanes) && g.lanes[i] != ""
		switch ch, connector := kind[i]; {
		case i == commit:
			cell = Cell{Lane: i, Commit: true}
		case connector:
			cell = Cell{Ch: ch, Lane: i}
		case passing && i > lo && i < hi:
			cell = Cell{Ch: '┼', Lane: i}
		case passing:
			cell = Cell{Ch: '│', Lane: i}
		}
		cells = append(cells, cell)
		if i == width-1 {
			break
		}
		gap := Cell{Ch: ' ', Lane: -1}
		if i >= lo && i < hi {
			gap = Cell{Ch: '─', Lane: nearestConnector(kind, commit, i)}
		}
		cells = append(cells, gap)
	}
	return cells
}

// nearestConnector colors the gap after lane i with the connector it leads to,
// moving away from the commit.
func nearestConnector(kind map[int]rune, commit, i int) int {
	if i >= commit {
		for j := i + 1; ; j++ {
			if _, ok := kind[j]; ok {
				return j
			}
		}
	}
	for j := i; ; j-- {
		if _, ok := kind[j]; ok {
			return j
		}
	}
}

func (g *Graph) alloc() int {
	for i, h := range g.lanes {
		if h == "" {
			return i
		}
	}
	g.lanes = append(g.lanes, "")
	return len(g.lanes) - 1
}

func (g *Graph) trim() {
	for len(g.lanes) > 0 && g.lanes[len(g.lanes)-1] == "" {
		g.lanes = g.lanes[:len(g.lanes)-1]
	}
}
