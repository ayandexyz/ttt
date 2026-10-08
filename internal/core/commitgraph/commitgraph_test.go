package commitgraph

import (
	"strings"
	"testing"
)

type commit struct {
	hash    string
	parents []string
}

func render(commits []commit) []string {
	var g Graph
	var rows []string
	for _, c := range commits {
		var b strings.Builder
		for _, cell := range g.Next(c.hash, c.parents) {
			if cell.Commit {
				b.WriteRune('●')
			} else {
				b.WriteRune(cell.Ch)
			}
		}
		rows = append(rows, strings.TrimRight(b.String(), " "))
	}
	return rows
}

func assertRows(t *testing.T, got, want []string) {
	t.Helper()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("graph:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestLinearHistoryIsOneLane(t *testing.T) {
	assertRows(t, render([]commit{
		{"c", []string{"b"}},
		{"b", []string{"a"}},
		{"a", nil},
	}), []string{"●", "●", "●"})
}

func TestMergeOpensAndClosesALane(t *testing.T) {
	assertRows(t, render([]commit{
		{"m", []string{"a2", "b1"}},
		{"b1", []string{"a1"}},
		{"a2", []string{"a1"}},
		{"a1", nil},
	}), []string{
		"●─╮",
		"│ ●",
		"● │",
		"●─╯",
	})
}

func TestSeveralLanesConvergeOnOneCommit(t *testing.T) {
	assertRows(t, render([]commit{
		{"m", []string{"a", "x"}},
		{"n", []string{"a", "y"}},
		{"x", []string{"a"}},
		{"y", []string{"a"}},
		{"a", nil},
	}), []string{
		"●─╮",
		"│ │ ●─╮",
		"│ ● │ │",
		"│ │ │ ●",
		"●─┴─┴─╯",
	})
}

func TestConnectorCrossingALaneDrawsAJunction(t *testing.T) {
	assertRows(t, render([]commit{
		{"m", []string{"a", "x"}},
		{"p", []string{"a"}},
		{"a", nil},
		{"x", nil},
	}), []string{
		"●─╮",
		"│ │ ●",
		"●─┼─╯",
		"  ●",
	})
}

func TestCommitsContinueAcrossPages(t *testing.T) {
	var g Graph
	g.Next("m", []string{"a", "b"})
	cells := g.Next("b", []string{"a"})
	if len(cells) != 3 || cells[0].Ch != '│' || !cells[2].Commit || cells[2].Lane != 1 {
		t.Fatalf("second row = %+v, want the branch commit on lane 1", cells)
	}
}
