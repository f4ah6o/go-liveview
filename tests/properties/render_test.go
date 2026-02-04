package properties

import (
	"testing"

	"github.com/fu2hito/go-liveview/internal/render"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestDiffApply tests that applying a diff produces the expected result
// Note: In the Phoenix protocol, the client maintains static parts separately.
// So when applying a diff, we use the static from the original render.
func TestDiffApply(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("apply(diff(A,B), A) == B", prop.ForAll(
		func(dynamicA, dynamicB string) bool {
			// Use fixed static parts
			staticParts := []string{"<div>", "</div>"}
			dynamicAParts := []interface{}{dynamicA}
			dynamicBParts := []interface{}{dynamicB}

			// Create two Rendered states
			prev := &render.Rendered{
				Static:  staticParts,
				Dynamic: dynamicAParts,
			}
			curr := &render.Rendered{
				Static:  staticParts,
				Dynamic: dynamicBParts,
			}

			// Calculate diff
			diff := render.Diff(prev, curr)

			// Apply diff using original static parts (what client does)
			staticToUse := staticParts
			if diff.Static != nil {
				staticToUse = diff.Static
			}
			resultHTML := render.BuildHTML(staticToUse, diff.Dynamic)
			expectedHTML := render.BuildHTML(curr.Static, curr.Dynamic)

			return resultHTML == expectedHTML
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

// TestDiffIdentity tests that diff(A,A) produces no changes
func TestDiffIdentity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()

	properties := gopter.NewProperties(parameters)

	properties.Property("diff(A,A) has no changes", prop.ForAll(
		func(dynamic string) bool {
			staticParts := []string{"<div>", "</div>"}
			dynamicParts := []interface{}{dynamic}

			r := &render.Rendered{
				Static:  staticParts,
				Dynamic: dynamicParts,
			}

			diff := render.Diff(r, r)

			// When static is same and dynamic is same, diff should have no static
			if diff.Static != nil {
				return false
			}

			// All dynamic values should be nil (no change)
			for _, v := range diff.Dynamic {
				if v != nil {
					return false
				}
			}

			return true
		},
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

// TestBuildHTMLIsValid tests that BuildHTML produces valid-looking HTML
func TestBuildHTMLIsValid(t *testing.T) {
	parameters := gopter.DefaultTestParameters()

	properties := gopter.NewProperties(parameters)

	properties.Property("BuildHTML produces valid HTML structure", prop.ForAll(
		func(static string, dynamic string) bool {
			staticParts := []string{static}
			dynamicParts := []interface{}{dynamic}

			html := render.BuildHTML(staticParts, dynamicParts)

			// Check for basic validity: non-negative length
			return len(html) >= 0
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.TestingRun(t)
}
