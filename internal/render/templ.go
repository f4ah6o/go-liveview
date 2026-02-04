package render

import (
	"regexp"
)

var dynamicMarkerRegex = regexp.MustCompile(`<!--\$\d+-->(.*?)<!--/\$\d+-->`)

// ParseTemplOutput parses HTML output from templ and extracts static/dynamic parts
// This is a simple regex-based parser for the initial implementation
func ParseTemplOutput(html string) *Rendered {
	// Find all dynamic markers
	matches := dynamicMarkerRegex.FindAllStringSubmatchIndex(html, -1)

	if len(matches) == 0 {
		// No dynamic content
		return &Rendered{
			Static:  []string{html},
			Dynamic: []interface{}{},
		}
	}

	static := []string{}
	dynamic := []interface{}{}

	lastEnd := 0
	for _, match := range matches {
		// Add static part before this dynamic marker
		static = append(static, html[lastEnd:match[0]])

		// Extract dynamic content (inside the markers)
		dynamicContent := html[match[2]:match[3]]
		dynamic = append(dynamic, dynamicContent)

		lastEnd = match[1]
	}

	// Add remaining static part
	if lastEnd < len(html) {
		static = append(static, html[lastEnd:])
	}

	return &Rendered{
		Static:  static,
		Dynamic: dynamic,
	}
}

// ParseTemplOutputWithNesting parses nested components
func ParseTemplOutputWithNesting(html string) *Rendered {
	// First, handle nested components by replacing them with placeholders
	html = handleNestedComponents(html)

	// Then parse as usual
	return ParseTemplOutput(html)
}

func handleNestedComponents(html string) string {
	// Look for component markers and replace them with placeholders
	// This is a simplified version - full implementation would handle arbitrary nesting
	return html
}

// MergeRendered merges multiple Rendered components
func MergeRendered(components ...*Rendered) *Rendered {
	if len(components) == 0 {
		return &Rendered{Static: []string{}, Dynamic: []interface{}{}}
	}

	if len(components) == 1 {
		return components[0]
	}

	result := &Rendered{
		Static:  []string{},
		Dynamic: []interface{}{},
	}

	for i, comp := range components {
		if i > 0 && len(result.Static) > 0 && len(comp.Static) > 0 {
			// Merge last static of previous with first static of current
			result.Static[len(result.Static)-1] += comp.Static[0]
			result.Static = append(result.Static, comp.Static[1:]...)
		} else {
			result.Static = append(result.Static, comp.Static...)
		}
		result.Dynamic = append(result.Dynamic, comp.Dynamic...)
	}

	return result
}
