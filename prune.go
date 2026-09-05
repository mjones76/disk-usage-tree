package dutree

import "path/filepath"

// Prune returns a copy of the tree with any entry whose name matches one of
// the given patterns removed, along with everything under it. Sizes on the
// surviving ancestors are reduced so they no longer count the pruned
// subtrees. Patterns use filepath.Match syntax and are checked against each
// entry's base name rather than its full path, so a pattern like "*.log"
// matches a matching file at any depth, not just at the root.
//
// The root itself can match and be pruned away, in which case Prune returns
// nil. A malformed pattern never matches rather than aborting the prune,
// since this is meant for quick filtering, not pattern validation.
func Prune(e *Entry, patterns ...string) *Entry {
	if e == nil || len(patterns) == 0 {
		return e
	}
	return prune(e, patterns)
}

func prune(e *Entry, patterns []string) *Entry {
	if matchesAny(e.Name, patterns) {
		return nil
	}
	if len(e.Children) == 0 {
		return e
	}

	clone := *e
	clone.Children = nil
	for _, c := range e.Children {
		kept := prune(c, patterns)
		if kept == nil {
			clone.Size -= c.Size
			continue
		}
		clone.Children = append(clone.Children, kept)
	}
	return &clone
}

func matchesAny(name string, patterns []string) bool {
	for _, p := range patterns {
		if ok, err := filepath.Match(p, name); err == nil && ok {
			return true
		}
	}
	return false
}
