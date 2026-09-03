// Package dutree walks a directory tree and reports how much space each
// file and subdirectory occupies, similar to `du -s` but returning a tree a
// caller can render, filter, or serialize however they like.
package dutree

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
)

// Entry is one node in a scanned directory tree. For a file, Size is its
// length in bytes. For a directory, Size is the sum of everything under it.
type Entry struct {
	Path     string   `json:"path"`
	Name     string   `json:"name"`
	IsDir    bool     `json:"is_dir"`
	Size     int64    `json:"size"`
	Children []*Entry `json:"children,omitempty"`
}

// Options controls how Scan walks the filesystem.
type Options struct {
	// SkipHidden excludes files and directories whose name starts with a
	// dot from the returned tree (their size is still not counted).
	SkipHidden bool
	// MaxDepth limits how many levels of children are kept in the returned
	// tree. Sizes are still computed from the full tree underneath. Zero
	// means no limit.
	MaxDepth int
}

// Scan walks the tree rooted at root and returns its aggregated sizes.
// Entries the caller can't read (permission errors, or a file removed
// mid-walk) are skipped rather than aborting the whole scan, since one
// locked-down directory shouldn't stop the rest of the report.
//
// Sibling subdirectories are scanned concurrently, bounded by GOMAXPROCS,
// so a wide tree on fast storage doesn't sit idle waiting on one directory
// at a time.
func Scan(root string, opts Options) (*Entry, error) {
	info, err := os.Lstat(root)
	if err != nil {
		return nil, err
	}
	sem := make(chan struct{}, workerLimit())
	e, _ := scan(root, info, opts, 0, sem)
	return e, nil
}

func workerLimit() int {
	if n := runtime.GOMAXPROCS(0); n > 1 {
		return n
	}
	return 1
}

func scan(path string, info os.FileInfo, opts Options, depth int, sem chan struct{}) (*Entry, error) {
	name := info.Name()
	if opts.SkipHidden && depth > 0 && name[0] == '.' {
		return nil, nil
	}

	e := &Entry{
		Path:  path,
		Name:  name,
		IsDir: info.IsDir(),
	}

	if !info.IsDir() {
		e.Size = info.Size()
		return e, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		// Exists but unreadable (permissions, etc). Report it as an empty
		// leaf instead of failing the whole scan.
		return e, nil
	}

	children := make([]*Entry, len(entries))
	var wg sync.WaitGroup
	for i, de := range entries {
		childInfo, err := de.Info()
		if err != nil {
			continue
		}
		childPath := filepath.Join(path, de.Name())

		// Hand the child to a worker if one is free; otherwise scan it
		// inline. The non-blocking select keeps this from deadlocking
		// against itself as goroutines recurse and compete for the same
		// bounded pool of slots.
		select {
		case sem <- struct{}{}:
			wg.Add(1)
			go func(i int, path string, info os.FileInfo) {
				defer wg.Done()
				defer func() { <-sem }()
				child, _ := scan(path, info, opts, depth+1, sem)
				children[i] = child
			}(i, childPath, childInfo)
		default:
			child, _ := scan(childPath, childInfo, opts, depth+1, sem)
			children[i] = child
		}
	}
	wg.Wait()

	for _, child := range children {
		if child == nil {
			continue
		}
		e.Size += child.Size
		if opts.MaxDepth == 0 || depth < opts.MaxDepth {
			e.Children = append(e.Children, child)
		}
	}

	sort.Slice(e.Children, func(i, j int) bool {
		return e.Children[i].Size > e.Children[j].Size
	})

	return e, nil
}

// Flatten returns every entry in the tree as a single slice, parents before
// their children, in the same order Scan discovered them.
func Flatten(e *Entry) []*Entry {
	if e == nil {
		return nil
	}
	out := []*Entry{e}
	for _, c := range e.Children {
		out = append(out, Flatten(c)...)
	}
	return out
}
