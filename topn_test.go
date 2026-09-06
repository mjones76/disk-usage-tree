package dutree

import (
	"path/filepath"
	"testing"
)

func TestLargestOrdersBySizeDescending(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "small.txt"), 10)
	writeFile(t, filepath.Join(root, "big.txt"), 300)
	writeFile(t, filepath.Join(root, "sub", "medium.txt"), 100)

	e, err := Scan(root, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	got := Largest(e, 2)
	if len(got) != 2 {
		t.Fatalf("len(Largest) = %d, want 2", len(got))
	}
	if got[0].Name != "big.txt" || got[0].Size != 300 {
		t.Fatalf("got[0] = %q (%d bytes), want big.txt (300)", got[0].Name, got[0].Size)
	}
	// "sub" (100 bytes, a directory) outranks "small.txt" (10 bytes).
	if got[1].Name != "sub" || got[1].Size != 100 {
		t.Fatalf("got[1] = %q (%d bytes), want sub (100)", got[1].Name, got[1].Size)
	}
}

func TestLargestExcludesRoot(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.txt"), 10)

	e, err := Scan(root, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	for _, entry := range Largest(e, 10) {
		if entry == e {
			t.Fatalf("Largest included the root entry")
		}
	}
}

func TestLargestCapsAtAvailableEntries(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.txt"), 10)
	writeFile(t, filepath.Join(root, "b.txt"), 20)

	e, err := Scan(root, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	got := Largest(e, 100)
	if len(got) != 2 {
		t.Fatalf("len(Largest) = %d, want 2", len(got))
	}
}

func TestLargestZeroOrNegativeN(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.txt"), 10)

	e, err := Scan(root, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if got := Largest(e, 0); got != nil {
		t.Fatalf("Largest(e, 0) = %v, want nil", got)
	}
	if got := Largest(e, -1); got != nil {
		t.Fatalf("Largest(e, -1) = %v, want nil", got)
	}
}

func TestLargestNilRoot(t *testing.T) {
	if got := Largest(nil, 5); got != nil {
		t.Fatalf("Largest(nil, 5) = %v, want nil", got)
	}
}
