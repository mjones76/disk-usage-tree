package dutree

import (
	"path/filepath"
	"testing"
)

func TestPruneRemovesMatchingSubtreeAndAdjustsSizes(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.txt"), 100)
	writeFile(t, filepath.Join(root, "node_modules", "b.js"), 250)

	e, err := Scan(root, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if e.Size != 350 {
		t.Fatalf("root size before prune = %d, want 350", e.Size)
	}

	pruned := Prune(e, "node_modules")
	if pruned.Size != 100 {
		t.Fatalf("root size after prune = %d, want 100", pruned.Size)
	}
	if len(pruned.Children) != 1 {
		t.Fatalf("children after prune = %d, want 1", len(pruned.Children))
	}
	if pruned.Children[0].Name != "a.txt" {
		t.Fatalf("surviving child = %q, want %q", pruned.Children[0].Name, "a.txt")
	}
}

func TestPruneMatchesGlobAtAnyDepth(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "keep.txt"), 10)
	writeFile(t, filepath.Join(root, "debug.log"), 20)
	writeFile(t, filepath.Join(root, "sub", "trace.log"), 30)

	e, err := Scan(root, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	pruned := Prune(e, "*.log")
	if pruned.Size != 10 {
		t.Fatalf("root size after prune = %d, want 10", pruned.Size)
	}
	for _, entry := range Flatten(pruned) {
		if filepath.Ext(entry.Name) == ".log" {
			t.Fatalf("found unpruned entry %q", entry.Name)
		}
	}
}

func TestPruneRootMatch(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.txt"), 10)

	e, err := Scan(root, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if pruned := Prune(e, filepath.Base(root)); pruned != nil {
		t.Fatalf("Prune(root) = %v, want nil", pruned)
	}
}

func TestPruneNoPatternsReturnsSameTree(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.txt"), 10)

	e, err := Scan(root, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if pruned := Prune(e); pruned != e {
		t.Fatalf("Prune with no patterns returned a different tree")
	}
}

func TestPruneDoesNotMutateOriginalTree(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.txt"), 100)
	writeFile(t, filepath.Join(root, "node_modules", "b.js"), 250)

	e, err := Scan(root, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	Prune(e, "node_modules")
	if e.Size != 350 {
		t.Fatalf("original tree size mutated: got %d, want 350", e.Size)
	}
	if len(e.Children) != 2 {
		t.Fatalf("original tree children mutated: got %d, want 2", len(e.Children))
	}
}
