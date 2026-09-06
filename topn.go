package dutree

import "sort"

// Largest returns the n entries with the biggest Size in the tree rooted at
// e, largest first. The root itself is excluded: it always contains
// everything else, so ranking it alongside its own contents wouldn't tell a
// caller anything. Both files and directories are eligible, so a huge
// node_modules directory can outrank any individual file.
//
// If the tree has fewer than n entries under the root, Largest returns all
// of them. A nil root or n <= 0 returns nil.
func Largest(e *Entry, n int) []*Entry {
	if e == nil || n <= 0 {
		return nil
	}

	var candidates []*Entry
	for _, c := range e.Children {
		candidates = append(candidates, Flatten(c)...)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Size > candidates[j].Size
	})

	if n > len(candidates) {
		n = len(candidates)
	}
	return candidates[:n]
}
