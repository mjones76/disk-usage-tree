// Package dutree walks a directory tree and reports how much space each
// file and subdirectory occupies, similar to `du -s` but returning a tree a
// caller can render, filter, or serialize however they like.
package dutree

import (
	"os"
	"path/filepath"
	"sort"
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
func Scan(root string, opts Options) (*Entry, error) {
	info, err := os.Lstat(root)
	if err != nil {
		return nil, err
	}
	e, _ := scan(root, info, opts, 0)
	return e, nil
}

func scan(path string, info os.FileInfo, opts Options, depth int) (*Entry, error) {
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

	for _, de := range entries {
		childInfo, err := de.Info()
		if err != nil {
			continue
		}
		child, err := scan(filepath.Join(path, de.Name()), childInfo, opts, depth+1)
		if err != nil || child == nil {
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
