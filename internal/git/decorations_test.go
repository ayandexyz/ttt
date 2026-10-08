package git

import (
	"reflect"
	"testing"
)

func TestParseDecorations(t *testing.T) {
	got := parseDecorations("HEAD -> refs/heads/feat/x, tag: refs/tags/v1.0, refs/remotes/origin/main, refs/remotes/origin/HEAD, refs/heads/main")
	want := []Decoration{
		{Name: "feat/x", Kind: DecorationHead},
		{Name: "v1.0", Kind: DecorationTag},
		{Name: "origin/main", Kind: DecorationRemote},
		{Name: "main", Kind: DecorationBranch},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseDecorations = %+v, want %+v", got, want)
	}
	if got := parseDecorations("HEAD"); !reflect.DeepEqual(got, []Decoration{{Name: "HEAD", Kind: DecorationHead}}) {
		t.Errorf("detached = %+v", got)
	}
}
