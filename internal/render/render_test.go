package render_test

import (
	"testing"

	"github.com/fu2hito/go-liveview/internal/render"
)

func TestDiffSimple(t *testing.T) {
	staticParts := []string{"<div>", "</div>"}

	prev := &render.Rendered{
		Static:  staticParts,
		Dynamic: []interface{}{"a"},
	}

	curr := &render.Rendered{
		Static:  staticParts,
		Dynamic: []interface{}{"b"},
	}

	diff := render.Diff(prev, curr)
	t.Logf("Diff: %+v", diff)

	// When static parts are the same, diff.Static should be nil
	// but we need to use the original static parts for rendering
	staticToUse := staticParts
	if diff.Static != nil {
		staticToUse = diff.Static
	}

	result := render.BuildHTML(staticToUse, diff.Dynamic)
	expected := "<div>b</div>"

	if result != expected {
		t.Errorf("BuildHTML: got %q, want %q", result, expected)
	}
}
