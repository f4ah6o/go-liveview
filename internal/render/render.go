package render

// Rendered represents a rendered LiveView template with static and dynamic parts
type Rendered struct {
	Static      []string      `json:"s"`
	Dynamic     []interface{} `json:"d"`
	Fingerprint string        `json:"fingerprint,omitempty"`
}

// IsEqual checks if two Rendered structs are equal
func (r *Rendered) IsEqual(other *Rendered) bool {
	if len(r.Static) != len(other.Static) {
		return false
	}
	for i := range r.Static {
		if r.Static[i] != other.Static[i] {
			return false
		}
	}
	if len(r.Dynamic) != len(other.Dynamic) {
		return false
	}
	// Deep comparison of dynamic parts
	return deepEqualDynamic(r.Dynamic, other.Dynamic)
}

func deepEqualDynamic(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		switch av := a[i].(type) {
		case string:
			bv, ok := b[i].(string)
			if !ok || av != bv {
				return false
			}
		case *Rendered:
			bv, ok := b[i].(*Rendered)
			if !ok || !av.IsEqual(bv) {
				return false
			}
		case []interface{}:
			bv, ok := b[i].([]interface{})
			if !ok || !deepEqualDynamic(av, bv) {
				return false
			}
		default:
			if a[i] != b[i] {
				return false
			}
		}
	}
	return true
}

// Patch represents a diff between two Rendered states
type Patch struct {
	Static  []string      `json:"s,omitempty"`
	Dynamic []interface{} `json:"d"`
	Append  bool          `json:"a,omitempty"`
	Prepend bool          `json:"p,omitempty"`
}

// Diff calculates the difference between two Rendered states
func Diff(prev, curr *Rendered) *Patch {
	// If static parts differ or prev is nil, send full current state
	if prev == nil || !staticEqual(prev.Static, curr.Static) {
		return &Patch{
			Static:  curr.Static,
			Dynamic: curr.Dynamic,
		}
	}

	// Static parts are equal, calculate dynamic diff
	diff := diffDynamic(prev.Dynamic, curr.Dynamic)
	return &Patch{
		Dynamic: diff,
	}
}

func staticEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func diffDynamic(prev, curr []interface{}) []interface{} {
	result := make([]interface{}, len(curr))
	for i := range curr {
		if i < len(prev) {
			// Compare previous and current values
			switch cv := curr[i].(type) {
			case string:
				pv, ok := prev[i].(string)
				if ok && pv == cv {
					// No change
					result[i] = nil
				} else {
					result[i] = cv
				}
			case *Rendered:
				pv, ok := prev[i].(*Rendered)
				if ok {
					childDiff := Diff(pv, cv)
					if childDiff.Static == nil {
						// Only dynamic parts changed
						result[i] = childDiff.Dynamic
					} else {
						// Structure changed
						result[i] = cv
					}
				} else {
					result[i] = cv
				}
			case []interface{}:
				pv, ok := prev[i].([]interface{})
				if ok {
					childDiff := diffDynamic(pv, cv)
					// Check if any changes
					hasChange := false
					for _, v := range childDiff {
						if v != nil {
							hasChange = true
							break
						}
					}
					if hasChange {
						result[i] = childDiff
					} else {
						result[i] = nil
					}
				} else {
					result[i] = cv
				}
			default:
				if prev[i] == curr[i] {
					result[i] = nil
				} else {
					result[i] = cv
				}
			}
		} else {
			// New element
			result[i] = curr[i]
		}
	}
	return result
}

// BuildHTML constructs HTML from static and dynamic parts
func BuildHTML(static []string, dynamic []interface{}) string {
	if len(static) == 0 {
		return ""
	}

	result := ""
	for i := 0; i < len(static); i++ {
		result += static[i]
		if i < len(dynamic) && dynamic[i] != nil {
			result += renderDynamic(dynamic[i])
		}
	}
	return result
}

func renderDynamic(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case *Rendered:
		return BuildHTML(v.Static, v.Dynamic)
	case []interface{}:
		result := ""
		for _, item := range v {
			result += renderDynamic(item)
		}
		return result
	case nil:
		return ""
	default:
		return ""
	}
}
