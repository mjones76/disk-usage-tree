# disk-usage-tree

A Go library for measuring how space is used under a directory. It walks a
tree once and gives back the aggregated size of every file and folder in
it, sorted largest first, so a caller can print a report, filter it, or
serve it as JSON.

This is deliberately not a CLI. It's the part of `du` that does the work,
without opinions about flags or output format baked into a binary. Wire it
into whatever front end you need.

## Why

Every time I've wanted "what's eating my disk" logic in a Go tool, I've
ended up rewriting the same recursive walk-and-sum. This is that logic,
written once, with a tree structure a caller can actually do something
with instead of a flat list of paths.

## Usage

```go
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mjones76/disk-usage-tree"
)

func main() {
	jsonOut := flag.Bool("json", false, "emit JSON instead of a text report")
	flag.Parse()

	root := "."
	if flag.NArg() > 0 {
		root = flag.Arg(0)
	}

	tree, err := dutree.Scan(root, dutree.Options{SkipHidden: true})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *jsonOut {
		dutree.WriteJSON(os.Stdout, tree)
	} else {
		dutree.WriteText(os.Stdout, tree)
	}
}
```

Text output:

```
1.2GiB   myproject
  800.0MiB node_modules
  300.0MiB .git
  100.0MiB src
```

JSON output (`--json`):

```json
{
  "path": "myproject",
  "name": "myproject",
  "is_dir": true,
  "size": 1288490188,
  "children": [
    { "path": "myproject/node_modules", "name": "node_modules", "is_dir": true, "size": 838860800 }
  ]
}
```

The two outputs describe the same tree, so anything the text report can
show, a script consuming the JSON can compute too.

## API

- `Scan(root string, opts Options) (*Entry, error)` — walk `root` and
  return the sized tree.
- `Options.SkipHidden` — leave out dotfiles and dotdirs.
- `Options.MaxDepth` — cap how many levels of children come back (sizes
  are still totaled from the full tree underneath).
- `Flatten(e *Entry) []*Entry` — the tree as a single slice, for callers
  that want to sort or filter across the whole scan.
- `Prune(e *Entry, patterns ...string) *Entry` — drop entries whose name
  matches a `filepath.Match` pattern (e.g. `"node_modules"`, `"*.log"`),
  along with everything under them, and adjust ancestor sizes to match.
  Returns a new tree; the one passed in is left untouched.
- `WriteText` / `WriteJSON` — render a tree in either format.
- `HumanSize(bytes int64) string` — the `du -h` style formatter used by
  `WriteText`, exported in case you want it for your own output.

Unreadable files and directories are skipped rather than aborting the
scan; a single locked-down folder shouldn't stop the rest of the report.

## Status

Early. The core scan, both output formats, and pruning by pattern work and
are tested. See the roadmap for what's still missing.

## Roadmap

- top-N largest entries helper
- `context.Context` support for cancelling long scans
- benchmarks for scan performance on large trees
